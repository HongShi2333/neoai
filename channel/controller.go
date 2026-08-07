package channel

import (
	"chat/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type SyncChargeForm struct {
	Overwrite bool           `json:"overwrite"`
	Data      ChargeSequence `json:"data"`
}

type FetchModelsForm struct {
	Endpoint string `json:"endpoint"`
	Secret   string `json:"secret"`
}

type ModelsResponse struct {
	Object string         `json:"object"`
	Data   []ModelItem    `json:"data"`
}

type ModelItem struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

func FetchChannelModels(c *gin.Context) {
	var form FetchModelsForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	endpoint := strings.TrimSpace(form.Endpoint)
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "endpoint is required",
		})
		return
	}

	if strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint[:len(endpoint)-1]
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if form.Secret != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", form.Secret)
	}

	raw, err := utils.GetRaw(fmt.Sprintf("%s/v1/models", endpoint), headers)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	resp, err := utils.UnmarshalString[ModelsResponse](raw)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	models := make([]string, 0)
	for _, item := range resp.Data {
		if item.Id != "" {
			models = append(models, item.Id)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   models,
	})
}

func GetInfo(c *gin.Context) {
	c.JSON(http.StatusOK, SystemInstance.AsInfo())
}

func AttachmentService(c *gin.Context) {
	// /attachments/:hash -> ~/storage/attachments/:hash
	hash := c.Param("hash")
	c.File(fmt.Sprintf("storage/attachments/%s", hash))
}

func DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.DeleteChannel(utils.ParseInt(id))
	if state == nil {
		ConduitInstance.Load()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func ActivateChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.ActivateChannel(utils.ParseInt(id))
	if state == nil {
		ConduitInstance.Load()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func DeactivateChannel(c *gin.Context) {
	id := c.Param("id")
	state := ConduitInstance.DeactivateChannel(utils.ParseInt(id))
	if state == nil {
		ConduitInstance.Load()
	}

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetChannelList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   ConduitInstance.Sequence,
	})
}

func GetChannel(c *gin.Context) {
	id := c.Param("id")
	channel := ConduitInstance.Sequence.GetChannelById(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": channel != nil,
		"data":   channel,
	})
}

func CreateChannel(c *gin.Context) {
	var channel Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := ConduitInstance.CreateChannel(&channel)
	if state == nil {
		ConduitInstance.Load()
	}
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func UpdateChannel(c *gin.Context) {
	var channel Channel
	if err := c.ShouldBindJSON(&channel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	id := c.Param("id")
	channel.Id = utils.ParseInt(id)

	state := ConduitInstance.UpdateChannel(channel.Id, &channel)
	if state == nil {
		ConduitInstance.Load()
	}
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func SetCharge(c *gin.Context) {
	var charge Charge
	if err := c.ShouldBindJSON(&charge); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := ChargeInstance.SetRule(charge)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetChargeList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   ChargeInstance.ListRules(),
	})
}

func DeleteCharge(c *gin.Context) {
	id := c.Param("id")
	state := ChargeInstance.DeleteRule(utils.ParseInt(id))

	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func SyncCharge(c *gin.Context) {
	var form SyncChargeForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
	}

	state := ChargeInstance.SyncRules(form.Data, form.Overwrite)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"data":   SystemInstance,
	})
}

func UpdateConfig(c *gin.Context) {
	var config SystemConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := SystemInstance.UpdateConfig(&config)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}

func GetPlanConfig(c *gin.Context) {
	c.JSON(http.StatusOK, PlanInstance)
}

func UpdatePlanConfig(c *gin.Context) {
	var config PlanManager
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	state := PlanInstance.UpdateConfig(&config)
	c.JSON(http.StatusOK, gin.H{
		"status": state == nil,
		"error":  utils.GetError(state),
	})
}
