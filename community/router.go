package community

import (
        "encoding/json"
        "net/http"
        "strconv"
        "sync"
        "time"

        "neoai/auth"
        "neoai/utils"

        "github.com/gin-gonic/gin"
        "github.com/gorilla/websocket"
)

// router.go — HTTP + WebSocket endpoints for the community feature.
//
// REST endpoints (all under /community):
//   GET    /community/channels                — list channels the caller can see
//   POST   /community/channels                — admin: create a channel
//   GET    /community/channels/:id            — fetch a single channel
//   POST   /community/channels/:id           — admin: update a channel
//   DELETE /community/channels/:id           — admin: delete a channel
//   GET    /community/channels/:id/messages  — recent messages (query: ?limit=200)
//   POST   /community/channels/:id/messages  — post a message
//   POST   /community/messages/:id          — edit a message (own only)
//   DELETE /community/messages/:id           — delete (own, or any if admin)
//   GET    /community/ws                     — websocket real-time fanout
//
// WebSocket protocol:
//   Client sends:  {"action": "subscribe", "channel_id": <id>}
//                  {"action": "send",      "channel_id": <id>, "content": "..."}
//   Server pushes: {"type": "message", "data": <Message>}
//                  {"type": "delete",  "message_id": <id>}
//                  {"type": "error",   "message": "..."}

// ---------------------------------------------------------------------------
// in-memory connection registry
// ---------------------------------------------------------------------------

type subscriber struct {
        user      *auth.User
        conn      *websocket.Conn
        channels  map[int64]bool
        closeChan chan struct{}
        sendChan  chan []byte
}

var (
        subs   = make(map[*subscriber]struct{})
        subsMu sync.RWMutex
)

func registerSub(s *subscriber) {
        subsMu.Lock()
        subs[s] = struct{}{}
        subsMu.Unlock()
}

func unregisterSub(s *subscriber) {
        subsMu.Lock()
        delete(subs, s)
        subsMu.Unlock()
        close(s.sendChan)
}

func broadcastMessage(channelID int64, msg *Message) {
        payload, _ := json.Marshal(map[string]interface{}{
                "type": "message",
                "data": msg,
        })
        subsMu.RLock()
        defer subsMu.RUnlock()
        for s := range subs {
                if !s.channels[channelID] {
                        continue
                }
                select {
                case s.sendChan <- payload:
                default:
                        // drop if backpressured rather than block the writer
                }
        }
}

func broadcastDelete(channelID, messageID int64) {
        payload, _ := json.Marshal(map[string]interface{}{
                "type":       "delete",
                "channel_id": channelID,
                "message_id": messageID,
        })
        subsMu.RLock()
        defer subsMu.RUnlock()
        for s := range subs {
                if !s.channels[channelID] {
                        continue
                }
                select {
                case s.sendChan <- payload:
                default:
                }
        }
}

// ---------------------------------------------------------------------------
// REST handlers
// ---------------------------------------------------------------------------

type channelForm struct {
        Name         string         `json:"name"`
        Topic        string         `json:"topic"`
        Visibility  Visibility     `json:"visibility"`
        SendPerm     SendPermission `json:"send_permission"`
        VisibleRoles []string       `json:"visible_roles"`
        VisibleUsers []int64        `json:"visible_users"`
        Senders      []int64        `json:"senders"`
}

func toChannel(form channelForm) *Channel {
        return &Channel{
                Name:         form.Name,
                Topic:        form.Topic,
                Visibility:   form.Visibility,
                SendPerm:     form.SendPerm,
                VisibleRoles: form.VisibleRoles,
                VisibleUsers: form.VisibleUsers,
                Senders:      form.Senders,
        }
}

func ListChannelsAPI(c *gin.Context) {
        db := utils.GetDBFromContext(c)
        user := auth.GetUser(c)

        channels, err := ListChannels(db)
        if err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }

        // ?all=1 — admins can see every channel regardless of visibility
        // rules. Non-admin callers are always filtered by CanView.
        showAll := c.Query("all") == "1"
        isAdmin := user != nil && user.IsAdmin(db)

        out := make([]*Channel, 0, len(channels))
        for i := range channels {
                if showAll && isAdmin {
                        out = append(out, &channels[i])
                } else if CanView(db, &channels[i], user) {
                        out = append(out, &channels[i])
                }
        }
        c.JSON(http.StatusOK, gin.H{"status": true, "data": out})
}

