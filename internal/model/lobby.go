package model

type Lobby struct {
	Owner    string        `json:"owner"`
	Players  []User        `json:"users"`
	Settings SettingsLobby `json:"settings"`
}

type Question struct {
	Id         string `json:"id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	TimeGiving int32  `json:"timeGiving"`
}

type SettingsLobby struct {
	Leaders []string `json:"leaders"`
	Time    int32    `json:"time"`
	//Количество сессии
	//Количество раундов
}
