package main

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	handler2 "loud-question/internal/handler"
	"loud-question/internal/model"
	"net/http"
	"os"
)

func main() {
	router := gin.Default()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	handler := handler2.NewHandler(logger, make(map[string]model.User))

	handler.Register(router)

	s := &http.Server{
		Addr:    ":10000",
		Handler: router,
	}

	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
