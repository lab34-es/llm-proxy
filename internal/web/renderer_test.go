package web

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	r := NewRenderer()
	assert.NotNil(t, r)
	assert.NotNil(t, r.templates)

	// Verify all expected templates are loaded.
	expectedPages := []string{"login", "providers", "keys", "usage", "playground"}
	for _, page := range expectedPages {
		_, ok := r.templates[page]
		assert.True(t, ok, "template %q should be loaded", page)
	}
}

func TestRenderer_RenderLogin(t *testing.T) {
	r := NewRenderer()
	var buf bytes.Buffer

	err := r.Render(&buf, "login", map[string]interface{}{"Error": ""}, nil)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "login") // should contain login-related content
}

func TestRenderer_RenderNotFound(t *testing.T) {
	r := NewRenderer()
	var buf bytes.Buffer

	err := r.Render(&buf, "nonexistent", nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestFuncMap_FormatTime(t *testing.T) {
	fn := funcMap["formatTime"].(func(time.Time) string)
	tm := time.Date(2025, 1, 15, 10, 30, 45, 0, time.UTC)
	assert.Equal(t, "2025-01-15 10:30:45 UTC", fn(tm))
}

func TestFuncMap_ShortID(t *testing.T) {
	fn := funcMap["shortID"].(func(string) string)
	assert.Equal(t, "abcdefgh", fn("abcdefgh-1234-5678"))
	assert.Equal(t, "short", fn("short"))
	assert.Equal(t, "12345678", fn("12345678"))
}

func TestFuncMap_IsRevoked(t *testing.T) {
	fn := funcMap["isRevoked"].(func(*time.Time) bool)
	assert.False(t, fn(nil))
	now := time.Now()
	assert.True(t, fn(&now))
}

func TestFuncMap_Add(t *testing.T) {
	fn := funcMap["add"].(func(int, int) int)
	assert.Equal(t, 5, fn(2, 3))
}

func TestFuncMap_Sub(t *testing.T) {
	fn := funcMap["sub"].(func(int, int) int)
	assert.Equal(t, 1, fn(3, 2))
}

func TestFuncMap_Seq(t *testing.T) {
	fn := funcMap["seq"].(func(int) []int)
	assert.Equal(t, []int{0, 1, 2}, fn(3))
	assert.Equal(t, []int{}, fn(0))
}
