package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"loud-question/internal/model"
	"sync"
	"time"
)

const (
	sendMessage  = "sendMessage"
	startSession = "startSession"
	endSession   = "endSession"
	startGame    = "startGame"
	joinGame     = "joinGame"
)

const (
	createLobby    = "createLobby"
	joinLobby      = "joinLobby"
	leftLobby      = "leftLobby"
	deleteLobby    = "deleteLobby"
	changeSettings = "changeSettings"
)

type LobbyService interface {
	CreateLobby(ctx context.Context, userId string) (*Hub, error)
	GetLobbies(ctx context.Context) map[string]model.Lobby
	GetHubs() map[string]*Hub
	JoinLobby(ctx context.Context, lobbyId string, userId string) (*Hub, error)
	DeleteLobby(ctx context.Context, idLobby string) error
	ChangeSettings(ctx context.Context, idLobby string, newSettings model.SettingsLobby) error
	LeftLobby(ctx context.Context, lobbyId string, userId string) (*Hub, error)
	StartSession(lobbyId string) (model.Session, error)
}

type RoundService interface {
	CreateSession(ctx context.Context, leaderId string, sessionType model.SessionType) (model.Session, error)
}

type UserService interface {
	AddUser(ctx context.Context, username string) string
	GetUser(ctx context.Context, id string) (model.User, error)
	GetUsers(ctx context.Context) (map[string]model.User, error)
}

type Hub struct {
	Id            string
	Clients       map[string]*Client
	Broadcast     chan Message
	Register      chan *Client
	Unregister    chan *Client
	Logger        *slog.Logger
	Lobby         model.Lobby
	lobbyService  LobbyService
	RoundService  RoundService
	mu            sync.Mutex
	startGameTime time.Time
	gameTimer     *time.Timer
}

func NewHub(logger *slog.Logger, lobbyService LobbyService, id string, lobby model.Lobby, roundService RoundService) *Hub {
	return &Hub{
		Id:           id,
		Clients:      make(map[string]*Client),
		Broadcast:    make(chan Message),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Logger:       logger,
		Lobby:        lobby,
		lobbyService: lobbyService,
		RoundService: roundService,
	}
}

