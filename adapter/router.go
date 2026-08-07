package adapter

import (
	"neoai/adapter/midjourney"
	"github.com/gin-gonic/gin"
)

func Register(app *gin.RouterGroup) {
	app.POST("/mj/notify", midjourney.NotifyAPI)
}
