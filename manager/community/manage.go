package community

import (
	"chat/auth"
	"chat/globals"
	"chat/utils"
	"database/sql"
	"strings"

	"github.com/gin-gonic/gin"
)

// parseSlice decodes a JSON string column into a string slice.
func parseSlice(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	if slice := utils.UnmarshalForm[[]string](raw); slice != nil {
		return *slice
	}
	return []string{}
}

func formatTime(createdAt []uint8) string {
	if t := utils.ConvertTime(createdAt); t != nil {
		return t.Format("2006-01-02 15:04:05")
	}
	return ""
}

func normalizeVisibility(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case VisibilityPrivate:
		return VisibilityPrivate
	default:
		return VisibilityPublic
	}
}

// scanChannel scans a channel row from the given columns.
// Column order: id, name, description, topic, visibility, visible_groups,
// post_groups, members, posters, position, created_by, created_at.
func scanChannel(rows *sql.Rows) (Channel, error) {
	var ch Channel
	var description, topic, visibleGroups, postGroups, members, posters, createdBy sql.NullString
	var createdAt []uint8
	if err := rows.Scan(
		&ch.Id, &ch.Name, &description, &topic, &ch.Visibility,
		&visibleGroups, &postGroups, &members, &posters,
		&ch.Position, &createdBy, &createdAt,
	); err != nil {
		return ch, err
	}
	ch.Description = description.String
	ch.Topic = topic.String
	ch.VisibleGroups = parseSlice(visibleGroups.String)
	ch.PostGroups = parseSlice(postGroups.String)
	ch.Members = parseSlice(members.String)
	ch.Posters = parseSlice(posters.String)
	ch.CreatedBy = createdBy.String
	ch.CreatedAt = formatTime(createdAt)
	ch.Visibility = normalizeVisibility(ch.Visibility)
	return ch, nil
}

func channelSelectColumns() string {
	return `id, name, description, topic, visibility, visible_groups, post_groups, members, posters, position, created_by, created_at`
}

