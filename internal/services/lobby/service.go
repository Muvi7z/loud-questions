package lobby

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/sqids/sqids-go"
	"log/slog"
	"loud-question/internal/model"
	"loud-question/internal/services/question"
	"loud-question/internal/services/round"
	"loud-question/internal/websocket"
)

// Управляет хабами
type Service struct {
	logger *slog.Logger
	users  map[string]model.User
	hubs   map[string]*websocket.Hub
}

func New(log *slog.Logger, users map[string]model.User) *Service {
	return &Service{
		logger: log,
		users:  users,
		hubs:   make(map[string]*websocket.Hub),
	}
}

func (s *Service) AddUser(ctx context.Context, username string) string {

	u := model.User{
		Uuid:     uuid.New().String(),
		Username: username,
	}
	s.users[u.Uuid] = u

	return u.Uuid
}

func (s *Service) GetUser(ctx context.Context, id string) (model.User, error) {
	u, ok := s.users[id]
	if !ok {
		return model.User{}, errors.New("user not found")
	}
	return u, nil
}

func (s *Service) GetUsers(ctx context.Context) (map[string]model.User, error) {
	return s.users, nil
}

func (s *Service) CreateLobby(ctx context.Context, userId string) (*websocket.Hub, error) {
	//проверка того что юзер существует

	allId := make([]string, len(s.hubs)*2)

	i := 0
	for _, h := range s.hubs {
		allId[i] = h.Id
		i++
	}

	sqid, _ := sqids.New(sqids.Options{
		Blocklist: allId,
	})
	id, _ := sqid.Encode([]uint64{1, 2, 3})

	s.logger.Info("creating lobby", id)

	u, ok := s.users[userId]
	if !ok {
		s.logger.Error("user not found", userId)
		return nil, errors.New("user not found")
	}
	roundId := uuid.New().String()
	l := model.Lobby{
		Id:      id,
		Owner:   userId,
		Players: []model.User{u},
		Settings: model.SettingsLobby{
			SessionCount: 4,
			RoundCount:   1,
			Time:         90,
		},
		CurrentRound: roundId,
		Rounds: []model.Round{
			{
				Id:       roundId,
				Sessions: nil,
			},
		},
	}

	questionService := question.New()

	sessionService := round.New(s.logger, questionService)

	hub := websocket.NewHub(s.logger, s, id, l, sessionService)

	hub.Lobby = l

	s.hubs[id] = hub

	return hub, nil
}

func (s *Service) JoinLobby(ctx context.Context, lobbyId string, userId string) (*websocket.Hub, error) {
	if h, ok := s.hubs[lobbyId]; ok {

		u, err := s.GetUser(ctx, userId)
		if err != nil {
			return nil, err
		}
		fmt.Println()
		existPlayer := false
		for _, player := range h.Lobby.Players {
			if player.Uuid == userId {
				existPlayer = true
			}
		}
		if existPlayer {
			return nil, errors.New("player already exist")
		}
		h.Lobby.Players = append(h.Lobby.Players, u)

		s.hubs[lobbyId] = h

		return h, nil
	}
	return nil, errors.New("not found lobby")
}

func (s *Service) DeleteLobby(ctx context.Context, idLobby string) error {
	delete(s.hubs, idLobby)
	return nil
}

func (s *Service) LeftLobby(ctx context.Context, lobbyId string, userId string) (*websocket.Hub, error) {
	if h, ok := s.hubs[lobbyId]; ok {

		var deleteId int
		for i, item := range h.Lobby.Players {
			if item.Uuid == userId {
				deleteId = i
			}
		}

		h.Lobby.Players = append(h.Lobby.Players[:deleteId], h.Lobby.Players[deleteId+1:]...)

		//s.hubs[lobbyId] = h

		return h, nil
	}
	return nil, errors.New("not found lobby")
}

func (s *Service) StartSession(lobbyId string) (model.Session, error) {
	if h, ok := s.hubs[lobbyId]; ok {
		if len(h.Lobby.Settings.Leaders) > 0 {
			//Рандомим ведущего
			var session model.Session
			var err error

			for i, r := range h.Lobby.Rounds {
				if h.Lobby.CurrentRound == r.Id {
					if len(r.Sessions) > h.Lobby.Settings.SessionCount {
						session, err = h.RoundService.CreateSession(context.Background(), h.Lobby.Settings.Leaders[0], model.SuperGameType)
						if err != nil {
							return model.Session{}, err
						}

					} else {
						session, err = h.RoundService.CreateSession(context.Background(), h.Lobby.Settings.Leaders[0], model.QuestionType)
						if err != nil {
							return model.Session{}, err
						}
					}
					r.Sessions = append(r.Sessions, session)
					h.Lobby.Rounds[i] = r
					h.Lobby.CurrentSession = session.Id
					return session, nil
				}
			}

		}
	}
	return model.Session{}, errors.New("no leader set")
}

func (s *Service) ChangeSettings(ctx context.Context, lobbyId string, newSettings model.SettingsLobby) error {
	if h, ok := s.hubs[lobbyId]; ok {
		h.Lobby.Settings = newSettings
		return nil
	}

	return errors.New("not found lobby")
}

func (s *Service) GetLobby(ctx context.Context, username string) model.Lobby {
	panic("implement me")
}

func (s *Service) GetLobbies(ctx context.Context) map[string]model.Lobby {
	ls := make(map[string]model.Lobby)

	for k, v := range s.hubs {
		ls[k] = model.Lobby{
			Id:       v.Id,
			Owner:    v.Lobby.Owner,
			Players:  v.Lobby.Players,
			Settings: v.Lobby.Settings,
		}
	}
	return ls
}

func (s *Service) GetHubs() map[string]*websocket.Hub {
	return s.hubs
}
