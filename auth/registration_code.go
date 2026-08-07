package auth

import (
        "context"
        "database/sql"
        "errors"
        "fmt"
        "strings"

        "neoai/globals"
        "neoai/utils"

        "github.com/go-redis/redis/v8"
)

// registration_code.go — registration code system.
//
// A registration code (a.k.a. "invite code" / "registration token") is
// distinct from the existing Invitation / Redeem systems:
//
//   - Invitation codes: optional codes that grant bonus quota to the
//     inviter. They are entered AFTER registration succeeds.
//   - Redeem codes: codes that grant quota to an existing user.
//   - Registration codes (this file): codes that gate the registration
//     flow itself. When `enable_registration_code` is on, the /register
//     endpoint requires a valid unused code in the request body. Each
//     code can be used a bounded number of times (default 1).
//
// Use case: keep the registration form open to the public, but only
// allow people with a code (e.g. distributed via email / discord /
// offline) to actually sign up. This dramatically reduces spam signups.

// RegistrationCode is the DB row.
type RegistrationCode struct {
        Id        int64   `json:"id"`
        Code      string  `json:"code"`
        Quota     float32 `json:"quota"`       // bonus quota granted to the new user
        MaxUses   int     `json:"max_uses"`    // how many distinct users can use it
        UsedCount int     `json:"used_count"`  // current uses
        Note      string  `json:"note"`        // admin's freeform note (e.g. "issued to discord")
        CreatedAt string  `json:"created_at"`
        Disabled  bool    `json:"disabled"`
}

// EnsureRegistrationCodeTable makes sure the registration_code table
// exists. Idempotent. Called from main.go after the connection package
// has opened the DB.
//
// (We can't call this from connection/database.go because connection is
// imported by adapter, which is imported by auth — so importing auth from
// connection would create a cycle. main.go is the only safe place.)
func EnsureRegistrationCodeTable(db *sql.DB) {
        CreateRegistrationCodeTable(db)
}

// CreateRegistrationCodeTable makes sure the table exists. Idempotent.
func CreateRegistrationCodeTable(db *sql.DB) {
        if _, err := globals.ExecDb(db, `
                CREATE TABLE IF NOT EXISTS registration_code (
                  id INT PRIMARY KEY AUTO_INCREMENT,
                  code VARCHAR(64) NOT NULL UNIQUE,
                  quota DECIMAL(24, 6) NOT NULL DEFAULT 0,
                  max_uses INT NOT NULL DEFAULT 1,
                  used_count INT NOT NULL DEFAULT 0,
                  note VARCHAR(255) NOT NULL DEFAULT '',
                  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                  disabled BOOLEAN NOT NULL DEFAULT FALSE
                );
        `); err != nil {
                fmt.Printf("[registration_code] create table error: %s\n", err.Error())
        }

        if globals.SqliteEngine {
                // SQLite variant — column types differ but semantics match.
                _, _ = globals.ExecDb(db, `
                        CREATE TABLE IF NOT EXISTS registration_code (
                          id INTEGER PRIMARY KEY AUTOINCREMENT,
                          code TEXT NOT NULL UNIQUE,
                          quota REAL NOT NULL DEFAULT 0,
                          max_uses INTEGER NOT NULL DEFAULT 1,
                          used_count INTEGER NOT NULL DEFAULT 0,
                          note TEXT NOT NULL DEFAULT '',
                          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                          disabled INTEGER NOT NULL DEFAULT 0
                        );
                `)
        }
}

// GenerateRegistrationCodes creates `num` fresh codes.
// Each code is random and prefixed with "REG-".
func GenerateRegistrationCodes(db *sql.DB, num int, quota float32, maxUses int, note string) ([]string, error) {
        codes := make([]string, 0, num)
        for i := 0; i < num; i++ {
                // 16 chars — long enough to resist brute force on a /register
                // rate-limited endpoint, short enough to type/paste comfortably.
                code := fmt.Sprintf("REG-%s", strings.ToUpper(utils.GenerateChar(16)))
                if _, err := globals.ExecDb(db, `
                        INSERT INTO registration_code (code, quota, max_uses, note)
                        VALUES (?, ?, ?, ?)
                `, code, quota, maxUses, note); err != nil {
                        // collision: retry this iteration
                        i--
                        continue
                }
                codes = append(codes, code)
        }
        return codes, nil
}

