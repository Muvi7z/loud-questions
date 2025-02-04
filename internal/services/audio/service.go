package audio

import "log/slog"

type Service struct {
	logger *slog.Logger
}

func New(log *slog.Logger) *Service {
	return &Service{
		logger: log,
	}
}
