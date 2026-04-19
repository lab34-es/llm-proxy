package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSPAHandler_ServesIndexHTML(t *testing.T) {
	e := echo.New()
	handler := SPAHandler()

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html")
}

func TestSPAHandler_ClientSideRouting(t *testing.T) {
	e := echo.New()
	handler := SPAHandler()

	// Request a client-side route — should still serve index.html.
	req := httptest.NewRequest(http.MethodGet, "/dashboard/providers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html")
}

func TestSPAHandler_ServesStaticAssets(t *testing.T) {
	e := echo.New()
	handler := SPAHandler()

	// The built dist/ contains at least index.html.
	req := httptest.NewRequest(http.MethodGet, "/dashboard/index.html", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "<html")
}
