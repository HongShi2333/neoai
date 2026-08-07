package channel

import "github.com/gin-gonic/gin"

func Register(app *gin.RouterGroup) {
        app.GET("/info", GetInfo)
        app.GET("/attachments/:hash", AttachmentService)

        app.GET("/admin/channel/list", GetChannelList)
        app.POST("/admin/channel/create", CreateChannel)
        app.GET("/admin/channel/get/:id", GetChannel)
        app.POST("/admin/channel/update/:id", UpdateChannel)
        app.GET("/admin/channel/delete/:id", DeleteChannel)
        app.GET("/admin/channel/activate/:id", ActivateChannel)
        app.GET("/admin/channel/deactivate/:id", DeactivateChannel)

        // Auto-discover available models from an upstream OpenAI-compatible
        // /v1/models endpoint. Returns []string of model IDs.
        app.POST("/admin/channel/fetch-models", FetchModelsAPI)

        app.GET("/admin/charge/list", GetChargeList)
        app.POST("/admin/charge/set", SetCharge)
        app.GET("/admin/charge/delete/:id", DeleteCharge)
        app.POST("/admin/charge/sync", SyncCharge)
        // JSON bulk import/export of charge rules (new-api style).
        app.POST("/admin/charge/json", ApplyChargeJSONAPI)
        app.GET("/admin/charge/json", ExportChargeJSONAPI)

        app.GET("/admin/config/view", GetConfig)
        app.POST("/admin/config/update", UpdateConfig)

        app.GET("/admin/plan/view", GetPlanConfig)
        app.POST("/admin/plan/update", UpdatePlanConfig)
}
