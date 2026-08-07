package community

import "time"

// Visibility controls who can SEE a channel in the sidebar / listing.
type Visibility string

const (
	// VisibilityPublic — anyone (including anonymous visitors) can see and read.
	VisibilityPublic Visibility = "public"
	// VisibilityMembers — any logged-in user can see and read.
	VisibilityMembers Visibility = "members"
	// VisibilityRoles — only users whose role is in `VisibleRoles` can see.
	VisibilityRoles Visibility = "roles"
	// VisibilityWhitelist — only user IDs in `VisibleUsers` can see.
	VisibilityWhitelist Visibility = "whitelist"
)

// SendPermission controls who can POST messages into a channel.
type SendPermission string

const (
	// SendEveryone — any user who can see the channel may post.
	SendEveryone SendPermission = "everyone"
	// SendAdmins — only admins may post (announcement-style).
	SendAdmins SendPermission = "admins"
	// SendWhitelist — only user IDs in `Senders` may post.
	SendWhitelist SendPermission = "whitelist"
)

// Channel is a Discord-style text channel.
//
// Permissions are split into two layers:
//   1. Visibility — who can see the channel exists at all.
//   2. SendPermission — among those who can see, who can post.
//
// Both layers are evaluated on every API request by `CanView` / `CanPost`,
// so admins can change the rules on the fly without restarting.
type Channel struct {
	ID            int64          `json:"id"`
	Name          string         `json:"name"`
	Topic         string         `json:"topic"`
	Visibility    Visibility    `json:"visibility"`
	SendPerm      SendPermission `json:"send_permission"`
	VisibleRoles  []string       `json:"visible_roles,omitempty"`
	VisibleUsers  []int64        `json:"visible_users,omitempty"`
	Senders       []int64        `json:"senders,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// Message is a single chat message inside a channel.
//
// Messages are persisted to MySQL (`community_message`) AND pushed in
// real-time over WebSocket to all viewers of the channel, so DB and
// live state stay in sync.
type Message struct {
	ID        int64     `json:"id"`
	ChannelID int64     `json:"channel_id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`  // denormalised for fast rendering
	Avatar    string    `json:"avatar,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	// EditedAt is non-null when the message has been edited.
	EditedAt *time.Time `json:"edited_at,omitempty"`
}
