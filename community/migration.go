package community

import (
	"database/sql"
	"fmt"
	"strings"

	"neoai/globals"
)

// Migrate creates the community_* tables if they don't exist, AND
// patches any old (pre-1.0) version of the schema by ALTER-ing in
// missing columns. This is necessary because `CREATE TABLE IF NOT
// EXISTS` is a no-op when the table already exists, even if it lacks
// newer columns — so users upgrading from an earlier build of NeoAI
// would hit "Unknown column 'send_permission' in 'field list'" errors.
//
// Idempotent — safe to call on every boot.
func Migrate(db *sql.DB) error {
	// ---- create tables if they don't exist yet ----
	if _, err := globals.ExecDb(db, `
		CREATE TABLE IF NOT EXISTS community_channel (
		  id INT PRIMARY KEY AUTO_INCREMENT,
		  name VARCHAR(64) NOT NULL,
		  topic VARCHAR(512) NOT NULL DEFAULT '',
		  visibility VARCHAR(32) NOT NULL DEFAULT 'members',
		  send_permission VARCHAR(32) NOT NULL DEFAULT 'everyone',
		  visible_roles TEXT,
		  visible_users TEXT,
		  senders TEXT,
		  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		);
	`); err != nil {
		// Some MySQL/MariaDB versions reject ON UPDATE CURRENT_TIMESTAMP
		// in combination with the second DATETIME column. Retry without
		// the ON UPDATE clause — the app code updates updated_at itself.
		if _, err2 := globals.ExecDb(db, `
			CREATE TABLE IF NOT EXISTS community_channel (
			  id INT PRIMARY KEY AUTO_INCREMENT,
			  name VARCHAR(64) NOT NULL,
			  topic VARCHAR(512) NOT NULL DEFAULT '',
			  visibility VARCHAR(32) NOT NULL DEFAULT 'members',
			  send_permission VARCHAR(32) NOT NULL DEFAULT 'everyone',
			  visible_roles TEXT,
			  visible_users TEXT,
			  senders TEXT,
			  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`); err2 != nil {
			return fmt.Errorf("create community_channel: %w (retry: %w)", err, err2)
		}
	}

	if _, err := globals.ExecDb(db, `
		CREATE TABLE IF NOT EXISTS community_message (
		  id INT PRIMARY KEY AUTO_INCREMENT,
		  channel_id INT NOT NULL,
		  user_id INT NOT NULL,
		  username VARCHAR(64) NOT NULL,
		  avatar VARCHAR(255) NOT NULL DEFAULT '',
		  content MEDIUMTEXT NOT NULL,
		  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		  edited_at DATETIME NULL,
		  INDEX idx_channel_time (channel_id, created_at),
		  FOREIGN KEY (channel_id) REFERENCES community_channel(id) ON DELETE CASCADE,
		  FOREIGN KEY (user_id) REFERENCES auth(id) ON DELETE CASCADE
		);
	`); err != nil {
		// Retry without FK / INDEX inline (older MySQL / some compat layers)
		if _, err2 := globals.ExecDb(db, `
			CREATE TABLE IF NOT EXISTS community_message (
			  id INT PRIMARY KEY AUTO_INCREMENT,
			  channel_id INT NOT NULL,
			  user_id INT NOT NULL,
			  username VARCHAR(64) NOT NULL,
			  avatar VARCHAR(255) NOT NULL DEFAULT '',
			  content MEDIUMTEXT NOT NULL,
			  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			  edited_at DATETIME NULL
			);
		`); err2 != nil {
			return fmt.Errorf("create community_message: %w (retry: %w)", err, err2)
		}
	}

	// ---- SQLite variants (no AUTO_INCREMENT / no ON UPDATE) ----
	if globals.SqliteEngine {
		if _, err := globals.ExecDb(db, `
			CREATE TABLE IF NOT EXISTS community_channel (
			  id INTEGER PRIMARY KEY AUTOINCREMENT,
			  name TEXT NOT NULL,
			  topic TEXT NOT NULL DEFAULT '',
			  visibility TEXT NOT NULL DEFAULT 'members',
			  send_permission TEXT NOT NULL DEFAULT 'everyone',
			  visible_roles TEXT,
			  visible_users TEXT,
			  senders TEXT,
			  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
		`); err != nil {
			return fmt.Errorf("create community_channel (sqlite): %w", err)
		}
		if _, err := globals.ExecDb(db, `
			CREATE TABLE IF NOT EXISTS community_message (
			  id INTEGER PRIMARY KEY AUTOINCREMENT,
			  channel_id INTEGER NOT NULL,
			  user_id INTEGER NOT NULL,
			  username TEXT NOT NULL,
			  avatar TEXT NOT NULL DEFAULT '',
			  content TEXT NOT NULL,
			  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			  edited_at DATETIME NULL
			);
		`); err != nil {
			return fmt.Errorf("create community_message (sqlite): %w", err)
		}
		_, _ = globals.ExecDb(db, `CREATE INDEX IF NOT EXISTS idx_community_msg_channel_time ON community_message(channel_id, created_at)`)
	}

	// ---- patch missing columns on existing tables (the real fix) ----
	if err := patchCommunityChannelColumns(db); err != nil {
		// Log but don't abort boot — admin can still read existing channels.
		fmt.Printf("[community] channel column patch warning: %s\n", err.Error())
	}
	if err := patchCommunityMessageColumns(db); err != nil {
		fmt.Printf("[community] message column patch warning: %s\n", err.Error())
	}

	return nil
}

