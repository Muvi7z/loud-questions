package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"loud-question/internal/handler"
	"loud-question/internal/model"
)

type Hub struct {
	Clients      map[*Client]bool
	broadcast    chan []byte
	Register     chan *Client
	unregister   chan *Client
	Logger       *slog.Logger
	lobby        model.Lobby
	lobbyService handler.LobbyService
}

const (
	createLobby = "createLobby"
)

func NewHub(logger *slog.Logger, lobbyService handler.LobbyService) *Hub {
	return &Hub{
		Clients:      make(map[*Client]bool),
		broadcast:    make(chan []byte),
		Register:     make(chan *Client),
		unregister:   make(chan *Client),
		Logger:       logger,
		lobbyService: lobbyService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.Clients {
				select {
				case client.send <- message:

					var msgDto Message

					err := json.Unmarshal(message, &msgDto)
					if err != nil {
						h.Logger.Error(err.Error())
						return
					}

					fmt.Println(msgDto.Type, msgDto.Data)

					switch msgDto.Type {
					case createLobby:
						if clDto, ok := msgDto.Data["userId"]; ok {
							userId, ok := clDto.(string)
							if !ok {
								h.Logger.Error(err.Error())
							}
							lobby, err := h.lobbyService.CreateLobby(context.Background(), userId)
							if err != nil {
								h.Logger.Error(err.Error())
								return
							}
							h.lobby = lobby
						}
					}
				default:
					close(client.send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
