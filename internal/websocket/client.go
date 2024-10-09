package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log/slog"
	"loud-question/internal/model"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 2048
)

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
	logger       *slog.Logger
}

type Message struct {
	Type string         `json:"type"`
	Data map[string]any `json:"data"`
}

func NewClient(hub *Hub, conn *websocket.Conn, lobbyService LobbyService, logger *slog.Logger) *Client {
	return &Client{
		Hub:          hub,
		conn:         conn,
		lobbyService: lobbyService,
		logger:       logger,
		Send:         make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
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
		case createLobby:
			if clDto, ok := msgDto.Data["userId"]; ok {
				userId, ok := clDto.(string)
				if !ok {
					c.logger.Error(err.Error())
				}
				hub, err := c.lobbyService.CreateLobby(context.Background(), userId)
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

				lobbyDto := GetLobbyDto{
					Id:       hub.Id,
					Owner:    hub.Lobby.Owner,
					Players:  hub.Lobby.Players,
					Settings: hub.Lobby.Settings,
				}

				res, _ := json.Marshal(&lobbyDto)

				c.Send <- res
			}
		case joinLobby:
			//if clDto, ok := msgDto.Data["userId"]; ok {
			//
			//}
		default:

		}

		// создаем или подключаемся к хабу

		//обращаемся в сервис для создания, возвращает хабу

		//через горутину запускаем слушителя сообщений в хабу

		//там будут методы для изменения локального лобби игры

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

			fmt.Println(len(c.Send))

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
