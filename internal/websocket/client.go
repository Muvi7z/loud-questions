package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
	"loud-question/internal/model"
	"net/http"
	"strconv"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2048
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrPlayerExistLobby = errors.New("player already exist")
)

//Нужно сделать реконект, при обновлении страницы, возвращать лобби по id пользователя, закрывать предыдущий коннект

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Hub          *Hub
	conn         *websocket.Conn
	Send         chan []byte
	User         model.User
	lobbyService LobbyService
	userGetter   UserGetter
	logger       *slog.Logger
}

type Message struct {
	Type   string          `json:"type"`
	SendBy string          `json:"sendBy"`
	Data   json.RawMessage `json:"data"`
}

type UserGetter interface {
	GetUser(ctx context.Context, id string) (model.User, error)
}

func NewClient(hub *Hub, conn *websocket.Conn, lobbyService LobbyService, userGetter UserGetter, logger *slog.Logger) *Client {
	return &Client{
		Hub:          hub,
		conn:         conn,
		lobbyService: lobbyService,
		userGetter:   userGetter,
		logger:       logger,
		Send:         make(chan []byte, 256),
	}
}

func (c *Client) JoinLobby(data JoinLobbyDto) (model.Lobby, error) {
	c.logger.Info("attending to join lobby", data.LobbyId)

	ctx := context.Background()

	u, err := c.userGetter.GetUser(ctx, data.UserId)
	if err != nil {
		return model.Lobby{}, err
	}

	hub, err := c.lobbyService.JoinLobby(ctx, data.LobbyId, data.UserId)
	if err != nil {
		return model.Lobby{}, err
	}

	lobbyDto := model.Lobby{
		Id:             hub.Id,
		Owner:          hub.Lobby.Owner,
		Players:        hub.Lobby.Players,
		Rounds:         hub.Lobby.Rounds,
		CurrentRound:   hub.Lobby.CurrentRound,
		CurrentSession: hub.Lobby.CurrentSession,
		Settings:       hub.Lobby.Settings,
	}

	c.Hub = hub
	c.User = u
	c.Hub.Register <- c
	return lobbyDto, nil
}

