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

func New(l *slog.Logger, questionService question.QuestionsService) *Service {
	return &Service{
		logger:          l,
		Sessions:        make(map[string]model.Session),
		QuestionService: questionService,
	}
}

func (s *Service) CreateSession(ctx context.Context, leaderId string, sessionType model.SessionType) (model.Session, error) {

	q, err := s.QuestionService.GetRandomQuestion()
	if err != nil {
		s.logger.Error(err.Error())
	}

	sessionID := uuid.New().String()

	session := model.Session{
		Id:       sessionID,
		Type:     sessionType,
		LeaderId: leaderId,
		Status:   model.StartStatus,
		Question: q,
	}

	s.Sessions[sessionID] = session

	return session, nil

}
