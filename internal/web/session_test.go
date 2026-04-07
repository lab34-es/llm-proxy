package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignAndVerifyToken(t *testing.T) {
	token := "my-admin-token"
	secret := "my-secret"

	signed := signToken(token, secret)
	assert.NotEmpty(t, signed)
	assert.Contains(t, signed, ".")

	decoded, ok := verifyToken(signed, secret)
	assert.True(t, ok)
	assert.Equal(t, token, decoded)
}

func TestVerifyToken_InvalidSignature(t *testing.T) {
	signed := signToken("token", "secret1")
	_, ok := verifyToken(signed, "secret2") // different secret
	assert.False(t, ok)
}

func TestVerifyToken_MalformedToken(t *testing.T) {
	_, ok := verifyToken("no-dot-separator", "secret")
	assert.False(t, ok)
}

func TestVerifyToken_InvalidBase64Payload(t *testing.T) {
	// Craft a token with valid HMAC but broken base64 payload.
	_, ok := verifyToken("!!!invalid.sig", "secret")
	assert.False(t, ok)
}

func TestSetSessionCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	setSessionCookie(c, "admin-token", "secret")

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, sessionCookieName, cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
	assert.Equal(t, "/dashboard", cookies[0].Path)
	assert.True(t, cookies[0].HttpOnly)
}

func TestClearSessionCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	clearSessionCookie(c)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, sessionCookieName, cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
	assert.Equal(t, -1, cookies[0].MaxAge)
}

func TestGetSessionToken_Valid(t *testing.T) {
	e := echo.New()
	signed := signToken("admin-token", "secret")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: signed})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	token, ok := getSessionToken(c, "secret")
	assert.True(t, ok)
	assert.Equal(t, "admin-token", token)
}

func TestGetSessionToken_NoCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, ok := getSessionToken(c, "secret")
	assert.False(t, ok)
}

func TestGetSessionToken_EmptyCookie(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: ""})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_, ok := getSessionToken(c, "secret")
	assert.False(t, ok)
}

func TestSessionAuth_ValidSession(t *testing.T) {
	e := echo.New()
	signed := signToken("admin-token", "secret")

	req := httptest.NewRequest(http.MethodGet, "/dashboard/providers", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: signed})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SessionAuth("admin-token", "secret")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSessionAuth_InvalidSession(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/dashboard/providers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SessionAuth("admin-token", "secret")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err) // redirect is not an error
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "/dashboard/login", rec.Header().Get("Location"))
}

func TestSessionAuth_WrongToken(t *testing.T) {
	e := echo.New()
	signed := signToken("wrong-token", "secret")

	req := httptest.NewRequest(http.MethodGet, "/dashboard/providers", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: signed})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := SessionAuth("admin-token", "secret")(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := handler(c)
	assert.NoError(t, err) // redirect
	assert.Equal(t, http.StatusFound, rec.Code)
}
