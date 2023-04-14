package controller

import (
	"anileha/config"
	"anileha/db"
	"anileha/rest"
	"anileha/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
)

func registerWebsocketController(
	config *config.Config,
	log *zap.Logger,
	engine *gin.Engine,
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

	engine.GET("/room", func(c *gin.Context) {
		rooms := roomService.GetAll()
		c.JSON(http.StatusOK, rooms)
	})

	engine.GET("/room/ws/:roomId", func(c *gin.Context) {
		roomIdString := c.Param("roomId")
		if roomIdString == "" {
			c.Error(rest.ErrBadRequest("blank room id"))
			return
		}

		session := sessions.Default(c)

		user := session.Get(rest.UserKey)
		if user == nil {
			c.Error(rest.ErrUnauthorizedInst)
			return
		}

		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.Error(rest.ErrInternal("failed to upgrade connection to websocket"))
			return
		}

		authUser := user.(*db.AuthUser)

		roomService.HandleConnection(ws, authUser, roomIdString)
	})
}

var WebsocketControllerExport = fx.Options(fx.Invoke(registerWebsocketController))
