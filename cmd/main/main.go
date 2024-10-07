package main

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	handler2 "loud-question/internal/handler"
	"loud-question/internal/model"
	"loud-question/internal/services/lobby"
	"net/http"
	"os"
)

func main() {
	router := gin.Default()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	lobbyService := lobby.New(logger, make(map[string]model.User))

	handler := handler2.NewHandler(logger, lobbyService)

	handler.Register(router)

	logger.Info("run server")
	//err := router.Run(":10000")
	//if err != nil {
	//	return
	//}

	s := &http.Server{
		Addr:    ":10000",
		Handler: router,
	}

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