// listChannelsAdmin returns all channels (admin view).
func listChannelsAdmin(c *gin.Context) ([]Channel, error) {
	db := utils.GetDBFromContext(c)
	rows, err := globals.QueryDb(db, `
		SELECT `+channelSelectColumns()+` FROM community_channel
		ORDER BY position ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Channel
	for rows.Next() {
		ch, err := scanChannel(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, ch)
	}
	return list, nil
}

// listChannelsForUser returns only channels the user is allowed to view.
func listChannelsForUser(c *gin.Context, user *auth.User) ([]Channel, error) {
	all, err := listChannelsAdmin(c)
	if err != nil {
		return nil, err
	}
	db := utils.GetDBFromContext(c)
	var list []Channel
	for _, ch := range all {
		if canView(db, &ch, user) {
			list = append(list, ch)
		}
	}
	return list, nil
}

func getChannelById(c *gin.Context, id int) (*Channel, error) {
	db := utils.GetDBFromContext(c)
	rows, err := globals.QueryDb(db, `
		SELECT `+channelSelectColumns()+` FROM community_channel WHERE id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	ch, err := scanChannel(rows)
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func createChannel(c *gin.Context, req channelRequest, username string) (*Channel, error) {
	db := utils.GetDBFromContext(c)
	req.Visibility = normalizeVisibility(req.Visibility)
	if strings.TrimSpace(req.Name) == "" {
		req.Name = "new-channel"
	}
	res, err := globals.ExecDb(db, `
		INSERT INTO community_channel
		(name, description, topic, visibility, visible_groups, post_groups, members, posters, position, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, req.Name, req.Description, req.Topic, req.Visibility,
		utils.Marshal(req.VisibleGroups), utils.Marshal(req.PostGroups),
		utils.Marshal(req.Members), utils.Marshal(req.Posters),
		req.Position, username,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return getChannelById(c, int(id))
}

func updateChannel(c *gin.Context, id int, req channelRequest) (*Channel, error) {
	db := utils.GetDBFromContext(c)
	req.Visibility = normalizeVisibility(req.Visibility)
	if _, err := globals.ExecDb(db, `
		UPDATE community_channel SET
		  name = ?, description = ?, topic = ?, visibility = ?,
		  visible_groups = ?, post_groups = ?, members = ?, posters = ?,
		  position = ?
		WHERE id = ?
	`, req.Name, req.Description, req.Topic, req.Visibility,
		utils.Marshal(req.VisibleGroups), utils.Marshal(req.PostGroups),
		utils.Marshal(req.Members), utils.Marshal(req.Posters),
		req.Position, id,
	); err != nil {
		return nil, err
	}
	return getChannelById(c, id)
}

func deleteChannel(c *gin.Context, id int) error {
	db := utils.GetDBFromContext(c)
	if _, err := globals.ExecDb(db, `DELETE FROM community_message WHERE channel_id = ?`, id); err != nil {
		return err
	}
	_, err := globals.ExecDb(db, `DELETE FROM community_channel WHERE id = ?`, id)
	return err
}

// listMessages returns messages for a channel in ascending order.
// If before > 0, only messages with id < before are returned (backward pagination).
func listMessages(c *gin.Context, channelId, before, limit int) ([]Message, error) {
	db := utils.GetDBFromContext(c)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rows *sql.Rows
	var err error
	if before > 0 {
		rows, err = globals.QueryDb(db, `
			SELECT id, channel_id, sender_id, sender_username, content, created_at
			FROM community_message
			WHERE channel_id = ? AND id < ?
			ORDER BY id DESC
			LIMIT ?
		`, channelId, before, limit)
	} else {
		rows, err = globals.QueryDb(db, `
			SELECT id, channel_id, sender_id, sender_username, content, created_at
			FROM community_message
			WHERE channel_id = ?
			ORDER BY id DESC
			LIMIT ?
		`, channelId, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var desc []Message
	for rows.Next() {
		var msg Message
		var createdAt []uint8
		if err := rows.Scan(&msg.Id, &msg.ChannelId, &msg.SenderId, &msg.SenderUsername, &msg.Content, &createdAt); err != nil {
			return nil, err
		}
		msg.CreatedAt = formatTime(createdAt)
		desc = append(desc, msg)
	}

	// reverse to ascending order
	asc := make([]Message, 0, len(desc))
	for i := len(desc) - 1; i >= 0; i-- {
		asc = append(asc, desc[i])
	}
	return asc, nil
}

// sendMessage persists a message and returns it.
func sendMessage(c *gin.Context, channelId int, user *auth.User, content string) (*Message, error) {
	db := utils.GetDBFromContext(c)
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, nil
	}
	uid := user.GetID(db)
	res, err := globals.ExecDb(db, `
		INSERT INTO community_message (channel_id, sender_id, sender_username, content)
		VALUES (?, ?, ?, ?)
	`, channelId, uid, user.Username, content)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	var createdAt []uint8
	if err := globals.QueryRowDb(db, `
		SELECT created_at FROM community_message WHERE id = ?
	`, id).Scan(&createdAt); err != nil {
		return nil, err
	}
	return &Message{
		Id:             id,
		ChannelId:      channelId,
		SenderId:       uid,
		SenderUsername: user.Username,
		Content:        content,
		CreatedAt:      formatTime(createdAt),
	}, nil
}

// canView reports whether the user is allowed to see the channel.
func canView(db *sql.DB, ch *Channel, user *auth.User) bool {
	if user != nil && user.IsAdmin(db) {
		return true
	}
	if ch.Visibility == VisibilityPrivate {
		return user != nil && utils.Contains(user.Username, ch.Members)
	}
	// public channel
	if len(ch.VisibleGroups) == 0 {
		return user != nil
	}
	return user != nil && auth.HitGroups(db, user, ch.VisibleGroups)
}

// canPost reports whether the user is allowed to post in the channel.
func canPost(db *sql.DB, ch *Channel, user *auth.User) bool {
	if user == nil || !canView(db, ch, user) {
		return false
	}
	if user.IsAdmin(db) {
		return true
	}
	if utils.Contains(user.Username, ch.Posters) {
		return true
	}
	if len(ch.PostGroups) == 0 {
		return true
	}
	return auth.HitGroups(db, user, ch.PostGroups)
}
