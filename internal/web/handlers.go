package web

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lab34/llm-proxy/internal/models"
	"github.com/lab34/llm-proxy/internal/store"
	"github.com/labstack/echo/v4"
)

// DashboardHandler serves all dashboard UI pages.
type DashboardHandler struct {
	providers  *store.ProviderStore
	keys       *store.APIKeyStore
	usage      *store.UsageStore
	adminToken string
	secret     string
	// rawKeys caches the plaintext key for the playground (only available right after creation within the session).
	// For the playground, we need the raw keys — they are passed from the DB lookup.
	// Since raw keys are not stored, the playground uses keys created during this session.
}

// NewDashboardHandler creates a new handler for the dashboard UI.
func NewDashboardHandler(providers *store.ProviderStore, keys *store.APIKeyStore, usage *store.UsageStore, adminToken, secret string) *DashboardHandler {
	return &DashboardHandler{
		providers:  providers,
		keys:       keys,
		usage:      usage,
		adminToken: adminToken,
		secret:     secret,
	}
}

// ── Auth pages ─────────────────────────────────────────────────────────

func (h *DashboardHandler) LoginPage(c echo.Context) error {
	// If already logged in, redirect to providers
	if _, ok := getSessionToken(c, h.secret); ok {
		return c.Redirect(http.StatusFound, "/dashboard/providers")
	}
	return c.Render(http.StatusOK, "login", map[string]interface{}{
		"Error": "",
	})
}

func (h *DashboardHandler) LoginSubmit(c echo.Context) error {
	token := c.FormValue("token")
	if token != h.adminToken {
		return c.Render(http.StatusOK, "login", map[string]interface{}{
			"Error": "Invalid admin token.",
		})
	}
	setSessionCookie(c, token, h.secret)
	return c.Redirect(http.StatusFound, "/dashboard/providers")
}

func (h *DashboardHandler) Logout(c echo.Context) error {
	clearSessionCookie(c)
	return c.Redirect(http.StatusFound, "/dashboard/login")
}

func (h *DashboardHandler) Index(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/dashboard/providers")
}

// ── Providers ──────────────────────────────────────────────────────────

type providersPageData struct {
	Title     string
	Active    string
	Flash     string
	Error     string
	Providers []models.Provider
}

func (h *DashboardHandler) ProvidersPage(c echo.Context) error {
	providers, err := h.providers.List()
	if err != nil {
		return h.renderError(c, "providers", "Failed to load providers: "+err.Error())
	}
	return c.Render(http.StatusOK, "providers", providersPageData{
		Title:     "Providers",
		Active:    "providers",
		Flash:     c.QueryParam("flash"),
		Providers: providers,
	})
}

func (h *DashboardHandler) CreateProvider(c echo.Context) error {
	name := c.FormValue("name")
	baseURL := c.FormValue("base_url")
	apiKey := c.FormValue("api_key")

	if name == "" || baseURL == "" || apiKey == "" {
		return h.renderProvidersWithError(c, "Name, Base URL, and API Key are required.")
	}

	_, err := h.providers.Create(name, baseURL, apiKey)
	if err != nil {
		return h.renderProvidersWithError(c, "Failed to create provider: "+err.Error())
	}

	return c.Redirect(http.StatusFound, "/dashboard/providers?flash=Provider+created+successfully")
}

func (h *DashboardHandler) DeleteProvider(c echo.Context) error {
	id := c.Param("id")
	if err := h.providers.Delete(id); err != nil {
		return c.Redirect(http.StatusFound, "/dashboard/providers?flash=Failed+to+delete+provider")
	}
	return c.Redirect(http.StatusFound, "/dashboard/providers?flash=Provider+deleted")
}

func (h *DashboardHandler) renderProvidersWithError(c echo.Context, errMsg string) error {
	providers, _ := h.providers.List()
	return c.Render(http.StatusOK, "providers", providersPageData{
		Title:     "Providers",
		Active:    "providers",
		Error:     errMsg,
		Providers: providers,
	})
}

// ── API Keys ───────────────────────────────────────────────────────────

type keysPageData struct {
	Title     string
	Active    string
	Flash     string
	Error     string
	Keys      []models.APIKey
	Providers []models.Provider
	NewKey    string // only set right after creation
}

func (h *DashboardHandler) KeysPage(c echo.Context) error {
	keys, err := h.keys.List()
	if err != nil {
		return h.renderError(c, "keys", "Failed to load keys: "+err.Error())
	}
	providers, err := h.providers.List()
	if err != nil {
		return h.renderError(c, "keys", "Failed to load providers: "+err.Error())
	}
	return c.Render(http.StatusOK, "keys", keysPageData{
		Title:     "API Keys",
		Active:    "keys",
		Flash:     c.QueryParam("flash"),
		Keys:      keys,
		Providers: providers,
	})
}

func (h *DashboardHandler) CreateKey(c echo.Context) error {
	name := c.FormValue("name")
	providerID := c.FormValue("provider_id")
	rpmStr := c.FormValue("rate_limit_rpm")

	if name == "" || providerID == "" {
		return h.renderKeysWithError(c, "Name and Provider are required.", "")
	}

	rpm := 60
	if rpmStr != "" {
		if v, err := strconv.Atoi(rpmStr); err == nil {
			rpm = v
		}
	}

	// Verify provider exists
	p, err := h.providers.GetByID(providerID)
	if err != nil || p == nil {
		return h.renderKeysWithError(c, "Provider not found.", "")
	}

	k, rawKey, err := h.keys.Create(name, providerID, rpm)
	if err != nil {
		return h.renderKeysWithError(c, "Failed to create key: "+err.Error(), "")
	}
	_ = k

	// Re-render with the new key displayed
	keys, _ := h.keys.List()
	providers, _ := h.providers.List()
	return c.Render(http.StatusOK, "keys", keysPageData{
		Title:     "API Keys",
		Active:    "keys",
		Flash:     "API key created successfully.",
		Keys:      keys,
		Providers: providers,
		NewKey:    rawKey,
	})
}

