package round

import (
	"context"
	"github.com/google/uuid"
	"loud-question/internal/model"
	"time"
)

type Service struct {
	Sessions map[string]model.Session
}

func New() *Service {
	return &Service{}
}

func (s *Service) StartSession(ctx context.Context, lobby model.Lobby) (model.Session, error) {

	//Рандомим ведущего

	//Рандомим вопрос

	sessionID := uuid.New().String()

	session := model.Session{
		Id:       sessionID,
		Type:     model.QuestionType,
		LeaderId: "4",
		Question: model.Question{
			Id:         "1",
			Question:   "Q",
			Answer:     "E",
			TimeGiving: time.Minute,
		},
	}

	s.Sessions[sessionID] = session

	return session, nil
}
