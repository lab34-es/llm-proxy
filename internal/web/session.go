package web

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	sessionCookieName = "llmp_session"
	sessionMaxAge     = 24 * time.Hour
)

// signToken creates an HMAC-SHA256 signed token from the admin token.
// Format: base64(adminToken) + "." + base64(hmac)
func signToken(adminToken, secret string) string {
	payload := base64.RawURLEncoding.EncodeToString([]byte(adminToken))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return payload + "." + sig
}

// verifyToken checks the HMAC signature and returns the admin token if valid.
func verifyToken(token, secret string) (string, bool) {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return "", false
	}
	payload, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return "", false
	}

	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", false
	}
	return string(decoded), true
}

// setSessionCookie sets the signed session cookie.
func setSessionCookie(c echo.Context, adminToken, secret string) {
	signed := signToken(adminToken, secret)
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    signed,
		Path:     "/dashboard",
		MaxAge:   int(sessionMaxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie removes the session cookie.
func clearSessionCookie(c echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/dashboard",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// getSessionToken extracts and verifies the admin token from the session cookie.
func getSessionToken(c echo.Context, secret string) (string, bool) {
	cookie, err := c.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return "", false
	}
	return verifyToken(cookie.Value, secret)
}

// SessionAuth returns Echo middleware that redirects to login if no valid session.
func SessionAuth(adminToken, secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, ok := getSessionToken(c, secret)
			if !ok || token != adminToken {
				return c.Redirect(http.StatusFound, "/dashboard/login")
			}
			// Store in context for handlers to use in API calls
			c.Set("admin_token", token)
			return next(c)
		}
	}
}