func CreateChannelAPI(c *gin.Context) {
        if auth.RequireAdmin(c) == nil {
                return
        }
        var form channelForm
        if err := c.ShouldBindJSON(&form); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        if len(form.Name) == 0 || len(form.Name) > 64 {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": "name length must be 1..64"})
                return
        }
        if form.Visibility == "" {
                form.Visibility = VisibilityMembers
        }
        if form.SendPerm == "" {
                form.SendPerm = SendEveryone
        }
        db := utils.GetDBFromContext(c)
        ch := toChannel(form)
        if err := CreateChannel(db, ch); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, gin.H{"status": true, "data": ch})
}

func GetChannelAPI(c *gin.Context) {
        db := utils.GetDBFromContext(c)
        user := auth.GetUser(c)
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        ch, err := GetChannel(db, id)
        if err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": "channel not found"})
                return
        }
        if !CanView(db, ch, user) {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": "forbidden"})
                return
        }
        c.JSON(http.StatusOK, gin.H{"status": true, "data": ch})
}

func UpdateChannelAPI(c *gin.Context) {
        if auth.RequireAdmin(c) == nil {
                return
        }
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        var form channelForm
        if err := c.ShouldBindJSON(&form); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        db := utils.GetDBFromContext(c)
        ch := toChannel(form)
        ch.ID = id
        if err := UpdateChannel(db, ch); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, gin.H{"status": true, "data": ch})
}

func DeleteChannelAPI(c *gin.Context) {
        if auth.RequireAdmin(c) == nil {
                return
        }
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        db := utils.GetDBFromContext(c)
        if err := DeleteChannel(db, id); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, gin.H{"status": true})
}

type postMessageForm struct {
        Content string `json:"content"`
}

func ListMessagesAPI(c *gin.Context) {
        db := utils.GetDBFromContext(c)
        user := auth.GetUser(c)
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        ch, err := GetChannel(db, id)
        if err != nil || !CanView(db, ch, user) {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": "forbidden"})
                return
        }
        limit, _ := strconv.Atoi(c.Query("limit"))
        msgs, err := ListMessages(db, id, limit)
        if err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, gin.H{"status": true, "data": msgs})
}

func PostMessageAPI(c *gin.Context) {
        user := auth.RequireAuth(c)
        if user == nil {
                return
        }
        db := utils.GetDBFromContext(c)
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        ch, err := GetChannel(db, id)
        if err != nil || !CanPost(db, ch, user) {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": "forbidden"})
                return
        }
        var form postMessageForm
        if err := c.ShouldBindJSON(&form); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        if len(form.Content) == 0 || len(form.Content) > 8000 {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": "content length must be 1..8000"})
                return
        }
        msg := &Message{
                ChannelID: id,
                UserID:    user.GetID(db),
                Username:  user.Username,
                Content:   form.Content,
        }
        if err := PostMessage(db, msg); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        broadcastMessage(id, msg)
        c.JSON(http.StatusOK, gin.H{"status": true, "data": msg})
}

func EditMessageAPI(c *gin.Context) {
        user := auth.RequireAuth(c)
        if user == nil {
                return
        }
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        var form postMessageForm
        if err := c.ShouldBindJSON(&form); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        db := utils.GetDBFromContext(c)
        uid := user.GetID(db)
        if err := EditMessage(db, id, uid, form.Content); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, gin.H{"status": true})
}

func DeleteMessageAPI(c *gin.Context) {
        user := auth.RequireAuth(c)
        if user == nil {
                return
        }
        id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
        db := utils.GetDBFromContext(c)
        uid := user.GetID(db)
        isAdmin := user.IsAdmin(db)
        if err := DeleteMessage(db, id, uid, isAdmin); err != nil {
                c.JSON(http.StatusOK, gin.H{"status": false, "error": err.Error()})
                return
        }
        c.JSON(http.StatusOK, gin.H{"status": true})
}

// ---------------------------------------------------------------------------
// WebSocket endpoint
// ---------------------------------------------------------------------------