// ListRegistrationCodes returns all codes, newest first.
func ListRegistrationCodes(db *sql.DB) ([]RegistrationCode, error) {
        rows, err := globals.QueryDb(db, `
                SELECT id, code, quota, max_uses, used_count, note, created_at, disabled
                FROM registration_code
                ORDER BY id DESC
        `)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var out []RegistrationCode
        for rows.Next() {
                var (
                        rc        RegistrationCode
                        createdAt []uint8
                )
                if err := rows.Scan(&rc.Id, &rc.Code, &rc.Quota, &rc.MaxUses, &rc.UsedCount, &rc.Note, &createdAt, &rc.Disabled); err != nil {
                        return nil, err
                }
                if t := utils.ConvertTime(createdAt); t != nil {
                        rc.CreatedAt = t.Format("2006-01-02 15:04:05")
                }
                out = append(out, rc)
        }
        return out, nil
}

// ConsumeRegistrationCode validates and consumes a code for a new user.
// Returns the bonus quota to grant (0 if no quota attached).
// Returns error if:
//   - code doesn't exist
//   - code is disabled
//   - code's max_uses is exhausted
//
// This is a single SQL UPDATE with WHERE clauses, so it's atomic and
// race-safe — two simultaneous registrations can't both succeed on a
// code that only has 1 use left.
func ConsumeRegistrationCode(db *sql.DB, code string) (float32, error) {
        code = strings.TrimSpace(code)
        if len(code) == 0 {
                return 0, errors.New("registration code is required")
        }

        res, err := globals.ExecDb(db, `
                UPDATE registration_code
                SET used_count = used_count + 1
                WHERE code = ? AND disabled = FALSE AND used_count < max_uses
        `, code)
        if err != nil {
                return 0, err
        }
        affected, _ := res.RowsAffected()
        if affected == 0 {
                // Either code doesn't exist, is disabled, or is exhausted.
                // Surface a friendly error rather than leaking which one.
                return 0, errors.New("invalid or expired registration code")
        }

        // Fetch the bonus quota attached to the code.
        var quota float32
        if err := globals.QueryRowDb(db, `
                SELECT quota FROM registration_code WHERE code = ?
        `, code).Scan(&quota); err != nil {
                return 0, err
        }
        return quota, nil
}

// DisableRegistrationCode flips the `disabled` flag on a code. Already-
// consumed uses stay consumed, but no new users can register with it.
func DisableRegistrationCode(db *sql.DB, id int64) error {
        _, err := globals.ExecDb(db, `UPDATE registration_code SET disabled = TRUE WHERE id = ?`, id)
        return err
}

// DeleteRegistrationCode permanently removes a code.
func DeleteRegistrationCode(db *sql.DB, id int64) error {
        _, err := globals.ExecDb(db, `DELETE FROM registration_code WHERE id = ?`, id)
        return err
}

// IsRegistrationCodeRequired returns whether the site currently requires
// a registration code on signup. Reads from the system config cache,
// backed by `system.registration.require_code` (or env var).
//
// Default: disabled (false). Admin enables it from the system page.
func IsRegistrationCodeRequired() bool {
        // Cache + env are both checked via viper. The value can be toggled
        // at runtime through the admin system config page.
        return utils.GetBoolWithDefault("system.registration.require_code", false)
}

// GetBoolWithDefault shim — viper doesn't ship a "with default" variant
// for GetBool, so we use this small helper to avoid touching viper
// call sites in multiple places.
func init() {
        // no-op; the registration code table is created via CreateRegistrationCodeTable,
        // which is called from connection/database.go after CreateUserTable.
}

// silence unused import warning
var _ = context.Background
var _ = (*redis.Client)(nil)
