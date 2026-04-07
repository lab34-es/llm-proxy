package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/lab34/llm-proxy/internal/db"
	"github.com/lab34/llm-proxy/internal/store"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDashboard(t *testing.T) (*DashboardHandler, *store.ProviderStore, *store.APIKeyStore, *store.UsageStore, *echo.Echo) {
	t.Helper()
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })

	ps := store.NewProviderStore(database)
	ks := store.NewAPIKeyStore(database)
	us := store.NewUsageStore(database)
	gs := store.NewGuardrailStore(database)
	ges := store.NewGuardrailEventStore(database)

	h := NewDashboardHandler(ps, ks, us, gs, ges, "admin-token", "secret")

	e := echo.New()
	e.Renderer = NewRenderer()

	return h, ps, ks, us, e
}

func TestIndex_Redirect(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Index(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/dashboard/providers", rec.Header().Get("Location"))
}

func TestLoginPage_NotLoggedIn(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestLoginPage_AlreadyLoggedIn(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	signed := signToken("admin-token", "secret")
	req := httptest.NewRequest(http.MethodGet, "/dashboard/login", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: signed})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/dashboard/providers", rec.Header().Get("Location"))
}

func TestLoginSubmit_Success(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{"token": {"admin-token"}}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginSubmit(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/dashboard/providers", rec.Header().Get("Location"))

	// Should have set a cookie.
	cookies := rec.Result().Cookies()
	assert.NotEmpty(t, cookies)
}

func TestLoginSubmit_InvalidToken(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{"token": {"wrong-token"}}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.LoginSubmit(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	// Should re-render login page with error.
}

func TestLogout(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Logout(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/dashboard/login", rec.Header().Get("Location"))
}

func TestProvidersPage(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/providers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.ProvidersPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateProvider_Dashboard_Success(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{
		"name":     {"openai"},
		"base_url": {"https://api.openai.com"},
		"api_key":  {"sk-test"},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/providers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateProvider(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
}

func TestCreateProvider_Dashboard_MissingFields(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{"name": {"openai"}}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/providers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateProvider(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // re-renders with error
}

func TestDeleteProvider_Dashboard(t *testing.T) {
	h, ps, _, _, e := setupDashboard(t)

	p, err := ps.Create("todel", "https://del.com", "key")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/providers/"+p.ID+"/delete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(p.ID)

	err = h.DeleteProvider(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
}

func TestDeleteProvider_Dashboard_NotFound(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/providers/nonexistent/delete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.DeleteProvider(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code) // redirect with flash error
}

func TestKeysPage(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/keys", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.KeysPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateKey_Dashboard_Success(t *testing.T) {
	h, ps, _, _, e := setupDashboard(t)

	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)

	form := url.Values{
		"name":           {"my-key"},
		"provider_id":    {p.ID},
		"rate_limit_rpm": {"60"},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/keys", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.CreateKey(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // renders with new key displayed
}

func TestCreateKey_Dashboard_MissingFields(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{"name": {"my-key"}}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/keys", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateKey(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // re-renders with error
}

func TestCreateKey_Dashboard_ProviderNotFound(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{
		"name":        {"my-key"},
		"provider_id": {"nonexistent"},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/keys", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateKey(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // re-renders with error
}

func TestCreateKey_Dashboard_DefaultRPM(t *testing.T) {
	h, ps, _, _, e := setupDashboard(t)

	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)

	// No rate_limit_rpm — should default to 60.
	form := url.Values{
		"name":        {"my-key"},
		"provider_id": {p.ID},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/keys", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.CreateKey(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRevokeKey_Dashboard(t *testing.T) {
	h, ps, ks, _, e := setupDashboard(t)

	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("torevoke", p.ID, 0)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/keys/"+k.ID+"/revoke", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(k.ID)

	err = h.RevokeKey(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
}

func TestRevokeKey_Dashboard_NotFound(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/keys/nonexistent/revoke", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.RevokeKey(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code) // redirect with flash
}

func TestUsagePage(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/usage", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.UsagePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUsagePage_WithFilters(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/usage?page=2&api_key_id=k1&provider_id=p1&start=2024-01-01T00:00&end=2025-01-01T00:00", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.UsagePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUsagePage_WithRecords(t *testing.T) {
	h, ps, ks, us, e := setupDashboard(t)

	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	k, _, err := ks.Create("k", p.ID, 0)
	require.NoError(t, err)
	err = us.Record(k.ID, p.ID, "gpt-4", 100, 50, 150)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/usage", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.UsagePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUsagePage_InvalidPage(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/usage?page=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.UsagePage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPlaygroundPage(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/playground", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.PlaygroundPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestPlaygroundPage_WithKeys(t *testing.T) {
	h, ps, ks, _, e := setupDashboard(t)

	p, err := ps.Create("test", "https://test.com", "key")
	require.NoError(t, err)
	_, _, err = ks.Create("k1", p.ID, 0)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/playground", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.PlaygroundPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRenderError(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.renderError(c, "login", "something went wrong")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}
