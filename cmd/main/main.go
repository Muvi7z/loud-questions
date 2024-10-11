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
	u := make(map[string]model.User)
	u["13ebe966-acaf-4d98-9014-7bb8527d00ae"] = model.User{
		Uuid:     "13ebe966-acaf-4d98-9014-7bb8527d00ae",
		Username: "Muvi",
		Score:    0,
	}
	u["3dd1a57c-a3b3-487a-babf-a2fc79a785cb"] = model.User{
		Uuid:     "3dd1a57c-a3b3-487a-babf-a2fc79a785cb",
		Username: "Haachamacha",
		Score:    0,
	}
	lobbyService := lobby.New(logger, u)

	handler := handler2.NewHandler(logger, lobbyService, lobbyService)

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
