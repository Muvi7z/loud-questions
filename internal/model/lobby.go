package model

type Lobby struct {
	Owner    string        `json:"owner"`
	Players  []User        `json:"users"`
	Settings SettingsLobby `json:"settings"`
}

type SettingsLobby struct {
}
