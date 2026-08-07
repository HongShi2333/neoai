package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"neoai/globals"
	"neoai/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// profile_username.go — self-service username change.
//
// POST /profile/username
// Body: { "username": "<new name>" }
//
// Authenticated users can rename themselves. The new name must:
//   - be 2..24 chars (matches the existing register-form validator)
//   - not be already taken by another user
//   - not equal the current username (no-op is rejected to surface bugs)
//
// The user's auth cache (nio:user:<old> and nio:user:<new>) is cleared
// after the rename so the next JWT validation reflects the change.
//
// The JWT itself stays valid because it stores username+password-hash,
// and only the username column is updated. Clients that hold an old
// token will continue to authenticate as the new username after the
// cache invalidation propagates.

type profileUsernameForm struct {
	Username string `json:"username" binding:"required"`
}

func UpdateProfileUsernameAPI(c *gin.Context) {
	user := RequireAuth(c)
	if user == nil {
		return
	}

	var form profileUsernameForm
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(200, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	db := utils.GetDBFromContext(c)
	cache := utils.GetCacheFromContext(c)

	newName := strings.TrimSpace(form.Username)
	if err := validateProfileUsername(newName); err != nil {
		c.JSON(200, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	uid := user.GetID(db)
	var current string
	if err := globals.QueryRowDb(db, "SELECT username FROM auth WHERE id = ?", uid).Scan(&current); err != nil {
		c.JSON(200, gin.H{
			"status": false,
			"error":  "user not found",
		})
		return
	}

	if current == newName {
		c.JSON(200, gin.H{
			"status": false,
			"error":  "new username is the same as the current one",
		})
		return
	}

	// uniqueness check (excluding self)
	var count int
	if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE username = ? AND id <> ?", newName, uid).Scan(&count); err != nil {
		c.JSON(200, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}
	if count > 0 {
		c.JSON(200, gin.H{
			"status": false,
			"error":  fmt.Sprintf("username %q is already taken", newName),
		})
		return
	}

	if err := applyProfileUsername(db, cache, uid, current, newName); err != nil {
		c.JSON(200, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  true,
		"username": newName,
	})
}

func validateProfileUsername(name string) error {
	if len(name) < 2 || len(name) > 24 {
		return fmt.Errorf("username length must be between 2 and 24")
	}
	// Reject obviously bad characters. The same regex as the register form
	// is intentionally NOT enforced here so admins who imported users with
	// non-standard names can keep editing them.
	if strings.ContainsAny(name, "\r\n\t") {
		return fmt.Errorf("username contains invalid characters")
	}
	return nil
}

func applyProfileUsername(db *sql.DB, cache *redis.Client, id int64, oldName, newName string) error {
	if _, err := globals.ExecDb(db, `UPDATE auth SET username = ? WHERE id = ?`, newName, id); err != nil {
		return err
	}
	if cache != nil {
		ctx := context.Background()
		cache.Del(ctx, fmt.Sprintf("nio:user:%s", oldName))
		cache.Del(ctx, fmt.Sprintf("nio:user:%s", newName))
	}
	return nil
}
