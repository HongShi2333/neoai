package community

import (
	"database/sql"
	"fmt"

	"neoai/globals"
)

// Migrate creates the community_* tables if they don't exist.
// Idempotent — safe to call on every boot.
func Migrate(db *sql.DB) error {
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
		return fmt.Errorf("create community_channel: %w", err)
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
		return fmt.Errorf("create community_message: %w", err)
	}

	if globals.SqliteEngine {
		// SQLite doesn't support ON UPDATE CURRENT_TIMESTAMP — handled in app code.
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
			CREATE INDEX IF NOT EXISTS idx_community_msg_channel_time ON community_message(channel_id, created_at);
		`); err != nil {
			return fmt.Errorf("create community_message (sqlite): %w", err)
		}
	}

	return nil
}