// ReadPump Получиение сообщений от клиента и отправка в хаб
func (c *Client) ReadPump() {
	defer func() {
		fmt.Println("close connection", c.User.Username)
		if c.Hub != nil {
			c.Hub.Unregister <- c
		}
		err := c.conn.Close()
		if err != nil {
			return
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	//c.conn.SetReadDeadline(time.Now().Add(pongWait))

	c.conn.SetPongHandler(func(appData string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		return nil
	})
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//log
			}
			break
		}

		var msgDto Message

		err = json.Unmarshal(msg, &msgDto)
		if err != nil {
			c.logger.Error(err.Error())
			return
		}

		switch msgDto.Type {
		case changeSettings:

			c.logger.Info("attending to change settings lobby")

			c.Hub.Broadcast <- msgDto
		case createLobby:
			// Добавить удаление из предыдущего лобби
			c.logger.Info("attending to create lobby")
			ctx := context.Background()
			fmt.Println(msgDto.Data)

			var data CreateLobbyDto

			err := json.Unmarshal(msgDto.Data, &data)
			if err != nil {
				c.logger.Error("Ошибка преобразования")
				msg, _ := json.Marshal(ErrorMessage{
					Message: "invalid credentials",
					Code:    "400",
				})
				c.Send <- msg
			}

			u, err := c.userGetter.GetUser(ctx, data.UserId)
			if err != nil {
				c.logger.Error(err.Error())
				msg, _ := json.Marshal(ErrorMessage{
					Message: err.Error(),
					Code:    "404",
				})
				c.Send <- msg
				break
			}

			hub, err := c.lobbyService.CreateLobby(ctx, data.UserId)
			if err != nil {
				c.logger.Error(err.Error())
				msg, _ := json.Marshal(ErrorMessage{
					Message: err.Error(),
					Code:    "404",
				})
				c.Send <- msg
				break
			}

			go hub.Run()
			c.Hub = hub
			c.User = u

			c.Hub.Register <- c

			lobbyDto := model.Lobby{
				Id:       hub.Id,
				Owner:    hub.Lobby.Owner,
				Players:  hub.Lobby.Players,
				Rounds:   hub.Lobby.Rounds,
				Settings: hub.Lobby.Settings,
			}

			lobbyByte, _ := json.Marshal(&lobbyDto)
			msgCreate := Message{
				Type:   msgDto.Type,
				SendBy: data.UserId,
				Data:   lobbyByte,
			}

			msgCreateByte, _ := json.Marshal(&msgCreate)

			c.Send <- msgCreateByte

		case joinLobby:
			var data JoinLobbyDto
			err := json.Unmarshal(msgDto.Data, &data)
			if err != nil {
				c.logger.Error(err.Error())
				msg, _ := json.Marshal(ErrorMessage{
					Message: "invalid credentials",
					Code:    "400",
				})
				c.Send <- msg
				break
			}

			lobbyDto, err := c.JoinLobby(data)
			if err != nil {
				c.logger.Error(err.Error())
				var msg []byte
				if errors.Is(err, ErrPlayerExistLobby) {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusBadRequest),
					})
				} else if errors.Is(err, ErrUserNotFound) {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusNotFound),
					})
				} else {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusBadRequest),
					})
				}

				c.Send <- msg
				break
			}

			lobbyByte, _ := json.Marshal(&lobbyDto)

			msgRes := Message{
				Type:   joinLobby,
				SendBy: data.UserId,
				Data:   lobbyByte,
			}

			c.Hub.Broadcast <- msgRes
		case startGame:
			c.logger.Info("starting game")

			c.Hub.Broadcast <- msgDto
		case joinGame:
			var data JoinLobbyDto
			err := json.Unmarshal(msgDto.Data, &data)
			if err != nil {
				c.logger.Error(err.Error())
				msg, _ := json.Marshal(ErrorMessage{
					Message: "invalid credentials",
					Code:    "400",
				})
				c.Send <- msg
				break
			}

			_, err = c.JoinLobby(data)
			if err != nil {
				c.logger.Error(err.Error())
				var msg []byte
				if errors.Is(err, ErrPlayerExistLobby) {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusBadRequest),
					})
				} else if errors.Is(err, ErrUserNotFound) {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusNotFound),
					})
				} else {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusBadRequest),
					})
				}

				c.Send <- msg
				break
			}

			//lobbyByte, _ := json.Marshal(&lobbyDto)
			//
			//msgRes := Message{
			//	Type:   joinLobby,
			//	SendBy: data.UserId,
			//	Data:   lobbyByte,
			//}

			c.Hub.Broadcast <- msgDto

		case leftLobby:
			var data JoinLobbyDto
			err := json.Unmarshal(msgDto.Data, &data)
			if err != nil {
				msg, _ := json.Marshal(ErrorMessage{
					Message: "invalid credentials",
					Code:    "400",
				})
				c.Send <- msg
				break
			}

			c.logger.Info("attending to join lobby", data.LobbyId)

			ctx := context.Background()

			hub, err := c.lobbyService.JoinLobby(ctx, data.LobbyId, data.UserId)
			if err != nil {
				c.logger.Error(err.Error())
				var msg []byte
				if errors.Is(err, ErrPlayerExistLobby) {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusBadRequest),
					})
				} else if errors.Is(err, ErrUserNotFound) {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusNotFound),
					})
				} else {
					msg, _ = json.Marshal(ErrorMessage{
						Message: err.Error(),
						Code:    strconv.Itoa(http.StatusBadRequest),
					})
				}

				c.Send <- msg
				break
			}

			lobbyDto := model.Lobby{
				Id:       hub.Id,
				Owner:    hub.Lobby.Owner,
				Players:  hub.Lobby.Players,
				Rounds:   hub.Lobby.Rounds,
				Settings: hub.Lobby.Settings,
			}

			lobbyByte, _ := json.Marshal(&lobbyDto)

			msgRes := Message{
				Type:   leftLobby,
				SendBy: data.UserId,
				Data:   lobbyByte,
			}

			c.Hub.Broadcast <- msgRes
		default:
			if c.Hub != nil {
				msgDto.SendBy = c.User.Uuid
				c.Hub.Broadcast <- msgDto
			}

		}

	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		if err != nil {
			return
		}
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				err := c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			//fmt.Println(len(c.Send))

			//dataBytes, err := json.Marshal(msg)
			//if err != nil {
			//	return
			//}
			w.Write(msg)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}

	}
}
