package lobby

import (
	"context"
	"loud-question/internal/model"
)

type LobbyService interface {
	CreateLobby(ctx context.Context, userId string) (model.Lobby, error)
	GetLobbies(ctx context.Context) map[string]model.Lobby
}

type UserService interface {
	AddUser(ctx context.Context, username string) string
	GetUser(ctx context.Context, id string) model.User
	GetUsers(ctx context.Context) (map[string]model.User, error)
}