package handler

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"log/slog"
	ws "loud-question/internal/websocket"
	"net/http"
	"os"
	"strconv"
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

func LiberalCORS(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE")

	if c.Request.Method == "OPTIONS" {
		if len(c.Request.Header["Access-Control-Request-Headers"]) > 0 {
			c.Header("Access-Control-Allow-Headers",
				c.Request.Header["Access-Control-Request-Headers"][0])
		}
		c.AbortWithStatus(http.StatusOK)
	}
}

func (h *Handler) Register(router *gin.Engine) *gin.Engine {
	router.GET("/ws", h.WsConnect)
	router.Use(LiberalCORS)
	router.POST("/signUp", h.SignUp)
	router.GET("/users", h.GetUsers)
	router.GET("/user/:userId", h.GetUser)
	router.GET("/lobbies", h.GetLobbies)
	router.GET("/hubs", h.GetHubs)
	router.GET("/song", h.GetSong)
	return router
}

func (h *Handler) GetSong(c *gin.Context) {
	rangeHeader := c.GetHeader("Range")
	if rangeHeader == "" {
		return
	}

	file, err := os.Open("rb.mp3") // Укажите путь к вашему аудиофайлу
	if err != nil {
		log.Println("Ошибка при открытии файла:", err)
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()

	fileSize := fileInfo.Size()

	start, end, err := ParseRangeHeader(rangeHeader, fileSize)

	/////Write
	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Length", strconv.FormatInt(end-start+1, 10))
	c.Header("Content-Type", "audio/mpeg")

	file.Seek(start, 0)
	buffer := make([]byte, 1024*8)
	toSend := end - start + 1

	for toSend > 0 {
		n, err := file.Read(buffer)
		if err != nil {
			break
		}
		if int64(n) > toSend {
			n = int(toSend)
		}

		c.Data(http.StatusPartialContent, "audio/mpeg", buffer[:n])
		toSend -= int64(n)
	}

	//dataChan := make(chan []byte)
	//var wg sync.WaitGroup
	//wg.Add(1)

	//go func() {
	//
	//	defer wg.Done()
	//
	//	buffer := make([]byte, 1024) // 1KB buffer size
	//	bytesToRead := end - start + 1
	//	for bytesToRead > 0 {
	//		n, err := file.Read(buffer)
	//		if err != nil && err != io.EOF {
	//			c.AbortWithStatus(http.StatusInternalServerError)
	//			return
	//		}
	//		if n == 0 {
	//			break
	//		}
	//		if int64(n) > bytesToRead {
	//			n = int(bytesToRead)
	//		}
	//		dataChan <- buffer[:n]
	//		bytesToRead -= int64(n)
	//	}
	//	close(dataChan)
	//}()
	//
	//go func() {
	//	defer wg.Wait()
	//	for chunk := range dataChan {
	//		c.Data(http.StatusOK, "audio/mpeg", chunk)
	//	}
	//}()
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

func (h *Handler) GetUser(c *gin.Context) {
	if userId := c.Param("userId"); userId != "" {
		users, _ := h.userService.GetUsers(context.Background())
		user, ok := users[userId]
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(200, user)
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		return
	}

}

func (h *Handler) GetLobbies(c *gin.Context) {
	lobbies := h.lobbyService.GetLobbies(context.Background())
	fmt.Println(lobbies)

	c.JSON(200, lobbies)
}

func (h *Handler) GetHubs(c *gin.Context) {
	hubs := h.lobbyService.GetHubs()

	c.JSON(200, &hubs)
}

func (h *Handler) WsConnect(c *gin.Context) {
	h.logger.Info("attending to connect ws")

	ws.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
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

	client := ws.NewClient(nil, conn, h.lobbyService, h.userService, h.logger)

	go client.WritePump()
	go client.ReadPump()
}
