package db

import "database/sql"

// migrate runs all schema migrations in order. Idempotent.
func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS providers (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL UNIQUE,
			base_url   TEXT NOT NULL,
			api_key    TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT (datetime('now')),
			updated_at DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE TABLE IF NOT EXISTS api_keys (
			id             TEXT PRIMARY KEY,
			name           TEXT NOT NULL,
			key_hash       TEXT NOT NULL UNIQUE,
			provider_id    TEXT NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
			rate_limit_rpm INTEGER NOT NULL DEFAULT 0,
			created_at     DATETIME NOT NULL DEFAULT (datetime('now')),
			revoked_at     DATETIME
		)`,

		`CREATE TABLE IF NOT EXISTS usage (
			id                TEXT PRIMARY KEY,
			api_key_id        TEXT NOT NULL REFERENCES api_keys(id),
			provider_id       TEXT NOT NULL REFERENCES providers(id),
			model             TEXT NOT NULL,
			prompt_tokens     INTEGER NOT NULL DEFAULT 0,
			completion_tokens INTEGER NOT NULL DEFAULT 0,
			total_tokens      INTEGER NOT NULL DEFAULT 0,
			created_at        DATETIME NOT NULL DEFAULT (datetime('now'))
		)`,

		`CREATE INDEX IF NOT EXISTS idx_usage_api_key_id ON usage(api_key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_provider_id ON usage(provider_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_created_at ON usage(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash)`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
