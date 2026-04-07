package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_JSONSerialization(t *testing.T) {
	p := Provider{
		ID:        "test-id",
		Name:      "openai",
		BaseURL:   "https://api.openai.com",
		APIKey:    "sk-secret",
		CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(p)
	require.NoError(t, err)

	// APIKey should NOT be in JSON (tagged json:"-").
	assert.NotContains(t, string(data), "sk-secret")
	assert.Contains(t, string(data), "openai")
	assert.Contains(t, string(data), "test-id")
}

func TestAPIKey_JSONSerialization(t *testing.T) {
	k := APIKey{
		ID:           "key-id",
		Name:         "my-key",
		KeyHash:      "somehash",
		ProviderID:   "provider-id",
		RateLimitRPM: 60,
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		RevokedAt:    nil,
	}

	data, err := json.Marshal(k)
	require.NoError(t, err)

	// KeyHash should NOT be in JSON (tagged json:"-").
	assert.NotContains(t, string(data), "somehash")
	assert.Contains(t, string(data), "my-key")
}

func TestAPIKey_RevokedAt_Omitempty(t *testing.T) {
	k := APIKey{
		ID:   "key-id",
		Name: "test",
	}

	data, err := json.Marshal(k)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "revoked_at")

	// With revoked_at set.
	now := time.Now()
	k.RevokedAt = &now
	data, err = json.Marshal(k)
	require.NoError(t, err)
	assert.Contains(t, string(data), "revoked_at")
}

func TestUsageRecord_JSONSerialization(t *testing.T) {
	r := UsageRecord{
		ID:               "usage-id",
		APIKeyID:         "key-id",
		ProviderID:       "provider-id",
		Model:            "gpt-4",
		PromptTokens:     100,
		CompletionTokens: 50,
		TotalTokens:      150,
		CreatedAt:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)
	assert.Contains(t, string(data), "gpt-4")
	assert.Contains(t, string(data), "150")
}
