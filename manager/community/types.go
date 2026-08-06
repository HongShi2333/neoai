package community

// Visibility constants for community channels.
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// Channel is a discord-like channel (频道).
type Channel struct {
	Id            int      `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Topic         string   `json:"topic"`
	Visibility    string   `json:"visibility"`
	VisibleGroups []string `json:"visible_groups"`
	PostGroups    []string `json:"post_groups"`
	Members       []string `json:"members"`
	Posters       []string `json:"posters"`
	Position      int      `json:"position"`
	CreatedBy     string   `json:"created_by"`
	CreatedAt     string   `json:"created_at"`
}

// Message is a persisted channel message synced to the database.
type Message struct {
	Id             int64  `json:"id"`
	ChannelId      int    `json:"channel_id"`
	SenderId       int64  `json:"sender_id"`
	SenderUsername string `json:"sender_username"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}

// channelRequest is the create/update payload for admin management.
type channelRequest struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Topic         string   `json:"topic"`
	Visibility    string   `json:"visibility"`
	VisibleGroups []string `json:"visible_groups"`
	PostGroups    []string `json:"post_groups"`
	Members       []string `json:"members"`
	Posters       []string `json:"posters"`
	Position      int      `json:"position"`
}

// sendRequest is the payload for sending a message.
type sendRequest struct {
	Content string `json:"content"`
}

// wsAuthForm is the first message over the websocket connection.
type wsAuthForm struct {
	Token     string `json:"token"`
	ChannelId int    `json:"channel_id"`
}

// wsIncoming is an incoming chat message over the websocket.
type wsIncoming struct {
	Content string `json:"content"`
}

// wsOutgoing is a broadcast chat message over the websocket.
type wsOutgoing struct {
	Type    string  `json:"type"` // "message" | "system"
	Message Message `json:"message"`
}

type commonResponse struct {
	Status bool   `json:"status"`
	Error  string `json:"error"`
}

type channelListResponse struct {
	Status bool      `json:"status"`
	Error  string    `json:"error"`
	Data   []Channel `json:"data"`
}

type channelResponse struct {
	Status bool     `json:"status"`
	Error  string   `json:"error"`
	Data   *Channel `json:"data"`
}

type messageListResponse struct {
	Status bool      `json:"status"`
	Error  string    `json:"error"`
	Data   []Message `json:"data"`
}

type sendMessageResponse struct {
	Status  bool     `json:"status"`
	Error   string   `json:"error"`
	Message *Message `json:"message"`
}
