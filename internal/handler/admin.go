package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/lab34/llm-proxy/internal/store"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	providers *store.ProviderStore
	keys      *store.APIKeyStore
	usage     *store.UsageStore
}

func NewAdminHandler(p *store.ProviderStore, k *store.APIKeyStore, u *store.UsageStore) *AdminHandler {
	return &AdminHandler{providers: p, keys: k, usage: u}
}

// ── Providers ──────────────────────────────────────────────────────────

type createProviderReq struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

func (h *AdminHandler) CreateProvider(c echo.Context) error {
	var req createProviderReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" || req.BaseURL == "" || req.APIKey == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name, base_url, and api_key are required")
	}
	p, err := h.providers.Create(req.Name, req.BaseURL, req.APIKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, p)
}

func (h *AdminHandler) ListProviders(c echo.Context) error {
	list, err := h.providers.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if list == nil {
		return c.JSON(http.StatusOK, []interface{}{})
	}
	return c.JSON(http.StatusOK, list)
}

func (h *AdminHandler) GetProvider(c echo.Context) error {
	p, err := h.providers.GetByID(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound, "provider not found")
	}
	return c.JSON(http.StatusOK, p)
}

type updateProviderReq struct {
	Name    *string `json:"name"`
	BaseURL *string `json:"base_url"`
	APIKey  *string `json:"api_key"`
}

func (h *AdminHandler) UpdateProvider(c echo.Context) error {
	var req updateProviderReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	p, err := h.providers.Update(c.Param("id"), req.Name, req.BaseURL, req.APIKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if p == nil {
		return echo.NewHTTPError(http.StatusNotFound, "provider not found")
	}
	return c.JSON(http.StatusOK, p)
}

func (h *AdminHandler) DeleteProvider(c echo.Context) error {
	if err := h.providers.Delete(c.Param("id")); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "provider not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ── API Keys ───────────────────────────────────────────────────────────

type createAPIKeyReq struct {
	Name         string `json:"name"`
	ProviderID   string `json:"provider_id"`
	RateLimitRPM int    `json:"rate_limit_rpm"`
}

func (h *AdminHandler) CreateAPIKey(c echo.Context) error {
	var req createAPIKeyReq
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if req.Name == "" || req.ProviderID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name and provider_id are required")
	}

	// Verify provider exists.
	p, err := h.providers.GetByID(req.ProviderID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if p == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "provider not found")
	}

	k, rawKey, err := h.keys.Create(req.Name, req.ProviderID, req.RateLimitRPM)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"id":             k.ID,
		"name":           k.Name,
		"key":            rawKey,
		"provider_id":    k.ProviderID,
		"rate_limit_rpm": k.RateLimitRPM,
		"created_at":     k.CreatedAt,
	})
}

func (h *AdminHandler) ListAPIKeys(c echo.Context) error {
	list, err := h.keys.List()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if list == nil {
		return c.JSON(http.StatusOK, []struct{}{})
	}
	return c.JSON(http.StatusOK, list)
}

func (h *AdminHandler) RevokeAPIKey(c echo.Context) error {
	if err := h.keys.Revoke(c.Param("id")); err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "api key not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ── Usage ──────────────────────────────────────────────────────────────

func (h *AdminHandler) QueryUsage(c echo.Context) error {
	q := store.UsageQuery{
		APIKeyID:   c.QueryParam("api_key_id"),
		ProviderID: c.QueryParam("provider_id"),
		Limit:      intParam(c, "limit", 100),
		Offset:     intParam(c, "offset", 0),
	}

	if s := c.QueryParam("start"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid start time")
		}
		q.Start = &t
	}
	if s := c.QueryParam("end"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid end time")
		}
		q.End = &t
	}

	result, err := h.usage.Query(q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

func intParam(c echo.Context, name string, def int) int {
	s := c.QueryParam(name)
	if s == "" {
		return def
	}
	var v int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return def
		}
		v = v*10 + int(ch-'0')
	}
	return v
}