type wsAction struct {
        Action    string `json:"action"`
        ChannelID int64  `json:"channel_id"`
        Content   string `json:"content"`
        MessageID int64  `json:"message_id"`
}

func WsAPI(c *gin.Context) {
        user := auth.GetUser(c)
        if user == nil {
                c.JSON(http.StatusUnauthorized, gin.H{"status": false, "error": "auth required"})
                return
        }

        upgrader := utils.CheckUpgrader(c, true)
        conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
                return
        }

        sub := &subscriber{
                user:      user,
                conn:      conn,
                channels:  map[int64]bool{},
                closeChan: make(chan struct{}),
                sendChan:  make(chan []byte, 64),
        }
        registerSub(sub)

        // writer goroutine — drains sendChan into the websocket
        go func() {
                defer conn.Close()
                for {
                        select {
                        case <-sub.closeChan:
                                return
                        case msg, ok := <-sub.sendChan:
                                if !ok {
                                        return
                                }
                                _ = conn.WriteMessage(websocket.TextMessage, msg)
                        }
                }
        }()

        defer func() {
                unregisterSub(sub)
                conn.Close()
        }()

        // reader loop
        for {
                var action wsAction
                if err := conn.ReadJSON(&action); err != nil {
                        break
                }
                switch action.Action {
                case "subscribe":
                        db := utils.GetDBFromContext(c)
                        ch, err := GetChannel(db, action.ChannelID)
                        if err != nil || !CanView(db, ch, user) {
                                sendErr(sub, "forbidden")
                                continue
                        }
                        sub.channels[action.ChannelID] = true
                case "unsubscribe":
                        delete(sub.channels, action.ChannelID)
                case "send":
                        db := utils.GetDBFromContext(c)
                        ch, err := GetChannel(db, action.ChannelID)
                        if err != nil || !CanPost(db, ch, user) {
                                sendErr(sub, "forbidden")
                                continue
                        }
                        if len(action.Content) == 0 || len(action.Content) > 8000 {
                                sendErr(sub, "invalid content length")
                                continue
                        }
                        msg := &Message{
                                ChannelID: action.ChannelID,
                                UserID:    user.GetID(db),
                                Username:  user.Username,
                                Content:   action.Content,
                        }
                        if err := PostMessage(db, msg); err != nil {
                                sendErr(sub, err.Error())
                                continue
                        }
                        broadcastMessage(action.ChannelID, msg)
                case "delete":
                        db := utils.GetDBFromContext(c)
                        uid := user.GetID(db)
                        isAdmin := user.IsAdmin(db)
                        if err := DeleteMessage(db, action.MessageID, uid, isAdmin); err != nil {
                                sendErr(sub, err.Error())
                                continue
                        }
                        broadcastDelete(action.ChannelID, action.MessageID)
                }
        }
}

func sendErr(s *subscriber, msg string) {
        payload, _ := json.Marshal(map[string]interface{}{
                "type":    "error",
                "message": msg,
        })
        select {
        case s.sendChan <- payload:
        case <-time.After(500 * time.Millisecond):
        }
}

// Register wires every community endpoint onto the given router group.
func Register(app *gin.RouterGroup) {
        // Periodically evict dead subscribers — keeps the in-memory registry
        // from leaking when clients vanish without sending a close frame.
        go func() {
                for {
                        time.Sleep(5 * time.Minute)
                        subsMu.Lock()
                        for s := range subs {
                                if s.conn == nil {
                                        delete(subs, s)
                                }
                        }
                        subsMu.Unlock()
                }
        }()

        app.GET("/community/channels", ListChannelsAPI)
        app.POST("/community/channels", CreateChannelAPI)
        app.GET("/community/channels/:id", GetChannelAPI)
        app.POST("/community/channels/:id", UpdateChannelAPI)
        app.DELETE("/community/channels/:id", DeleteChannelAPI)
        app.GET("/community/channels/:id/messages", ListMessagesAPI)
        app.POST("/community/channels/:id/messages", PostMessageAPI)
        app.POST("/community/messages/:id", EditMessageAPI)
        app.DELETE("/community/messages/:id", DeleteMessageAPI)
        app.GET("/community/ws", WsAPI)
}
