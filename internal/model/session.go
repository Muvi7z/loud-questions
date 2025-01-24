package model

type SessionType string

// type round
const (
	QuestionType  SessionType = "Question"
	SuperGameType             = "SuperGame"
)

// StartStatus status
const (
	StartStatus = "startStatus"
)

type Session struct {
	Id       string      `json:"id"`
	Type     SessionType `json:"type"`
	LeaderId string      `json:"leaderId"`
	Question Question    `json:"question"`
	Status   string      `json:"status"`
	IsWin    bool        `json:"isWin"`
}

type Round struct {
	Id       string    `json:"id"`
	Sessions []Session `json:"sessions"`
}
