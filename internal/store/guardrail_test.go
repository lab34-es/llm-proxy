package store

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuardrailStore_Create(t *testing.T) {
	database := setupTestDB(t)
	s := NewGuardrailStore(database)

	g, err := s.Create(`\bpassword\b`, "reject", "")
	require.NoError(t, err)
	assert.NotEmpty(t, g.ID)
	assert.Equal(t, `\bpassword\b`, g.Pattern)
	assert.Equal(t, "reject", g.Mode)
	assert.Equal(t, "", g.ReplaceBy)
	assert.False(t, g.CreatedAt.IsZero())
	assert.False(t, g.UpdatedAt.IsZero())
}

func TestGuardrailStore_Create_Replace(t *testing.T) {
	database := setupTestDB(t)
	s := NewGuardrailStore(database)

	g, err := s.Create(`secret`, "replace", "[REDACTED]")
	require.NoError(t, err)
	assert.Equal(t, "replace", g.Mode)
	assert.Equal(t, "[REDACTED]", g.ReplaceBy)
}

func TestGuardrailStore_List(t *testing.T) {
	database := setupTestDB(t)
	s := NewGuardrailStore(database)

	// Empty list.
	list, err := s.List()
	require.NoError(t, err)
	assert.Empty(t, list)

	// Add guardrails.
	_, err = s.Create(`pattern1`, "reject", "")
	require.NoError(t, err)
	_, err = s.Create(`pattern2`, "replace", "xxx")
	require.NoError(t, err)

	list, err = s.List()
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestGuardrailStore_GetByID(t *testing.T) {
	database := setupTestDB(t)
	s := NewGuardrailStore(database)

	created, err := s.Create(`test`, "reject", "")
	require.NoError(t, err)

	g, err := s.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, g)
	assert.Equal(t, created.ID, g.ID)
	assert.Equal(t, "test", g.Pattern)
}

func TestGuardrailStore_GetByID_NotFound(t *testing.T) {
	database := setupTestDB(t)
	s := NewGuardrailStore(database)

	g, err := s.GetByID("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, g)
}

func TestGuardrailStore_Delete(t *testing.T) {
	database := setupTestDB(t)
	s := NewGuardrailStore(database)

	created, err := s.Create(`todelete`, "reject", "")
	require.NoError(t, err)

	err = s.Delete(created.ID)
	assert.NoError(t, err)

	// Verify deleted.
	g, err := s.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, g)
}

func TestGuardrailStore_Delete_NotFound(t *testing.T) {
	database := setupTestDB(t)
	s := NewGuardrailStore(database)

	err := s.Delete("nonexistent")
	assert.Equal(t, sql.ErrNoRows, err)
}
