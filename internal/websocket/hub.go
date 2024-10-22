package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"loud-question/internal/model"
)

type LobbyService interface {
	CreateLobby(ctx context.Context, userId string) (*Hub, error)
	GetLobbies(ctx context.Context) map[string]GetLobbyDto
	JoinLobby(ctx context.Context, client *Client, lobbyId string, userId string) (*Hub, error)
	DeleteLobby(ctx context.Context, idLobby string) error
}

type SessionService interface {
	StartSession(ctx context.Context, lobby model.Lobby) (model.Session, error)
}

type UserService interface {
	AddUser(ctx context.Context, username string) string
	GetUser(ctx context.Context, id string) (model.User, error)
	GetUsers(ctx context.Context) (map[string]model.User, error)
}

const (
	sendMessage  = "sendMessage"
	startSession = "startSession"
)

type Hub struct {
	Id             string
	Clients        map[string]*Client
	Broadcast      chan Message
	Register       chan *Client
	Unregister     chan *Client
	Logger         *slog.Logger
	Lobby          model.Lobby
	lobbyService   LobbyService
	sessionService SessionService
}

const (
	createLobby = "createLobby"
	joinLobby   = "joinLobby"
	deleteLobby = "deleteLobby"
)

func NewHub(logger *slog.Logger, lobbyService LobbyService, id string, lobby model.Lobby) *Hub {
	return &Hub{
		Id:           id,
		Clients:      make(map[string]*Client),
		Broadcast:    make(chan Message),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		Logger:       logger,
		Lobby:        lobby,
		lobbyService: lobbyService,
	}
}

func (h *Hub) Run() {
	for {
		fmt.Println(fmt.Sprintf("run hub %s", h.Id))
		select {
		case client := <-h.Register:
			clientId := client.User.Uuid
			h.Clients[clientId] = client
		case client := <-h.Unregister:
			clientId := client.User.Uuid
			if _, ok := h.Clients[clientId]; ok {
				delete(h.Clients, clientId)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			fmt.Println(h.Clients)
			// Добавить отправку всем сообщения что присоединился пользователь
			switch message.Type {
			case deleteLobby:
				for _, client := range h.Clients {
					close(client.Send)

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
				for _, client := range h.Clients {
					if client.User.Uuid != message.SendBy {

						session, err := h.sessionService.StartSession(context.Background(), h.Lobby)
						if err != nil {
							return
						}

						_ = session

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
