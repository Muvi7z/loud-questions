package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"loud-question/internal/services/lobby"
)

type Hub struct {
	Clients      map[*Client]bool
	broadcast    chan []byte
	Register     chan *Client
	unregister   chan *Client
	Logger       *slog.Logger
	lobbyService lobby.LobbyService
}

const (
	createLobby = "createLobby"
	joinLobby   = "joinLobby"
)

func NewHub(logger *slog.Logger, lobbyService lobby.LobbyService) *Hub {
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
						l, err := h.lobbyService.CreateLobby(context.Background(), userId)
						if err != nil {
							h.Logger.Error(err.Error())
							msg, _ := json.Marshal(ErrorMessage{
								Message: err.Error(),
								Code:    "404",
							})
							client.lobbyId = l.Id
							client.send <- msg
						}

						l.Owner = "ss"
					}
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
