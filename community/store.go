package community

import (
        "database/sql"
        "encoding/json"
        "fmt"
        "strings"

        "neoai/auth"
        "neoai/globals"
        "neoai/utils"
)

// store.go — pure database operations for the community package.
// No HTTP / gin concerns here, so the same functions can be unit-tested
// against an in-memory sqlite db if needed.

func intListToJSON(ids []int64) string {
        if ids == nil {
                return "[]"
        }
        b, _ := json.Marshal(ids)
        return string(b)
}

func jsonToIntList(s string) []int64 {
        if strings.TrimSpace(s) == "" {
                return nil
        }
        var out []int64
        if err := json.Unmarshal([]byte(s), &out); err != nil {
                return nil
        }
        return out
}

func stringListToJSON(ids []string) string {
        if ids == nil {
                return "[]"
        }
        b, _ := json.Marshal(ids)
        return string(b)
}

func jsonToStringList(s string) []string {
        if strings.TrimSpace(s) == "" {
                return nil
        }
        var out []string
        if err := json.Unmarshal([]byte(s), &out); err != nil {
                return nil
        }
        return out
}

// CreateChannel persists a new channel. ID is filled in by the DB.
func CreateChannel(db *sql.DB, ch *Channel) error {
        visibleRoles := stringListToJSON(ch.VisibleRoles)
        visibleUsers := intListToJSON(ch.VisibleUsers)
        senders := intListToJSON(ch.Senders)

        res, err := globals.ExecDb(db, `
                INSERT INTO community_channel
                        (name, topic, visibility, send_permission, visible_roles, visible_users, senders)
                VALUES (?, ?, ?, ?, ?, ?, ?)
        `, ch.Name, ch.Topic, string(ch.Visibility), string(ch.SendPerm),
                visibleRoles, visibleUsers, senders)
        if err != nil {
                return err
        }
        id, _ := res.LastInsertId()
        ch.ID = id
        return nil
}

// UpdateChannel fully replaces a channel's mutable fields.
func UpdateChannel(db *sql.DB, ch *Channel) error {
        visibleRoles := stringListToJSON(ch.VisibleRoles)
        visibleUsers := intListToJSON(ch.VisibleUsers)
        senders := intListToJSON(ch.Senders)

        _, err := globals.ExecDb(db, `
                UPDATE community_channel SET
                        name = ?, topic = ?, visibility = ?, send_permission = ?,
                        visible_roles = ?, visible_users = ?, senders = ?
                WHERE id = ?
        `, ch.Name, ch.Topic, string(ch.Visibility), string(ch.SendPerm),
                visibleRoles, visibleUsers, senders, ch.ID)
        return err
}

// DeleteChannel removes a channel and (via FK cascade) all of its messages.
func DeleteChannel(db *sql.DB, id int64) error {
        _, err := globals.ExecDb(db, `DELETE FROM community_channel WHERE id = ?`, id)
        return err
}

// GetChannel loads a single channel by id, including permission lists.
func GetChannel(db *sql.DB, id int64) (*Channel, error) {
        var (
                ch             Channel
                vis            string
                send           string
                visibleRoles   string
                visibleUsers   string
                senders        string
        )
        err := globals.QueryRowDb(db, `
                SELECT id, name, topic, visibility, send_permission,
                       visible_roles, visible_users, senders
                FROM community_channel WHERE id = ?
        `, id).Scan(&ch.ID, &ch.Name, &ch.Topic, &vis, &send,
                &visibleRoles, &visibleUsers, &senders)
        if err != nil {
                return nil, err
        }
        ch.Visibility = Visibility(vis)
        ch.SendPerm = SendPermission(send)
        ch.VisibleRoles = jsonToStringList(visibleRoles)
        ch.VisibleUsers = jsonToIntList(visibleUsers)
        ch.Senders = jsonToIntList(senders)
        return &ch, nil
}

// ListChannels returns every channel row, without filtering by user — the
// caller is responsible for filtering with `CanView` afterwards.
func ListChannels(db *sql.DB) ([]Channel, error) {
        rows, err := globals.QueryDb(db, `
                SELECT id, name, topic, visibility, send_permission,
                       visible_roles, visible_users, senders
                FROM community_channel
                ORDER BY id ASC
        `)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var out []Channel
        for rows.Next() {
                var (
                        ch           Channel
                        vis          string
                        send         string
                        visRoles     string
                        visUsers     string
                        senders      string
                )
                if err := rows.Scan(&ch.ID, &ch.Name, &ch.Topic, &vis, &send,
                        &visRoles, &visUsers, &senders); err != nil {
                        return nil, err
                }
                ch.Visibility = Visibility(vis)
                ch.SendPerm = SendPermission(send)
                ch.VisibleRoles = jsonToStringList(visRoles)
                ch.VisibleUsers = jsonToIntList(visUsers)
                ch.Senders = jsonToIntList(senders)
                out = append(out, ch)
        }
        return out, nil
}

