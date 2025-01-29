package model

type Lobby struct {
	Id             string        `json:"id"`
	Owner          string        `json:"owner"`
	Players        []User        `json:"users"`
	Rounds         []Round       `json:"rounds"`
	CurrentRound   string        `json:"currentRound"`
	CurrentSession string        `json:"currentSession"`
	Settings       SettingsLobby `json:"settings"`
}

type Question struct {
	Id         string `json:"id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	TimeGiving int32  `json:"timeGiving"`
}

type SettingsLobby struct {
	Leaders      []string `json:"leaders"`
	Time         int      `json:"time"`
	SessionCount int      `json:"sessionCount"`
	RoundCount   int      `json:"roundCount"`
	//Время даваемое за вопрос
}
