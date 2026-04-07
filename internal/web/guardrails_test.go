package web

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGuardrailsPage(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/guardrails", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.GuardrailsPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestCreateGuardrail_Dashboard_Success_Reject(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{
		"pattern": {`\bpassword\b`},
		"mode":    {"reject"},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrails", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGuardrail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "/dashboard/guardrails")
}

func TestCreateGuardrail_Dashboard_Success_Replace(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{
		"pattern":    {"secret"},
		"mode":       {"replace"},
		"replace_by": {"[REDACTED]"},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrails", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGuardrail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
}

func TestCreateGuardrail_Dashboard_MissingFields(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{"pattern": {""}}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrails", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGuardrail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // re-renders with error
}

func TestCreateGuardrail_Dashboard_InvalidMode(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{
		"pattern": {"test"},
		"mode":    {"invalid"},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrails", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGuardrail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // re-renders with error
}

func TestCreateGuardrail_Dashboard_InvalidRegex(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	form := url.Values{
		"pattern": {"[invalid"},
		"mode":    {"reject"},
	}
	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrails", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.CreateGuardrail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code) // re-renders with error
}

func TestDeleteGuardrail_Dashboard_Success(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	// Create a guardrail first.
	g, err := h.guardrails.Create(`test`, "reject", "")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrails/"+g.ID+"/delete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(g.ID)

	err = h.DeleteGuardrail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Contains(t, rec.Header().Get("Location"), "Guardrail+deleted")
}

func TestDeleteGuardrail_Dashboard_NotFound(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrails/nonexistent/delete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.DeleteGuardrail(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code) // redirect with flash error
}

func TestDeleteGuardrailEvent_Dashboard_NotFound(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	req := httptest.NewRequest(http.MethodPost, "/dashboard/guardrail-events/nonexistent/delete", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	err := h.DeleteGuardrailEvent(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusFound, rec.Code) // redirect with flash error
}

func TestGuardrailsPage_WithData(t *testing.T) {
	h, _, _, _, e := setupDashboard(t)

	_, err := h.guardrails.Create(`\bpassword\b`, "reject", "")
	require.NoError(t, err)
	_, err = h.guardrails.Create(`secret`, "replace", "[REDACTED]")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/guardrails?flash=test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = h.GuardrailsPage(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "password")
	assert.Contains(t, rec.Body.String(), "secret")
}
