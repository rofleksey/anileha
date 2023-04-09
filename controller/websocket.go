package controller

import (
	"anileha/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func registerWebsocketController(
	config *config.Config,
	log *zap.Logger,
	engine *gin.Engine,
) {
	//upgrader := websocket.Upgrader{
	//	ReadBufferSize:  1024,
	//	WriteBufferSize: 1024,
	//	CheckOrigin: func(r *http.Request) bool {
	//		return true
	//	},
	//}
	//engine.GET("/ws/room/:roomId", func(c *gin.Context) {
	//	var authUser *db.AuthUser
	//
	//	session := sessions.Default(c)
	//	user := session.Get(rest.UserKey)
	//	if user != nil {
	//		authUser = user.(*db.AuthUser)
	//	}
	//
	//	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	//	if err != nil {
	//		c.Error(rest.ErrBadRequest(err.Error()))
	//		return
	//	}
	//})
}

var WebsocketControllerExport = fx.Options(fx.Invoke(registerWebsocketController))
