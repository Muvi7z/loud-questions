package handler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	ws "loud-question/internal/websocket"
	"net/http"
)

type Handler struct {
	logger       *slog.Logger
	lobbyService ws.LobbyService
	userService  ws.UserService
}

func NewHandler(logger *slog.Logger, lobbyService ws.LobbyService, userService ws.UserService) *Handler {
	return &Handler{
		logger:       logger,
		lobbyService: lobbyService,
		userService:  userService,
	}
}

func (h *Handler) Register(router *gin.Engine) *gin.Engine {
	router.GET("/ws", h.WsConnect)
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

	uuid := h.userService.AddUser(context.Background(), createUser.Username)

	c.JSON(200, gin.H{"token": uuid})
}

func (h *Handler) GetUsers(c *gin.Context) {
	users, _ := h.userService.GetUsers(context.Background())
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

	client := ws.NewClient(nil, conn, h.lobbyService, h.logger)

	go client.WritePump()
	go client.ReadPump()
}