// PostMessage inserts a new message row.
func PostMessage(db *sql.DB, msg *Message) error {
        res, err := globals.ExecDb(db, `
                INSERT INTO community_message (channel_id, user_id, username, avatar, content)
                VALUES (?, ?, ?, ?, ?)
        `, msg.ChannelID, msg.UserID, msg.Username, msg.Avatar, msg.Content)
        if err != nil {
                return err
        }
        id, _ := res.LastInsertId()
        msg.ID = id
        return nil
}

// EditMessage updates a message's content and stamps `edited_at`.
// Only the original author may edit.
func EditMessage(db *sql.DB, messageID, userID int64, content string) error {
        res, err := globals.ExecDb(db, `
                UPDATE community_message SET content = ?, edited_at = NOW()
                WHERE id = ? AND user_id = ?
        `, content, messageID, userID)
        if err != nil {
                return err
        }
        affected, _ := res.RowsAffected()
        if affected == 0 {
                return fmt.Errorf("message not found or not owned by user")
        }
        return nil
}

// DeleteMessage removes a message. Authors can delete their own; admins can
// delete anyone's. The `isAdmin` flag controls which path is used.
func DeleteMessage(db *sql.DB, messageID, userID int64, isAdmin bool) error {
        if isAdmin {
                _, err := globals.ExecDb(db, `DELETE FROM community_message WHERE id = ?`, messageID)
                return err
        }
        res, err := globals.ExecDb(db, `DELETE FROM community_message WHERE id = ? AND user_id = ?`,
                messageID, userID)
        if err != nil {
                return err
        }
        affected, _ := res.RowsAffected()
        if affected == 0 {
                return fmt.Errorf("message not found or not owned by user")
        }
        return nil
}

// ListMessages returns the most recent `limit` messages in a channel
// (default cap 200) ordered oldest-first so the client can render them
// directly without re-sorting.
func ListMessages(db *sql.DB, channelID int64, limit int) ([]Message, error) {
        if limit <= 0 || limit > 500 {
                limit = 200
        }
        // MySQL 8+ supports `LIMIT ?` directly, but the driver still wants
        // an int — pass it as a literal after bounds-checking above.
        rows, err := globals.QueryDb(db, fmt.Sprintf(`
                SELECT id, channel_id, user_id, username, avatar, content, created_at, edited_at
                FROM community_message
                WHERE channel_id = %d
                ORDER BY id DESC
                LIMIT %d
        `, channelID, limit))
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var out []Message
        for rows.Next() {
                var (
                        msg       Message
                        createdAt []uint8
                        editedAt  []uint8
                )
                if err := rows.Scan(&msg.ID, &msg.ChannelID, &msg.UserID, &msg.Username,
                        &msg.Avatar, &msg.Content, &createdAt, &editedAt); err != nil {
                        return nil, err
                }
                if t := utils.ConvertTime(createdAt); t != nil {
                        msg.CreatedAt = *t
                }
                if t := utils.ConvertTime(editedAt); t != nil {
                        msg.EditedAt = t
                }
                out = append(out, msg)
        }
        // reverse to oldest-first
        for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
                out[i], out[j] = out[j], out[i]
        }
        return out, nil
}

// CanView reports whether a user is allowed to see a channel.
// A nil user represents an anonymous visitor.
func CanView(db *sql.DB, ch *Channel, user *auth.User) bool {
        if user != nil && user.IsAdmin(db) {
                return true
        }
        switch ch.Visibility {
        case VisibilityPublic:
                return true
        case VisibilityMembers:
                return user != nil
        case VisibilityRoles:
                if user == nil {
                        return false
                }
                group := auth.GetGroup(db, user)
                for _, r := range ch.VisibleRoles {
                        if r == group {
                                return true
                        }
                }
                return false
        case VisibilityWhitelist:
                if user == nil {
                        return false
                }
                uid := user.GetID(db)
                for _, id := range ch.VisibleUsers {
                        if id == uid {
                                return true
                        }
                }
                return false
        }
        return false
}

// CanPost reports whether a user is allowed to send messages into a channel.
// Implicitly requires CanView — i.e. a user who can't see the channel
// can't post to it either.
func CanPost(db *sql.DB, ch *Channel, user *auth.User) bool {
        if user == nil {
                return false
        }
        if !CanView(db, ch, user) {
                return false
        }
        if user.IsAdmin(db) {
                return true
        }
        switch ch.SendPerm {
        case SendEveryone:
                return true
        case SendAdmins:
                return false
        case SendWhitelist:
                uid := user.GetID(db)
                for _, id := range ch.Senders {
                        if id == uid {
                                return true
                        }
                }
                return false
        }
        return false
}
