package lobby

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sqids/sqids-go"
	"log/slog"
	"loud-question/internal/model"
)

type Service struct {
	logger  *slog.Logger
	users   map[string]model.User
	lobbies map[string]model.Lobby
}

func New(log *slog.Logger, users map[string]model.User) *Service {
	return &Service{
		logger:  log,
		users:   users,
		lobbies: make(map[string]model.Lobby),
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

func (s *Service) GetUser(ctx context.Context, id string) model.User {
	return s.users[id]
}

func (s *Service) GetUsers(ctx context.Context) (map[string]model.User, error) {
	return s.users, nil
}

func (s *Service) CreateLobby(ctx context.Context, userId string) (model.Lobby, error) {
	allId := make([]string, len(s.lobbies)*2)

	i := 0
	for k, _ := range s.lobbies {
		allId[i] = k
		i++
	}

	sqid, _ := sqids.New(sqids.Options{
		Blocklist: []string{"86Rf07"},
	})
	id, _ := sqid.Encode([]uint64{1, 2, 3})

	s.logger.Info("creating lobby", id)

	u, ok := s.users[userId]
	if !ok {
		s.logger.Error("user not found", userId)
		return model.Lobby{}, errors.New("user not found")
	}

	l := model.Lobby{
		Id:       id,
		Owner:    userId,
		Players:  []model.User{u},
		Settings: model.SettingsLobby{},
	}

	s.lobbies[id] = l

	return l, nil
}

func (s *Service) GetLobby(ctx context.Context, username string) model.Lobby {
	panic("implement me")
}

func (s *Service) GetLobbies(ctx context.Context, usernames []string) map[string]model.Lobby {
	panic("implement me")
}
