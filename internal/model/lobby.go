package model

type Lobby struct {
	Id       string        `json:"id"`
	Owner    string        `json:"owner"`
	Players  []User        `json:"users"`
	Rounds   []Round       `json:"rounds"`
	Settings SettingsLobby `json:"settings"`
}

type Question struct {
	Id         string `json:"id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	TimeGiving int32  `json:"timeGiving"`
}

type SettingsLobby struct {
	Leaders      []string `json:"leaders"`
	Time         int32    `json:"time"`
	SessionCount int32    `json:"sessionCount"`
	SessionRound int32    `json:"sessionRound"`
	//Количество сессии
	//Количество раундов
}