// patchCommunityChannelColumns adds any columns that are missing from
// an older `community_channel` table. Each ALTER is wrapped in an
// existence check so we don't fire ALTER on a column that's already
// there (which would error on MySQL).
//
// We can't use information_schema on SQLite — for SQLite we parse
// `PRAGMA table_info(community_channel)` instead.
func patchCommunityChannelColumns(db *sql.DB) error {
	existing, err := existingColumns(db, "community_channel")
	if err != nil {
		return err
	}

	want := map[string]string{
		"name":            "VARCHAR(64) NOT NULL DEFAULT ''",
		"topic":           "VARCHAR(512) NOT NULL DEFAULT ''",
		"visibility":      "VARCHAR(32) NOT NULL DEFAULT 'members'",
		"send_permission": "VARCHAR(32) NOT NULL DEFAULT 'everyone'",
		"visible_roles":   "TEXT",
		"visible_users":   "TEXT",
		"senders":         "TEXT",
		"created_at":      "DATETIME DEFAULT CURRENT_TIMESTAMP",
		"updated_at":      "DATETIME DEFAULT CURRENT_TIMESTAMP",
	}

	for col, decl := range want {
		if _, ok := existing[col]; ok {
			continue
		}
		// Column is missing — ALTER it in. Use the same syntax for both
		// MySQL and SQLite; both accept `ALTER TABLE t ADD COLUMN c TYPE`.
		stmt := fmt.Sprintf("ALTER TABLE community_channel ADD COLUMN %s %s", col, decl)
		if _, err := globals.ExecDb(db, stmt); err != nil {
			// If this fails (e.g. unsupported syntax on this DB version),
			// surface the error so the admin knows which column failed.
			return fmt.Errorf("add column %s: %w", col, err)
		}
		fmt.Printf("[community] added missing column community_channel.%s\n", col)
	}
	return nil
}

// patchCommunityMessageColumns does the same for community_message.
func patchCommunityMessageColumns(db *sql.DB) error {
	existing, err := existingColumns(db, "community_message")
	if err != nil {
		return err
	}

	want := map[string]string{
		"channel_id": "INT NOT NULL DEFAULT 0",
		"user_id":    "INT NOT NULL DEFAULT 0",
		"username":   "VARCHAR(64) NOT NULL DEFAULT ''",
		"avatar":     "VARCHAR(255) NOT NULL DEFAULT ''",
		"content":    "MEDIUMTEXT NOT NULL",
		"created_at": "DATETIME DEFAULT CURRENT_TIMESTAMP",
		"edited_at":  "DATETIME NULL",
	}

	for col, decl := range want {
		if _, ok := existing[col]; ok {
			continue
		}
		stmt := fmt.Sprintf("ALTER TABLE community_message ADD COLUMN %s %s", col, decl)
		if _, err := globals.ExecDb(db, stmt); err != nil {
			return fmt.Errorf("add column %s: %w", col, err)
		}
		fmt.Printf("[community] added missing column community_message.%s\n", col)
	}
	return nil
}

// existingColumns returns the set of column names for a given table.
// Works on MySQL, MariaDB and SQLite — we try information_schema first
// (MySQL/MariaDB) and fall back to PRAGMA table_info (SQLite).
func existingColumns(db *sql.DB, table string) (map[string]struct{}, error) {
	out := map[string]struct{}{}

	// MySQL / MariaDB path
	rows, err := globals.QueryDb(db, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = DATABASE() AND table_name = ?
	`, table)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var col string
			if err := rows.Scan(&col); err != nil {
				return nil, err
			}
			out[strings.ToLower(col)] = struct{}{}
		}
		if len(out) > 0 {
			return out, nil
		}
		// fall through to SQLite path if information_schema returned nothing
	}

	// SQLite path
	pragmaRows, err := globals.QueryDb(db, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer pragmaRows.Close()
	for pragmaRows.Next() {
		var (
			cid       int
			name      string
			typ       string
			notnull   int
			dfltValue interface{}
			pk        int
		)
		if err := pragmaRows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		out[strings.ToLower(name)] = struct{}{}
	}
	return out, nil
}
