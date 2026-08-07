package admin

import (
        "neoai/channel"
        "neoai/globals"
        "neoai/utils"
        "context"
        "database/sql"
        "fmt"
        "math"
        "strings"
        "time"

        "github.com/go-redis/redis/v8"
)

// AuthLike is to solve the problem of import cycle
type AuthLike struct {
        ID int64 `json:"id"`
}

func (a *AuthLike) GetID(_ *sql.DB) int64 {
        return a.ID
}

func (a *AuthLike) HitID() int64 {
        return a.ID
}

func getUsersForm(db *sql.DB, page int64, search string) PaginationForm {
        // if search is empty, then search all users

        var users []interface{}
        var total int64

        if err := globals.QueryRowDb(db, `
                SELECT COUNT(*) FROM auth
                WHERE username LIKE ?
        `, "%"+search+"%").Scan(&total); err != nil {
                return PaginationForm{
                        Status:  false,
                        Message: err.Error(),
                }
        }

        rows, err := globals.QueryDb(db, `
                SELECT 
                    auth.id, auth.username, auth.email, auth.is_admin,
                    quota.quota, quota.used,
                    subscription.expired_at, subscription.total_month, subscription.enterprise, subscription.level,
                    auth.is_banned
                FROM auth
                LEFT JOIN quota ON quota.user_id = auth.id
                LEFT JOIN subscription ON subscription.user_id = auth.id
                WHERE auth.username LIKE ?
                ORDER BY auth.id LIMIT ? OFFSET ?
        `, "%"+search+"%", pagination, page*pagination)
        if err != nil {
                return PaginationForm{
                        Status:  false,
                        Message: err.Error(),
                }
        }

        for rows.Next() {
                var user UserData
                var (
                        email             sql.NullString
                        expired           []uint8
                        quota             sql.NullFloat64
                        usedQuota         sql.NullFloat64
                        totalMonth        sql.NullInt64
                        isEnterprise      sql.NullBool
                        subscriptionLevel sql.NullInt64
                        isBanned          sql.NullBool
                )
                if err := rows.Scan(&user.Id, &user.Username, &email, &user.IsAdmin, &quota, &usedQuota, &expired, &totalMonth, &isEnterprise, &subscriptionLevel, &isBanned); err != nil {
                        return PaginationForm{
                                Status:  false,
                                Message: err.Error(),
                        }
                }
                if email.Valid {
                        user.Email = email.String
                }
                if quota.Valid {
                        user.Quota = float32(quota.Float64)
                }
                if usedQuota.Valid {
                        user.UsedQuota = float32(usedQuota.Float64)
                }
                if totalMonth.Valid {
                        user.TotalMonth = totalMonth.Int64
                }
                if subscriptionLevel.Valid {
                        user.Level = int(subscriptionLevel.Int64)
                }
                stamp := utils.ConvertTime(expired)
                if stamp != nil {
                        user.IsSubscribed = stamp.After(time.Now())
                        user.ExpiredAt = stamp.Format("2006-01-02 15:04:05")
                }
                user.Enterprise = isEnterprise.Valid && isEnterprise.Bool
                user.IsBanned = isBanned.Valid && isBanned.Bool

                users = append(users, user)
        }

        return PaginationForm{
                Status: true,
                Total:  int(math.Ceil(float64(total) / float64(pagination))),
                Data:   users,
        }
}

// clearUserCache clears all cache keys starting with nio:user:
func clearUserCache(cache *redis.Client) error {
        ctx := context.Background()
        iter := cache.Scan(ctx, 0, "nio:user:*", 100).Iterator()
        for iter.Next(ctx) {
                if err := cache.Del(ctx, iter.Val()).Err(); err != nil {
                        return fmt.Errorf("failed to delete cache key %s: %v", iter.Val(), err)
                }
        }
        return iter.Err()
}

func passwordMigration(db *sql.DB, cache *redis.Client, id int64, password string) error {
        password = strings.TrimSpace(password)
        if len(password) < 6 || len(password) > 36 {
                return fmt.Errorf("password length must be between 6 and 36")
        }
        hash_passwd := utils.Sha2Encrypt(password)

        // Update password in database
        _, err := globals.ExecDb(db, `
                UPDATE auth SET password = ? WHERE id = ?
        `, hash_passwd, id)

        if err != nil {
                return err
        }

        // Clear all user related cache
        if err := clearUserCache(cache); err != nil {
                return fmt.Errorf("failed to clear user cache: %v", err)
        }

        return nil
}

func emailMigration(db *sql.DB, id int64, email string) error {
        _, err := globals.ExecDb(db, `
                UPDATE auth SET email = ? WHERE id = ?
        `, email, id)

        return err
}

