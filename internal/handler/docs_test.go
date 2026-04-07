package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsHandler_Spec(t *testing.T) {
	spec := []byte("openapi: 3.1.0\ninfo:\n  title: Test")
	h := NewDocsHandler(spec)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.Spec(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/x-yaml", rec.Header().Get("Content-Type"))
	assert.Equal(t, "openapi: 3.1.0\ninfo:\n  title: Test", rec.Body.String())
}

func TestDocsHandler_SwaggerUI(t *testing.T) {
	h := NewDocsHandler([]byte("spec"))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.SwaggerUI(c)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, rec.Body.String(), "swagger-ui")
	assert.Contains(t, rec.Body.String(), "/openapi.yaml")
}
