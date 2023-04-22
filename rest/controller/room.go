package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/rest/engine"
	"anileha/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
)

func registerWebsocketController(
	config *config.Config,
	log *zap.Logger,
	ginEngine *gin.Engine,
	userService *service.UserService,
	roomService *service.RoomService,
) {
	bufferSize := config.WebSocket.BufferSize

	upgrader := websocket.Upgrader{
		ReadBufferSize:  bufferSize,
		WriteBufferSize: bufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	authGroup := ginEngine.Group("/room")
	authGroup.Use(engine.AuthorizedMiddleware(log))

	authGroup.GET("", func(c *gin.Context) {
		rooms := roomService.GetAll()
		c.JSON(http.StatusOK, rooms)
	})

	authGroup.GET("/ws/:roomId", func(c *gin.Context) {
		roomIdString := c.Param("roomId")
		if roomIdString == "" {
			c.Error(engine.ErrBadRequest("blank room id"))
			return
		}

		authUser := c.MustGet(engine.UserKey).(*db.AuthUser)
		user, err := userService.GetById(authUser.ID)
		if err != nil {
			c.Error(err)
			return
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Error(engine.ErrInternal("failed to upgrade connection to websocket"))
			return
		}

		roomService.HandleConnection(ws, user, roomIdString)
	})
}

var WebsocketExport = fx.Options(fx.Invoke(registerWebsocketController))
