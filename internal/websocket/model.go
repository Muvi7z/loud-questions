package websocket

import "loud-question/internal/model"

type ErrorMessage struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type CreateLobbyDto struct {
	UserId string `json:"userId"`
}

type GetUserDto struct {
	User model.User `json:"user"`
}

type StartSessionDto struct {
	Session model.Session `json:"session"`
}

type JoinLobbyDto struct {
	UserId  string `json:"userId"`
	LobbyId string `json:"lobbyId"`
}

type GetSessionDto struct {
	Id     string `json:"id"`
	Leader string `json:"leader"`
}
