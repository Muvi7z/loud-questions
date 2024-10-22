package model

import "time"

type Lobby struct {
	Owner    string        `json:"owner"`
	Players  []User        `json:"users"`
	Settings SettingsLobby `json:"settings"`
}

type Question struct {
	Id         string        `json:"id"`
	Question   string        `json:"question"`
	Answer     string        `json:"answer"`
	TimeGiving time.Duration `json:"timeGiving"`
}

type SettingsLobby struct {
}
