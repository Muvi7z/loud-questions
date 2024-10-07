package websocket

type CreateLobbyDto struct {
	UserId string `json:"userId"`
}

type ErrorMessage struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
