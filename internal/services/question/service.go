package question

import (
	"errors"
	"loud-question/internal/model"
	"math/rand"
)

var qs = []model.Question{
	{
		Id:         "0",
		Question:   "Сколько месяцев в году имеют 28 дней?",
		Answer:     "Все",
		TimeGiving: 60,
	},
	{
		Id:         "1",
		Question:   "Согласно одной несерьезной новости, на открытии нового корпуса роддома президент... Что сделал?",
		Answer:     "Перерезал пуповину.",
		TimeGiving: 60,
	},
	{
		Id:         "2",
		Question:   "Что означает эмблема \"Мерседеса\"",
		Answer:     "Три стихии: земля",
		TimeGiving: 60,
	},
	{
		Id:         "3",
		Question:   "По-турецки даш - камень. Как будет звучать по-турецки черный камень",
		Answer:     "Карандаш",
		TimeGiving: 60,
	},
}

type Service struct {
	TempQuestions []model.Question
}

func New() *Service {
	return &Service{
		TempQuestions: qs,
	}
}

func (s *Service) GetQuestions() ([]model.Question, error) {
	return s.TempQuestions, nil
}

func (s *Service) GetQuestion(id string) (model.Question, error) {
	//var question model.Question

	for _, question := range s.TempQuestions {
		if question.Id == id {
			return question, nil
		}
	}

	return model.Question{}, errors.New("question not found")
}

func (s *Service) GetRandomQuestion() (model.Question, error) {

	i := rand.Intn(len(s.TempQuestions))
	//Check err
	return s.TempQuestions[i], nil
}
