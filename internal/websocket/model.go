package websocket

import "loud-question/internal/model"

type ErrorMessage struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

type CreateLobbyDto struct {
	UserId string `json:"userId"`
}

type GetLobbyDto struct {
	Id       string              `json:"id"`
	Owner    string              `json:"owner"`
	Players  []model.User        `json:"users"`
	Settings model.SettingsLobby `json:"settings"`
}
