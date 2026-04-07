package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen_InMemory(t *testing.T) {
	db, err := Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Verify WAL mode (in-memory uses "memory" instead of "wal", but the pragma still succeeds).
	var mode string
	err = db.QueryRow("PRAGMA journal_mode").Scan(&mode)
	require.NoError(t, err)

	// Verify foreign keys enabled.
	var fk int
	err = db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	require.NoError(t, err)
	assert.Equal(t, 1, fk)

	// Verify tables were created.
	tables := []string{"providers", "api_keys", "usage"}
	for _, table := range tables {
		var name string
		err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		require.NoError(t, err, "table %s should exist", table)
		assert.Equal(t, table, name)
	}
}

func TestOpen_FileDB(t *testing.T) {
	dir := t.TempDir()
	dsn := filepath.Join(dir, "test.db")

	db, err := Open(dsn)
	require.NoError(t, err)
	defer db.Close()

	// Verify file was created.
	_, err = os.Stat(dsn)
	assert.NoError(t, err)
}

func TestOpen_InvalidDSN(t *testing.T) {
	// Use a path that can't be created.
	_, err := Open("/nonexistent/path/that/cant/exist/db.sqlite")
	assert.Error(t, err)
}

func TestMigrate_Idempotent(t *testing.T) {
	db, err := Open(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Run migrate again — should be idempotent.
	err = migrate(db)
	assert.NoError(t, err)
}
