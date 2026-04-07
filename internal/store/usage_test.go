package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsageTestData(t *testing.T) (*UsageStore, string, string) {
	t.Helper()
	database := setupTestDB(t)
	ps := NewProviderStore(database)
	ks := NewAPIKeyStore(database)
	us := NewUsageStore(database)

	provider, err := ps.Create("test-provider", "https://test.com", "sk-test")
	require.NoError(t, err)

	key, _, err := ks.Create("test-key", provider.ID, 0)
	require.NoError(t, err)

	return us, key.ID, provider.ID
}

func TestUsageStore_Record(t *testing.T) {
	us, keyID, providerID := setupUsageTestData(t)

	err := us.Record(keyID, providerID, "gpt-4", 100, 50, 150)
	assert.NoError(t, err)
}

func TestUsageStore_Query_Empty(t *testing.T) {
	us, _, _ := setupUsageTestData(t)

	result, err := us.Query(UsageQuery{})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Records)
}

func TestUsageStore_Query_WithRecords(t *testing.T) {
	us, keyID, providerID := setupUsageTestData(t)

	// Insert multiple records.
	for i := 0; i < 5; i++ {
		err := us.Record(keyID, providerID, "gpt-4", 100, 50, 150)
		require.NoError(t, err)
	}

	result, err := us.Query(UsageQuery{})
	require.NoError(t, err)
	assert.Equal(t, 5, result.Total)
	assert.Len(t, result.Records, 5)
}

func TestUsageStore_Query_FilterByAPIKeyID(t *testing.T) {
	database := setupTestDB(t)
	ps := NewProviderStore(database)
	ks := NewAPIKeyStore(database)
	us := NewUsageStore(database)

	provider, err := ps.Create("test-provider", "https://test.com", "sk-test")
	require.NoError(t, err)

	key1, _, err := ks.Create("key1", provider.ID, 0)
	require.NoError(t, err)
	key2, _, err := ks.Create("key2", provider.ID, 0)
	require.NoError(t, err)

	err = us.Record(key1.ID, provider.ID, "gpt-4", 100, 50, 150)
	require.NoError(t, err)
	err = us.Record(key2.ID, provider.ID, "gpt-4", 200, 100, 300)
	require.NoError(t, err)

	result, err := us.Query(UsageQuery{APIKeyID: key1.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, key1.ID, result.Records[0].APIKeyID)
}

func TestUsageStore_Query_FilterByProviderID(t *testing.T) {
	database := setupTestDB(t)
	ps := NewProviderStore(database)
	ks := NewAPIKeyStore(database)
	us := NewUsageStore(database)

	p1, err := ps.Create("provider1", "https://p1.com", "k1")
	require.NoError(t, err)
	p2, err := ps.Create("provider2", "https://p2.com", "k2")
	require.NoError(t, err)

	key1, _, err := ks.Create("key1", p1.ID, 0)
	require.NoError(t, err)
	key2, _, err := ks.Create("key2", p2.ID, 0)
	require.NoError(t, err)

	err = us.Record(key1.ID, p1.ID, "gpt-4", 100, 50, 150)
	require.NoError(t, err)
	err = us.Record(key2.ID, p2.ID, "gpt-4", 200, 100, 300)
	require.NoError(t, err)

	result, err := us.Query(UsageQuery{ProviderID: p1.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
}

func TestUsageStore_Query_FilterByTimeRange(t *testing.T) {
	us, keyID, providerID := setupUsageTestData(t)

	err := us.Record(keyID, providerID, "gpt-4", 100, 50, 150)
	require.NoError(t, err)

	// Query with future start time — should find nothing.
	future := time.Now().Add(1 * time.Hour).UTC()
	result, err := us.Query(UsageQuery{Start: &future})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)

	// Query with end time in the future — should find the record.
	futureEnd := time.Now().Add(1 * time.Hour).UTC()
	pastStart := time.Now().Add(-1 * time.Hour).UTC()
	result, err = us.Query(UsageQuery{Start: &pastStart, End: &futureEnd})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.Total, 1)

	// Query with end time in the past — should find nothing.
	pastEnd := time.Now().Add(-1 * time.Hour).UTC()
	result, err = us.Query(UsageQuery{End: &pastEnd})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
}

func TestUsageStore_Query_Pagination(t *testing.T) {
	us, keyID, providerID := setupUsageTestData(t)

	for i := 0; i < 10; i++ {
		err := us.Record(keyID, providerID, "gpt-4", 100, 50, 150)
		require.NoError(t, err)
	}

	// Page 1.
	result, err := us.Query(UsageQuery{Limit: 3, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 10, result.Total)
	assert.Len(t, result.Records, 3)

	// Page 2.
	result, err = us.Query(UsageQuery{Limit: 3, Offset: 3})
	require.NoError(t, err)
	assert.Equal(t, 10, result.Total)
	assert.Len(t, result.Records, 3)
}

func TestUsageStore_Query_DefaultLimit(t *testing.T) {
	us, _, _ := setupUsageTestData(t)

	// Limit <= 0 should default to 100.
	result, err := us.Query(UsageQuery{Limit: 0})
	require.NoError(t, err)
	assert.NotNil(t, result)
}