func (h *Hub) StartGame(ctx context.Context) {
	for i, r := range h.Lobby.Rounds {
		if h.Lobby.CurrentRound == r.Id {
			for j, s := range r.Sessions {
				if h.Lobby.CurrentSession == s.Id {
					h.Lobby.Rounds[i].Sessions[j].Status = model.StartStatus
				}
			}

		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.startGameTime = time.Now()
	h.gameTimer = time.NewTimer(time.Duration(h.Lobby.Settings.Time) * time.Second)
	go func() {
		<-h.gameTimer.C
	}()
}

// Run Запуск хаба, получение сообщение и отправление ответа другим клиентам
func (h *Hub) Run() {
	for {
		fmt.Println(fmt.Sprintf("run hub %s", h.Id))
		select {
		case client := <-h.Register:
			clientId := client.User.Uuid
			h.Clients[clientId] = client
		case client := <-h.Unregister:
			clientId := client.User.Uuid
			for _, client := range h.Clients {
				if clientId != client.User.Uuid {
					resData, _ := json.Marshal(GetUserDto{
						User: h.Clients[clientId].User,
					})

					response := Message{
						Type: leftLobby,
						Data: resData,
					}

					// удалить из лобби
					resByte, err := json.Marshal(&response)
					if err != nil {
						h.Logger.Error("ошибка при выполнении marshal")
						break
					}

					client.Send <- resByte
				}
			}
			fmt.Println(h.Clients)
			if _, ok := h.Clients[clientId]; ok {
				_, err := h.lobbyService.LeftLobby(context.Background(), h.Id, client.User.Uuid)
				if err != nil {
					h.Logger.Error("error", err)
					break
				}

				delete(h.Clients, clientId)
				//close(client.Send)
			}
		case message := <-h.Broadcast:
			fmt.Println(h.Clients)

			switch message.Type {
			case changeSettings:
				for _, client := range h.Clients {
					var data model.SettingsLobby
					err := json.Unmarshal(message.Data, &data)
					if err != nil {
						msg, _ := json.Marshal(ErrorMessage{
							Message: "invalid credentials",
							Code:    "400",
						})
						client.Send <- msg
						break
					}

					err = h.lobbyService.ChangeSettings(context.Background(), h.Id, data)
					if err != nil {
						msg, _ := json.Marshal(ErrorMessage{
							Message: err.Error(),
							Code:    "500",
						})
						client.Send <- msg
						break
					}

					settingRes, _ := json.Marshal(h.Lobby.Settings)

					response := Message{
						Type: changeSettings,
						Data: settingRes,
					}

					// удалить из лобби
					resByte, err := json.Marshal(&response)
					if err != nil {
						h.Logger.Error("ошибка при выполнении marshal")
						break
					}
					client.Send <- resByte

				}
			case startGame:
				//Начинаем игру запускаем таймер
				h.StartGame(context.Background())
				lobbyDto := model.Lobby{
					Id:             h.Id,
					Owner:          h.Lobby.Owner,
					Players:        h.Lobby.Players,
					Rounds:         h.Lobby.Rounds,
					CurrentRound:   h.Lobby.CurrentRound,
					CurrentSession: h.Lobby.CurrentSession,
					Settings:       h.Lobby.Settings,
				}

				lobbyByte, err := json.Marshal(&lobbyDto)
				if err != nil {
					h.Logger.Error("ошибка при выполнении marshal")
					break
				}

				msg := Message{
					Type:   message.Type,
					SendBy: message.SendBy,
					Data:   lobbyByte,
				}

				msgByte, err := json.Marshal(&msg)
				if err != nil {
					h.Logger.Error("ошибка при выполнении marshal")
					break
				}

				for _, client := range h.Clients {
					client.StreamAudio()
					client.Send <- msgByte
				}

			case joinGame:
				h.mu.Lock()
				var session model.Session
				elapsed := time.Since(h.startGameTime)
				remaining := 19000*time.Second - elapsed //time.Duration(h.Lobby.Settings.Time)
				if remaining < 0 {
					remaining = 0
				}

				for _, r := range h.Lobby.Rounds {
					if h.Lobby.CurrentRound == r.Id {
						for _, s := range r.Sessions {
							if h.Lobby.CurrentSession == s.Id {
								session = s
							}
						}

					}
				}
				res := JoinGameDto{
					TimeGame:      int(remaining.Seconds()),
					MusicPosition: 0,
					Lobby:         h.Lobby,
					Session:       session,
				}

				resByte, _ := json.Marshal(res)

				msg := Message{
					Type:   message.Type,
					SendBy: message.SendBy,
					Data:   resByte,
				}
				msgByte, err := json.Marshal(&msg)
				if err != nil {
					h.Logger.Error("ошибка при выполнении marshal")
					break
				}

				for _, client := range h.Clients {
					if client.User.Uuid == message.SendBy {

					}
					client.Send <- msgByte

				}

				h.mu.Unlock()

			case joinLobby:
				for _, client := range h.Clients {
					resByte, err := json.Marshal(&message)
					if err != nil {
						h.Logger.Error("ошибка при выполнении marshal")
						break
					}
					client.Send <- resByte

				}
			case deleteLobby:
				for _, client := range h.Clients {
					close(client.Send)
					client.Hub = nil
					delete(h.Clients, client.User.Uuid)
				}
				close(h.Register)
				close(h.Unregister)
				close(h.Broadcast)
				err := h.lobbyService.DeleteLobby(context.Background(), h.Id)
				if err != nil {
					return
				}
				return
			case startSession:
				session, err := h.lobbyService.StartSession(h.Id)
				if err != nil {
					return
				}

				resData, _ := json.Marshal(StartSessionDto{
					Session: session,
				})

				response := Message{
					Type:   startSession,
					SendBy: message.SendBy,
					Data:   resData,
				}

				msgByte, err := json.Marshal(&response)
				if err != nil {
					h.Logger.Error("ошибка при выполнении marshal")
					break
				}
				for _, client := range h.Clients {
					client.Send <- msgByte

				}
			case endSession:
				for _, client := range h.Clients {

					//session, err := h.roundService.StartSession(context.Background(), h.Lobby)
					//if err != nil {
					//	return
					//}

					resData, _ := json.Marshal(StartSessionDto{
						Session: model.Session{},
					})

					response := Message{
						Type:   "startSession",
						SendBy: message.SendBy,
						Data:   resData,
					}

					msgByte, err := json.Marshal(&response)
					if err != nil {
						h.Logger.Error("ошибка при выполнении marshal")
						break
					}

					client.Send <- msgByte

				}
			case createLobby:
				for _, client := range h.Clients {
					if client.User.Uuid != message.SendBy {
						response := Message{
							Type: createLobby,
							Data: message.Data,
						}

						resByte, err := json.Marshal(&response)
						if err != nil {
							h.Logger.Error("ошибка при выполнении marshal")
							break
						}

						client.Send <- resByte
					}
				}
			case sendMessage:
				for _, client := range h.Clients {
					if client.User.Uuid != message.SendBy {

						msgByte, err := json.Marshal(message)
						if err != nil {
							h.Logger.Error("ошибка при выполнении marshal")
							break
						}

						client.Send <- msgByte
						//select {
						//case client.send <- message:
						//default:
						//	close(client.send)
						//	delete(h.Clients, client)
						//}
					}
				}
			}
		}
	}
}
