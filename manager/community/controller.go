package community

import (
	"chat/auth"
	"chat/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func parseAuth(c *gin.Context, token string) *auth.User {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}
	if strings.HasPrefix(token, "Bearer ") {
		token = token[7:]
	}
	if strings.HasPrefix(token, "sk-") {
		return auth.ParseApiKey(c, token)
	}
	return auth.ParseToken(c, token)
}

func bindChannelRequest(c *gin.Context) (*channelRequest, bool) {
	var form channelRequest
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, commonResponse{Status: false, Error: err.Error()})
		return nil, false
	}
	if strings.TrimSpace(form.Name) == "" {
		c.JSON(http.StatusOK, commonResponse{Status: false, Error: "name is required"})
		return nil, false
	}
	return &form, true
}

// ---- admin management endpoints (auto-gated by /admin path prefix) ----

func ListChannelsAdminAPI(c *gin.Context) {
	list, err := listChannelsAdmin(c)
	if err != nil {
		c.JSON(http.StatusOK, channelListResponse{Status: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, channelListResponse{Status: true, Data: list})
}

func CreateChannelAPI(c *gin.Context) {
	form, ok := bindChannelRequest(c)
	if !ok {
		return
	}
	username := utils.GetUserFromContext(c)
	ch, err := createChannel(c, *form, username)
	if err != nil {
		c.JSON(http.StatusOK, channelResponse{Status: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, channelResponse{Status: true, Data: ch})
}

func UpdateChannelAPI(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, channelResponse{Status: false, Error: "invalid id"})
		return
	}
	form, ok := bindChannelRequest(c)
	if !ok {
		return
	}
	ch, err := updateChannel(c, id, *form)
	if err != nil {
		c.JSON(http.StatusOK, channelResponse{Status: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, channelResponse{Status: true, Data: ch})
}

func DeleteChannelAPI(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, commonResponse{Status: false, Error: "invalid id"})
		return
	}
	if err := deleteChannel(c, id); err != nil {
		c.JSON(http.StatusOK, commonResponse{Status: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, commonResponse{Status: true})
}

// ---- user-facing endpoints ----

func ListChannelsAPI(c *gin.Context) {
	user := auth.RequireAuth(c)
	if user == nil {
		return
	}
	list, err := listChannelsForUser(c, user)
	if err != nil {
		c.JSON(http.StatusOK, channelListResponse{Status: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, channelListResponse{Status: true, Data: list})
}

func ListMessagesAPI(c *gin.Context) {
	user := auth.RequireAuth(c)
	if user == nil {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, messageListResponse{Status: false, Error: "invalid id"})
		return
	}
	db := utils.GetDBFromContext(c)
	ch, err := getChannelById(c, id)
	if err != nil || ch == nil || !canView(db, ch, user) {
		c.JSON(http.StatusOK, messageListResponse{Status: false, Error: "channel not found or no permission"})
		return
	}
	before, _ := strconv.Atoi(c.Query("before"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, err := listMessages(c, id, before, limit)
	if err != nil {
		c.JSON(http.StatusOK, messageListResponse{Status: false, Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, messageListResponse{Status: true, Data: list})
}

func SendMessageAPI(c *gin.Context) {
	user := auth.RequireAuth(c)
	if user == nil {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, sendMessageResponse{Status: false, Error: "invalid id"})
		return
	}
	var form sendRequest
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusOK, sendMessageResponse{Status: false, Error: err.Error()})
		return
	}
	db := utils.GetDBFromContext(c)
	ch, err := getChannelById(c, id)
	if err != nil || ch == nil || !canPost(db, ch, user) {
		c.JSON(http.StatusOK, sendMessageResponse{Status: false, Error: "no permission to post"})
		return
	}
	msg, err := sendMessage(c, id, user, form.Content)
	if err != nil {
		c.JSON(http.StatusOK, sendMessageResponse{Status: false, Error: err.Error()})
		return
	}
	if msg == nil {
		c.JSON(http.StatusOK, sendMessageResponse{Status: false, Error: "content is empty"})
		return
	}
	// broadcast to live websocket subscribers of this channel
	defaultHub.broadcast(id, wsOutgoing{Type: "message", Message: *msg})
	c.JSON(http.StatusOK, sendMessageResponse{Status: true, Message: msg})
}

// ChannelWSAPI upgrades to a websocket and streams channel messages in real time.
func ChannelWSAPI(c *gin.Context) {
	conn := utils.NewWebsocket(c, false)
	if conn == nil {
		return
	}
	defer conn.DeferClose()

	form, err := utils.ReadForm[wsAuthForm](conn)
	if err != nil {
		return
	}
	user := parseAuth(c, form.Token)
	if user == nil {
		return
	}

	db := utils.GetDBFromContext(c)
	ch, err := getChannelById(c, form.ChannelId)
	if err != nil || ch == nil {
		return
	}
	if !canView(db, ch, user) {
		return
	}

	cl := &client{channelId: ch.Id, conn: conn}
	defaultHub.register(cl)
	defer defaultHub.unregister(cl)

	for {
		var in wsIncoming
		if !conn.Receive(&in) {
			return
		}
		content := strings.TrimSpace(in.Content)
		if content == "" {
			continue
		}
		if !canPost(db, ch, user) {
			continue
		}
		msg, err := sendMessage(c, ch.Id, user, content)
		if err != nil || msg == nil {
			continue
		}
		defaultHub.broadcast(ch.Id, wsOutgoing{Type: "message", Message: *msg})
	}
}
