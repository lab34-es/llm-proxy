package web

import (
	"embed"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

//go:embed frontend/dist/*
var distFS embed.FS

// SPAHandler serves the React SPA from the embedded dist directory.
// For any path that doesn't match a static file, it serves index.html
// (so that client-side routing works).
func SPAHandler() echo.HandlerFunc {
	sub, err := fs.Sub(distFS, "frontend/dist")
	if err != nil {
		panic("failed to create sub-filesystem for frontend/dist: " + err.Error())
	}

	return func(c echo.Context) error {
		path := c.Request().URL.Path

		// Strip the /dashboard prefix to get the file path within dist/.
		filePath := strings.TrimPrefix(path, "/dashboard")
		filePath = strings.TrimPrefix(filePath, "/")

		if filePath == "" {
			filePath = "index.html"
		}

		// Try to open the file. If it exists, serve it directly.
		return serveFile(c, sub, filePath)
	}
}

func serveFile(c echo.Context, fsys fs.FS, path string) error {
	f, err := fsys.Open(path)
	if err != nil {
		// File not found — serve index.html for client-side routing.
		f, err = fsys.Open("index.html")
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "index.html not found")
		}
		path = "index.html"
	}
	defer f.Close()

	// Check if it's a directory; if so, serve index.html.
	stat, err := f.Stat()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if stat.IsDir() {
		f.Close()
		f, err = fsys.Open("index.html")
		if err != nil {
			return echo.NewHTTPError(http.StatusNotFound, "index.html not found")
		}
		defer f.Close()
		path = "index.html"
	}

	// Determine content type.
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Response().Header().Set("Content-Type", contentType)
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response(), f)
	return err
}
