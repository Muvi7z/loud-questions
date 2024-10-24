package question

import "loud-question/internal/model"

type QuestionsService interface {
	GetQuestions() ([]model.Question, error)
	GetQuestion(id string) (model.Question, error)
	GetRandomQuestion() (model.Question, error)
}