func (h *DashboardHandler) RevokeKey(c echo.Context) error {
	id := c.Param("id")
	if err := h.keys.Revoke(id); err != nil {
		return c.Redirect(http.StatusFound, "/dashboard/keys?flash=Failed+to+revoke+key")
	}
	return c.Redirect(http.StatusFound, "/dashboard/keys?flash=Key+revoked")
}

func (h *DashboardHandler) renderKeysWithError(c echo.Context, errMsg, newKey string) error {
	keys, _ := h.keys.List()
	providers, _ := h.providers.List()
	return c.Render(http.StatusOK, "keys", keysPageData{
		Title:     "API Keys",
		Active:    "keys",
		Error:     errMsg,
		Keys:      keys,
		Providers: providers,
		NewKey:    newKey,
	})
}

// ── Usage ──────────────────────────────────────────────────────────────

type usagePageData struct {
	Title                 string
	Active                string
	Flash                 string
	Error                 string
	Records               []models.UsageRecord
	Keys                  []models.APIKey
	Providers             []models.Provider
	TotalRecords          int
	TotalPromptTokens     int
	TotalCompletionTokens int
	TotalAllTokens        int
	FilterKeyID           string
	FilterProviderID      string
	FilterStart           string
	FilterEnd             string
	Page                  int
	TotalPages            int
	PaginationQuery       func(int) string
}

func (h *DashboardHandler) UsagePage(c echo.Context) error {
	const perPage = 50

	page := 1
	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}

	q := store.UsageQuery{
		APIKeyID:   c.QueryParam("api_key_id"),
		ProviderID: c.QueryParam("provider_id"),
		Limit:      perPage,
		Offset:     (page - 1) * perPage,
	}

	filterStart := c.QueryParam("start")
	filterEnd := c.QueryParam("end")

	if filterStart != "" {
		t, err := time.Parse("2006-01-02T15:04", filterStart)
		if err == nil {
			q.Start = &t
		}
	}
	if filterEnd != "" {
		t, err := time.Parse("2006-01-02T15:04", filterEnd)
		if err == nil {
			q.End = &t
		}
	}

	result, err := h.usage.Query(q)
	if err != nil {
		return h.renderError(c, "usage", "Failed to load usage: "+err.Error())
	}

	// Calculate totals from the current page
	var promptTotal, completionTotal, allTotal int
	if result.Records != nil {
		for _, r := range result.Records {
			promptTotal += r.PromptTokens
			completionTotal += r.CompletionTokens
			allTotal += r.TotalTokens
		}
	}

	totalPages := (result.Total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}

	keys, _ := h.keys.List()
	providers, _ := h.providers.List()

	paginationQuery := func(p int) string {
		qs := fmt.Sprintf("page=%d", p)
		if q.APIKeyID != "" {
			qs += "&api_key_id=" + q.APIKeyID
		}
		if q.ProviderID != "" {
			qs += "&provider_id=" + q.ProviderID
		}
		if filterStart != "" {
			qs += "&start=" + filterStart
		}
		if filterEnd != "" {
			qs += "&end=" + filterEnd
		}
		return qs
	}

	return c.Render(http.StatusOK, "usage", usagePageData{
		Title:                 "Usage",
		Active:                "usage",
		Records:               result.Records,
		Keys:                  keys,
		Providers:             providers,
		TotalRecords:          result.Total,
		TotalPromptTokens:     promptTotal,
		TotalCompletionTokens: completionTotal,
		TotalAllTokens:        allTotal,
		FilterKeyID:           q.APIKeyID,
		FilterProviderID:      q.ProviderID,
		FilterStart:           filterStart,
		FilterEnd:             filterEnd,
		Page:                  page,
		TotalPages:            totalPages,
		PaginationQuery:       paginationQuery,
	})
}

// ── Playground ─────────────────────────────────────────────────────────

type playgroundKeyInfo struct {
	ID         string
	Name       string
	ProviderID string
	RevokedAt  *time.Time
	RawKey     string // We don't have raw keys; playground will need to use keys entered by user
}

type playgroundPageData struct {
	Title  string
	Active string
	Flash  string
	Error  string
	Keys   []playgroundKeyInfo
}

func (h *DashboardHandler) PlaygroundPage(c echo.Context) error {
	keys, err := h.keys.List()
	if err != nil {
		return h.renderError(c, "playground", "Failed to load keys: "+err.Error())
	}

	// We cannot recover raw keys from hashes. The playground will need the user
	// to paste a key or we need a different approach. For now, we show a manual
	// key input field approach instead.
	var keyInfos []playgroundKeyInfo
	for _, k := range keys {
		keyInfos = append(keyInfos, playgroundKeyInfo{
			ID:         k.ID,
			Name:       k.Name,
			ProviderID: k.ProviderID,
			RevokedAt:  k.RevokedAt,
		})
	}

	return c.Render(http.StatusOK, "playground", playgroundPageData{
		Title:  "Playground",
		Active: "playground",
		Keys:   keyInfos,
	})
}

// ── Helpers ────────────────────────────────────────────────────────────

func (h *DashboardHandler) renderError(c echo.Context, page, errMsg string) error {
	return c.Render(http.StatusOK, page, map[string]interface{}{
		"Title":  "Error",
		"Active": page,
		"Error":  errMsg,
	})
}