// usernameMigration renames a user. Validates the new name (length 2..24,
// must be unique across the auth table) and clears the user's auth cache
// so the change is picked up by the next JWT validation.
//
// Used by:
//   - /admin/user/username   (admin renames any user)
//   - /profile/username      (user renames self)
func usernameMigration(db *sql.DB, cache *redis.Client, id int64, username string) error {
        username = strings.TrimSpace(username)
        if len(username) < 2 || len(username) > 24 {
                return fmt.Errorf("username length must be between 2 and 24")
        }

        // Look up the user's current username to (a) skip the no-op case
        // and (b) clear their old auth cache key after the rename.
        var current string
        if err := globals.QueryRowDb(db, "SELECT username FROM auth WHERE id = ?", id).Scan(&current); err != nil {
                return fmt.Errorf("user not found")
        }
        if current == username {
                // nothing to do
                return nil
        }

        // Uniqueness check — relies on the unique constraint on `username`,
        // but we want a friendlier error than a raw MySQL duplicate-key error.
        var count int
        if err := globals.QueryRowDb(db, "SELECT COUNT(*) FROM auth WHERE username = ? AND id <> ?", username, id).Scan(&count); err != nil {
                return err
        }
        if count > 0 {
                return fmt.Errorf("username %q is already taken", username)
        }

        if _, err := globals.ExecDb(db, `
                UPDATE auth SET username = ? WHERE id = ?
        `, username, id); err != nil {
                return err
        }

        // Clear auth cache for BOTH the old and new username so any cached
        // JWT validation result is invalidated and the next request hits the DB.
        if cache != nil {
                ctx := context.Background()
                cache.Del(ctx, fmt.Sprintf("nio:user:%s", current))
                cache.Del(ctx, fmt.Sprintf("nio:user:%s", username))
        }
        return nil
}

func setAdmin(db *sql.DB, id int64, isAdmin bool) error {
        _, err := globals.ExecDb(db, `
                UPDATE auth SET is_admin = ? WHERE id = ?
        `, isAdmin, id)

        return err
}

func banUser(db *sql.DB, id int64, isBanned bool) error {
        _, err := globals.ExecDb(db, `
                UPDATE auth SET is_banned = ? WHERE id = ?
        `, isBanned, id)

        return err
}

func quotaMigration(db *sql.DB, id int64, quota float32, override bool) error {
        // if quota is negative, then decrease quota
        // if quota is positive, then increase quota
        //
        // NeoAI: use a portable UPSERT pattern that works on both MySQL
        // and SQLite. SQLite doesn't support `ON DUPLICATE KEY UPDATE`,
        // and MySQL < 8.0 doesn't support `INSERT OR REPLACE`, so we
        // detect the engine and branch accordingly.
        if globals.SqliteEngine {
                // SQLite path
                if override {
                        _, err := globals.ExecDb(db, `
                                INSERT INTO quota (user_id, quota, used) VALUES (?, ?, 0)
                                ON CONFLICT(user_id) DO UPDATE SET quota = ?
                        `, id, quota, quota)
                        return err
                }
                _, err := globals.ExecDb(db, `
                        INSERT INTO quota (user_id, quota, used) VALUES (?, ?, 0)
                        ON CONFLICT(user_id) DO UPDATE SET quota = quota + ?
                `, id, quota, quota)
                return err
        }

        if override {
                _, err := globals.ExecDb(db, `
                        INSERT INTO quota (user_id, quota, used) VALUES (?, ?, 0)
                        ON DUPLICATE KEY UPDATE quota = ?
                `, id, quota, quota)
                return err
        }

        _, err := globals.ExecDb(db, `
                INSERT INTO quota (user_id, quota, used) VALUES (?, ?, 0)
                ON DUPLICATE KEY UPDATE quota = quota + ?
        `, id, quota, quota)
        return err
}

func subscriptionMigration(db *sql.DB, id int64, expired string) error {
        if globals.SqliteEngine {
                _, err := globals.ExecDb(db, `
                        INSERT INTO subscription (user_id, expired_at) VALUES (?, ?)
                        ON CONFLICT(user_id) DO UPDATE SET expired_at = ?
                `, id, expired, expired)
                return err
        }
        _, err := globals.ExecDb(db, `
                INSERT INTO subscription (user_id, expired_at) VALUES (?, ?)
                ON DUPLICATE KEY UPDATE expired_at = ?
        `, id, expired, expired)
        return err
}

func subscriptionLevelMigration(db *sql.DB, id int64, level int64) error {
        if level < 0 || level > 3 {
                return fmt.Errorf("invalid subscription level")
        }

        if globals.SqliteEngine {
                _, err := globals.ExecDb(db, `
                        INSERT INTO subscription (user_id, level) VALUES (?, ?)
                        ON CONFLICT(user_id) DO UPDATE SET level = ?
                `, id, level, level)
                return err
        }

        _, err := globals.ExecDb(db, `
                INSERT INTO subscription (user_id, level) VALUES (?, ?)
                ON DUPLICATE KEY UPDATE level = ?
        `, id, level, level)
        return err
}

func releaseUsage(db *sql.DB, cache *redis.Client, id int64) error {
        var level sql.NullInt64
        if err := globals.QueryRowDb(db, `
                SELECT level FROM subscription WHERE user_id = ?
        `, id).Scan(&level); err != nil {
                return err
        }

        if !level.Valid || level.Int64 == 0 {
                return fmt.Errorf("user is not subscribed")
        }

        u := &AuthLike{ID: id}

        plan := channel.PlanInstance.GetPlan(int(level.Int64))
        if !plan.ReleaseAll(u, cache) {
                return fmt.Errorf("cannot release usage")
        }

        return nil
}

func UpdateRootPassword(db *sql.DB, cache *redis.Client, password string) error {
        password = strings.TrimSpace(password)
        if len(password) < 6 || len(password) > 36 {
                return fmt.Errorf("password length must be between 6 and 36")
        }

        if _, err := globals.ExecDb(db, `
                UPDATE auth SET password = ? WHERE username = 'root'
        `, utils.Sha2Encrypt(password)); err != nil {
                return err
        }

        // Clear all user related cache
        if err := clearUserCache(cache); err != nil {
                return fmt.Errorf("failed to clear user cache: %v", err)
        }

        return nil
}
