package round

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"loud-question/internal/model"
	"loud-question/internal/services/question"
)

type Service struct {
	logger          *slog.Logger
	Sessions        map[string]model.Session
	playedQuestions []string
	QuestionService question.QuestionsService
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

	q, err := s.QuestionService.GetRandomQuestion()
	if err != nil {
		s.logger.Error(err.Error())
	}

	sessionID := uuid.New().String()

	session := model.Session{
		Id:       sessionID,
		Type:     model.QuestionType,
		LeaderId: "4",
		Status:   model.StartStatus,
		Question: q,
	}

	s.Sessions[sessionID] = session

	return session, nil
}
