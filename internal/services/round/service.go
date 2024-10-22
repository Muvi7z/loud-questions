package round

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"loud-question/internal/model"
	"time"
)

type Service struct {
	logger   *slog.Logger
	Sessions map[string]model.Session
}

func New(l *slog.Logger) *Service {
	return &Service{
		logger:   l,
		Sessions: make(map[string]model.Session),
	}
}

func (s *Service) StartSession(ctx context.Context, lobby model.Lobby) (model.Session, error) {

	//Рандомим ведущего

	//Рандомим вопрос

	sessionID := uuid.New().String()

	session := model.Session{
		Id:       sessionID,
		Type:     model.QuestionType,
		LeaderId: "4",
		Status:   model.StartStatus,
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
