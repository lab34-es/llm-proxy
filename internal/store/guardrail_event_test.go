package store

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupGuardrailEventTest(t *testing.T) (*GuardrailEventStore, *GuardrailStore, *ProviderStore, *APIKeyStore) {
	t.Helper()
	database := setupTestDB(t)
	gs := NewGuardrailStore(database)
	ges := NewGuardrailEventStore(database)
	ps := NewProviderStore(database)
	ks := NewAPIKeyStore(database)
	return ges, gs, ps, ks
}

func TestGuardrailEventStore_Record(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g, err := gs.Create(`\bpassword\b`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("test-key", p.ID, 0)
	require.NoError(t, err)

	ev, err := ges.Record(g.ID, k.ID, g.Pattern, g.Mode, "my password is secret")
	require.NoError(t, err)
	assert.NotEmpty(t, ev.ID)
	assert.Equal(t, g.ID, ev.GuardrailID)
	assert.Equal(t, k.ID, ev.APIKeyID)
	assert.Equal(t, g.Pattern, ev.Pattern)
	assert.Equal(t, "reject", ev.Mode)
	assert.Equal(t, "my password is secret", ev.InputText)
	assert.False(t, ev.CreatedAt.IsZero())
}

func TestGuardrailEventStore_List_Empty(t *testing.T) {
	ges, _, _, _ := setupGuardrailEventTest(t)

	result, err := ges.List(GuardrailEventQuery{})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Records)
}

func TestGuardrailEventStore_List_WithRecords(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g, err := gs.Create(`test`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("k", p.ID, 0)
	require.NoError(t, err)

	_, err = ges.Record(g.ID, k.ID, g.Pattern, g.Mode, "input1")
	require.NoError(t, err)
	_, err = ges.Record(g.ID, k.ID, g.Pattern, g.Mode, "input2")
	require.NoError(t, err)

	result, err := ges.List(GuardrailEventQuery{})
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Records, 2)
}

func TestGuardrailEventStore_List_FilterByGuardrailID(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g1, err := gs.Create(`pattern1`, "reject", "")
	require.NoError(t, err)
	g2, err := gs.Create(`pattern2`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("k", p.ID, 0)
	require.NoError(t, err)

	_, err = ges.Record(g1.ID, k.ID, g1.Pattern, g1.Mode, "input1")
	require.NoError(t, err)
	_, err = ges.Record(g2.ID, k.ID, g2.Pattern, g2.Mode, "input2")
	require.NoError(t, err)

	result, err := ges.List(GuardrailEventQuery{GuardrailID: g1.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, g1.ID, result.Records[0].GuardrailID)
}

func TestGuardrailEventStore_List_FilterByAPIKeyID(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g, err := gs.Create(`test`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k1, _, err := ks.Create("k1", p.ID, 0)
	require.NoError(t, err)
	k2, _, err := ks.Create("k2", p.ID, 0)
	require.NoError(t, err)

	_, err = ges.Record(g.ID, k1.ID, g.Pattern, g.Mode, "input1")
	require.NoError(t, err)
	_, err = ges.Record(g.ID, k2.ID, g.Pattern, g.Mode, "input2")
	require.NoError(t, err)

	result, err := ges.List(GuardrailEventQuery{APIKeyID: k1.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, k1.ID, result.Records[0].APIKeyID)
}

func TestGuardrailEventStore_List_Pagination(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g, err := gs.Create(`test`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("k", p.ID, 0)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		_, err = ges.Record(g.ID, k.ID, g.Pattern, g.Mode, "input")
		require.NoError(t, err)
	}

	result, err := ges.List(GuardrailEventQuery{Limit: 2, Offset: 0})
	require.NoError(t, err)
	assert.Equal(t, 5, result.Total)
	assert.Len(t, result.Records, 2)

	result, err = ges.List(GuardrailEventQuery{Limit: 2, Offset: 4})
	require.NoError(t, err)
	assert.Equal(t, 5, result.Total)
	assert.Len(t, result.Records, 1)
}

func TestGuardrailEventStore_GetByID(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g, err := gs.Create(`test`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("k", p.ID, 0)
	require.NoError(t, err)

	created, err := ges.Record(g.ID, k.ID, g.Pattern, g.Mode, "some input")
	require.NoError(t, err)

	ev, err := ges.GetByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, ev)
	assert.Equal(t, created.ID, ev.ID)
	assert.Equal(t, "some input", ev.InputText)
}

func TestGuardrailEventStore_GetByID_NotFound(t *testing.T) {
	ges, _, _, _ := setupGuardrailEventTest(t)

	ev, err := ges.GetByID("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, ev)
}

func TestGuardrailEventStore_Delete(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g, err := gs.Create(`test`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("k", p.ID, 0)
	require.NoError(t, err)

	created, err := ges.Record(g.ID, k.ID, g.Pattern, g.Mode, "input")
	require.NoError(t, err)

	err = ges.Delete(created.ID)
	assert.NoError(t, err)

	ev, err := ges.GetByID(created.ID)
	require.NoError(t, err)
	assert.Nil(t, ev)
}

func TestGuardrailEventStore_Delete_NotFound(t *testing.T) {
	ges, _, _, _ := setupGuardrailEventTest(t)

	err := ges.Delete("nonexistent")
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestGuardrailEventStore_CascadeDelete(t *testing.T) {
	ges, gs, ps, ks := setupGuardrailEventTest(t)

	g, err := gs.Create(`test`, "reject", "")
	require.NoError(t, err)
	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("k", p.ID, 0)
	require.NoError(t, err)

	_, err = ges.Record(g.ID, k.ID, g.Pattern, g.Mode, "input")
	require.NoError(t, err)

	// Deleting the guardrail should cascade-delete events.
	err = gs.Delete(g.ID)
	require.NoError(t, err)

	result, err := ges.List(GuardrailEventQuery{})
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
}
