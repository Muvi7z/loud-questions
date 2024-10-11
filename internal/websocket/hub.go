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
}

type UserService interface {
	AddUser(ctx context.Context, username string) string
	GetUser(ctx context.Context, id string) (model.User, error)
	GetUsers(ctx context.Context) (map[string]model.User, error)
}

const (
	sendMessage = "sendMessage"
)

type Hub struct {
	Id           string
	Clients      map[string]*Client
	Broadcast    chan Message
	Register     chan *Client
	Unregister   chan *Client
	Logger       *slog.Logger
	Lobby        model.Lobby
	lobbyService LobbyService
}

const (
	createLobby = "createLobby"
	joinLobby   = "joinLobby"
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
			msgByte, err := json.Marshal(message)
			if err != nil {
				h.Logger.Error("ошибка при выполнении marshal")
				break
			}

			switch message.Type {
			case sendMessage:
				for _, client := range h.Clients {
					if client.User.Uuid != message.SendBy {
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
