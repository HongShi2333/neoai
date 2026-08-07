package community

import "github.com/gin-gonic/gin"

func Register(app *gin.RouterGroup) {
	// admin management (auto-gated by the /admin path prefix via AuthMiddleware)
	app.GET("/admin/community/channel/list", ListChannelsAdminAPI)
	app.POST("/admin/community/channel/create", CreateChannelAPI)
	app.POST("/admin/community/channel/update/:id", UpdateChannelAPI)
	app.GET("/admin/community/channel/delete/:id", DeleteChannelAPI)

	// user-facing
	app.GET("/community/channels", ListChannelsAPI)
	app.GET("/community/channel/:id/messages", ListMessagesAPI)
	app.POST("/community/channel/:id/send", SendMessageAPI)
	app.GET("/community/ws", ChannelWSAPI)
}
