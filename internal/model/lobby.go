package model

type Lobby struct {
	Id       string        `json:"id"`
	Owner    string        `json:"owner"`
	Players  []User        `json:"users"`
	Settings SettingsLobby `json:"settings"`
}

type SettingsLobby struct {
}
