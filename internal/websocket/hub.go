package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"loud-question/internal/model"
)

type LobbyService interface {
	CreateLobby(ctx context.Context, userId string) (Hub, error)
	GetLobbies(ctx context.Context) map[string]GetLobbyDto
	JoinLobby(ctx context.Context, lobbyId string) Hub
}

type UserService interface {
	AddUser(ctx context.Context, username string) string
	GetUser(ctx context.Context, id string) model.User
	GetUsers(ctx context.Context) (map[string]model.User, error)
}

type Hub struct {
	Id           string
	Clients      map[string]*Client
	Broadcast    chan []byte
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
		Broadcast:    make(chan []byte),
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
			for _, client := range h.Clients {

				var msgDto Message

				err := json.Unmarshal(message, &msgDto)
				if err != nil {
					h.Logger.Error(err.Error())
					return
				}

				fmt.Println(msgDto.Type, msgDto.Data)
				switch msgDto.Type {
				case createLobby:
					fmt.Println(client, h.Lobby)
				case joinLobby:

				}
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
