package main

import (
	_ "embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"github.com/lab34/llm-proxy/internal/config"
	"github.com/lab34/llm-proxy/internal/db"
	"github.com/lab34/llm-proxy/internal/handler"
	"github.com/lab34/llm-proxy/internal/middleware"
	"github.com/lab34/llm-proxy/internal/proxy"
	"github.com/lab34/llm-proxy/internal/store"
	"github.com/lab34/llm-proxy/internal/web"
)

//go:embed openapi.yaml
var specBytes []byte

func main() {
	cfg := config.Load()

	if cfg.AdminToken == "" {
		log.Fatal("ADMIN_TOKEN environment variable is required")
	}

	database, err := db.Open(cfg.DSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	// Stores.
	providerStore := store.NewProviderStore(database)
	keyStore := store.NewAPIKeyStore(database)
	usageStore := store.NewUsageStore(database)
	guardrailStore := store.NewGuardrailStore(database)
	guardrailEventStore := store.NewGuardrailEventStore(database)

	// Handlers.
	adminH := handler.NewAdminHandler(providerStore, keyStore, usageStore, guardrailStore, guardrailEventStore)
	forwarder := proxy.NewForwarder(usageStore, guardrailStore, guardrailEventStore)
	proxyH := handler.NewProxyHandler(forwarder, providerStore)
	docsH := handler.NewDocsHandler(specBytes)
	dashH := web.NewDashboardHandler(providerStore, keyStore, usageStore, guardrailStore, guardrailEventStore, cfg.AdminToken, cfg.SessionSecret)

	// Middleware.
	rateLimiter := middleware.NewRateLimiter()

	// Echo setup.
	e := echo.New()
	e.HideBanner = true
	e.Renderer = web.NewRenderer()
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())

	// ── Documentation (public) ─────────────────────────────────────────
	e.GET("/openapi.yaml", docsH.Spec)
	e.GET("/docs", docsH.SwaggerUI)

	// ── Dashboard (template-based UI) ──────────────────────────────────
	dash := e.Group("/dashboard")
	dash.GET("", dashH.Index)
	dash.GET("/login", dashH.LoginPage)
	dash.POST("/login", dashH.LoginSubmit)
	dash.GET("/logout", dashH.Logout)

	// Serve embedded static assets.
	staticFS, _ := fs.Sub(web.StaticFS, "static")
	dash.GET("/static/*", echo.WrapHandler(http.StripPrefix("/dashboard/static/", http.FileServer(http.FS(staticFS)))))

	// Protected dashboard routes.
	dashAuth := dash.Group("", web.SessionAuth(cfg.AdminToken, cfg.SessionSecret))
	dashAuth.GET("/providers", dashH.ProvidersPage)
	dashAuth.POST("/providers", dashH.CreateProvider)
	dashAuth.POST("/providers/:id/delete", dashH.DeleteProvider)
	dashAuth.GET("/keys", dashH.KeysPage)
	dashAuth.POST("/keys", dashH.CreateKey)
	dashAuth.POST("/keys/:id/revoke", dashH.RevokeKey)
	dashAuth.GET("/usage", dashH.UsagePage)
	dashAuth.GET("/guardrails", dashH.GuardrailsPage)
	dashAuth.POST("/guardrails", dashH.CreateGuardrail)
	dashAuth.POST("/guardrails/:id/delete", dashH.DeleteGuardrail)
	dashAuth.POST("/guardrail-events/:id/delete", dashH.DeleteGuardrailEvent)
	dashAuth.GET("/playground", dashH.PlaygroundPage)

	// ── Admin routes (admin token auth) ────────────────────────────────
	admin := e.Group("/admin", middleware.AdminAuth(cfg.AdminToken))

	admin.POST("/providers", adminH.CreateProvider)
	admin.GET("/providers", adminH.ListProviders)
	admin.GET("/providers/:id", adminH.GetProvider)
	admin.PUT("/providers/:id", adminH.UpdateProvider)
	admin.DELETE("/providers/:id", adminH.DeleteProvider)

	admin.POST("/keys", adminH.CreateAPIKey)
	admin.GET("/keys", adminH.ListAPIKeys)
	admin.DELETE("/keys/:id", adminH.RevokeAPIKey)

	admin.GET("/usage", adminH.QueryUsage)

	admin.POST("/guardrails", adminH.CreateGuardrail)
	admin.GET("/guardrails", adminH.ListGuardrails)
	admin.GET("/guardrails/:id", adminH.GetGuardrail)
	admin.DELETE("/guardrails/:id", adminH.DeleteGuardrail)

	admin.GET("/guardrail-events", adminH.ListGuardrailEvents)
	admin.GET("/guardrail-events/:id", adminH.GetGuardrailEvent)
	admin.DELETE("/guardrail-events/:id", adminH.DeleteGuardrailEvent)

	// ── Proxy routes (proxy API key auth + rate limit) ─────────────────
	v1 := e.Group("/v1",
		middleware.ProxyAuth(keyStore),
		rateLimiter.Middleware(),
	)

	v1.POST("/chat/completions", proxyH.ChatCompletion)
	v1.GET("/models", proxyH.ListModels)

	// Start.
	log.Printf("LLM Proxy listening on %s", cfg.Addr)
	log.Printf("Swagger UI: http://localhost%s/docs", cfg.Addr)
	log.Printf("Dashboard:  http://localhost%s/dashboard", cfg.Addr)
	if err := e.Start(cfg.Addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}
