package handler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"loud-question/internal/services/lobby"
	ws "loud-question/internal/websocket"
	"net/http"
)

type Handler struct {
	logger       *slog.Logger
	lobbyService lobby.LobbyService
}

func NewHandler(logger *slog.Logger, lobbyService lobby.LobbyService) *Handler {
	return &Handler{
		logger:       logger,
		lobbyService: lobbyService,
	}
}

func (h *Handler) Register(router *gin.Engine) *gin.Engine {
	router.GET("/createLobby", h.WsConnect)
	router.POST("/joinLobby", h.WsConnect)
	router.POST("/signUp", h.SignUp)
	router.GET("/users", h.GetUsers)
	router.GET("/lobbies", h.GetLobbies)
	return router
}

func (h *Handler) SignUp(c *gin.Context) {
	var createUser struct {
		Username string `json:"username"`
	}

	err := c.BindJSON(&createUser)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uuid := h.lobbyService.AddUser(context.Background(), createUser.Username)

	c.JSON(200, gin.H{"token": uuid})
}

func (h *Handler) GetUsers(c *gin.Context) {
	users, _ := h.lobbyService.GetUsers(context.Background())
	fmt.Println(users)

	c.JSON(200, users)
}
func (h *Handler) GetLobbies(c *gin.Context) {
	lobbies := h.lobbyService.GetLobbies(context.Background())
	fmt.Println(lobbies)

	c.JSON(200, lobbies)
}

func (h *Handler) WsConnect(c *gin.Context) {
	h.logger.Info("attending to connect ws")
	conn, err := ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error(err.Error())
		return
	}
	//defer func(conn *websocket.Conn) {
	//	err := conn.Close()
	//	if err != nil {
	//		h.logger.Error(err.Error())
	//		return
	//	}
	//}(conn)

	hub := ws.NewHub(h.logger, h.lobbyService)
	go hub.Run()

	client := ws.NewClient(hub, conn)
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
