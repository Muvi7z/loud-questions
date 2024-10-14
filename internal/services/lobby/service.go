package lobby

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sqids/sqids-go"
	"log/slog"
	"loud-question/internal/model"
	"loud-question/internal/websocket"
)

// Service
type Service struct {
	logger *slog.Logger
	users  map[string]model.User
	hubs   map[string]websocket.Hub
}

func New(log *slog.Logger, users map[string]model.User) *Service {
	return &Service{
		logger: log,
		users:  users,
		hubs:   make(map[string]websocket.Hub),
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
		return model.User{}, errors.New("not found user")
	}
	return u, nil
}

func (s *Service) GetUsers(ctx context.Context) (map[string]model.User, error) {
	return s.users, nil
}

func (s *Service) CreateLobby(ctx context.Context, userId string) (*websocket.Hub, error) {
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

	l := model.Lobby{
		Owner:    userId,
		Players:  []model.User{u},
		Settings: model.SettingsLobby{},
	}

	hub := websocket.NewHub(s.logger, s, id, l)

	hub.Lobby = l

	s.hubs[id] = *hub

	return hub, nil
}

func (s *Service) JoinLobby(ctx context.Context, client *websocket.Client, lobbyId string, userId string) (*websocket.Hub, error) {
	if h, ok := s.hubs[lobbyId]; ok {

		u, err := s.GetUser(ctx, userId)
		if err != nil {
			return nil, err
		}

		h.Lobby.Players = append(h.Lobby.Players, u)

		s.hubs[lobbyId] = h

		return &h, nil
	}
	return nil, errors.New("not found lobby")
}

func (s *Service) DeleteLobby(ctx context.Context, idLobby string) error {
	delete(s.hubs, idLobby)
	return nil
}

func (s *Service) GetLobby(ctx context.Context, username string) model.Lobby {
	panic("implement me")
}

func (s *Service) GetLobbies(ctx context.Context) map[string]websocket.GetLobbyDto {
	ls := make(map[string]websocket.GetLobbyDto)

	for k, v := range s.hubs {
		ls[k] = websocket.GetLobbyDto{
			Id:       v.Id,
			Owner:    v.Lobby.Owner,
			Players:  v.Lobby.Players,
			Settings: v.Lobby.Settings,
		}
	}
	return ls
}
