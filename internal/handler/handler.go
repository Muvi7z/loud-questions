package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log/slog"
	"loud-question/internal/model"
	"net/http"
)

type Handler struct {
	logger *slog.Logger
	users  map[string]model.User
}

func NewHandler(logger *slog.Logger, users map[string]model.User) *Handler {
	return &Handler{
		logger: logger,
		users:  users,
	}
}

func (h *Handler) Register(router *gin.Engine) *gin.Engine {
	router.GET("/ws", h.WsConnect)
	router.POST("/signUp", h.SignUp)
	router.GET("/users", h.GetUsers)
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

	u := model.User{
		Uuid:     uuid.New().String(),
		Username: createUser.Username,
	}

	h.users[u.Uuid] = u

	c.JSON(200, gin.H{"token": u.Uuid})
}

func (h *Handler) GetUsers(c *gin.Context) {
	fmt.Println(h.users)

	c.JSON(200, h.users)
}

func (h *Handler) WsConnect(c *gin.Context) {

}
