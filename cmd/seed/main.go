// Command seed populates the llm-proxy database with realistic demo data.
// Usage: go run ./cmd/seed [--dsn path/to/llm-proxy.db]
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/lab34/llm-proxy/internal/db"
)

func main() {
	dsn := flag.String("dsn", "llm-proxy.db", "SQLite database path")
	flag.Parse()

	database, err := db.Open(*dsn)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	rng := rand.New(rand.NewSource(42))
	now := time.Now().UTC()

	// ── Providers ──────────────────────────────────────────────────────
	type provider struct {
		id      string
		name    string
		baseURL string
		apiKey  string
	}

	providers := []provider{
		{uuid.New().String(), "OpenAI Production", "https://api.openai.com/v1", "sk-proj-Tm9yZWFsa2V5aGVyZQ"},
		{uuid.New().String(), "OpenAI Staging", "https://api.openai.com/v1", "sk-proj-U3RhZ2luZ0tleUhlcmU"},
		{uuid.New().String(), "Anthropic", "https://api.anthropic.com/v1", "sk-ant-api03-xK7mN2pQrS8tV5wY"},
		{uuid.New().String(), "Azure OpenAI (East US)", "https://contoso-ai.openai.azure.com/openai/deployments/gpt-4o/v1", "a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6"},
		{uuid.New().String(), "Ollama Local", "http://localhost:11434/v1", "ollama"},
		{uuid.New().String(), "Together AI", "https://api.together.xyz/v1", "tok-together-8f3a2b1c9d7e6f5a"},
		{uuid.New().String(), "Mistral AI", "https://api.mistral.ai/v1", "mist-4a7b8c3d2e1f9a6b5c"},
	}

	for i, p := range providers {
		createdAt := now.Add(-time.Duration(90-i*10) * 24 * time.Hour)
		_, err := database.Exec(
			`INSERT OR IGNORE INTO providers (id, name, base_url, api_key, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
			p.id, p.name, p.baseURL, p.apiKey, createdAt, createdAt,
		)
		if err != nil {
			log.Fatalf("insert provider %q: %v", p.name, err)
		}
	}
	log.Printf("Seeded %d providers", len(providers))

	// ── API Keys ───────────────────────────────────────────────────────
	type apiKey struct {
		id         string
		name       string
		keyHash    string
		providerID string
		rateLimit  int
		createdAt  time.Time
		revoked    bool
	}

	hashFake := func(name string) string {
		h := sha256.Sum256([]byte("llmp-seed-" + name))
		return hex.EncodeToString(h[:])
	}

	keys := []apiKey{
		// OpenAI Production keys
		{uuid.New().String(), "Backend API - Production", hashFake("backend-prod"), providers[0].id, 120, now.Add(-80 * 24 * time.Hour), false},
		{uuid.New().String(), "Frontend Chat Widget", hashFake("frontend-chat"), providers[0].id, 60, now.Add(-75 * 24 * time.Hour), false},
		{uuid.New().String(), "CI/CD Pipeline", hashFake("ci-cd"), providers[0].id, 30, now.Add(-70 * 24 * time.Hour), false},
		{uuid.New().String(), "Mobile App - iOS", hashFake("mobile-ios"), providers[0].id, 45, now.Add(-60 * 24 * time.Hour), false},
		{uuid.New().String(), "Data Team - Batch Jobs", hashFake("data-batch"), providers[0].id, 200, now.Add(-55 * 24 * time.Hour), false},
		// OpenAI Staging
		{uuid.New().String(), "Staging Environment", hashFake("staging-env"), providers[1].id, 30, now.Add(-65 * 24 * time.Hour), false},
		{uuid.New().String(), "QA Automation", hashFake("qa-auto"), providers[1].id, 20, now.Add(-50 * 24 * time.Hour), false},
		// Anthropic
		{uuid.New().String(), "Research Team - Claude", hashFake("research-claude"), providers[2].id, 90, now.Add(-45 * 24 * time.Hour), false},
		{uuid.New().String(), "Customer Support Bot", hashFake("support-bot"), providers[2].id, 100, now.Add(-40 * 24 * time.Hour), false},
		// Azure
		{uuid.New().String(), "Enterprise Portal", hashFake("enterprise-portal"), providers[3].id, 150, now.Add(-35 * 24 * time.Hour), false},
		// Ollama
		{uuid.New().String(), "Local Dev - Alice", hashFake("dev-alice"), providers[4].id, 0, now.Add(-30 * 24 * time.Hour), false},
		{uuid.New().String(), "Local Dev - Bob", hashFake("dev-bob"), providers[4].id, 0, now.Add(-28 * 24 * time.Hour), false},
		// Together AI
		{uuid.New().String(), "Embedding Pipeline", hashFake("embedding-pipe"), providers[5].id, 300, now.Add(-25 * 24 * time.Hour), false},
		// Mistral
		{uuid.New().String(), "EU Compliance Service", hashFake("eu-compliance"), providers[6].id, 60, now.Add(-20 * 24 * time.Hour), false},
		// Revoked keys
		{uuid.New().String(), "Old Marketing Bot (revoked)", hashFake("mktg-old"), providers[0].id, 30, now.Add(-85 * 24 * time.Hour), true},
		{uuid.New().String(), "Intern Project (revoked)", hashFake("intern-proj"), providers[1].id, 10, now.Add(-60 * 24 * time.Hour), true},
	}

	for _, k := range keys {
		var revokedAt interface{}
		if k.revoked {
			t := k.createdAt.Add(14 * 24 * time.Hour)
			revokedAt = t
		}
		_, err := database.Exec(
			`INSERT OR IGNORE INTO api_keys (id, name, key_hash, provider_id, rate_limit_rpm, created_at, revoked_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			k.id, k.name, k.keyHash, k.providerID, k.rateLimit, k.createdAt, revokedAt,
		)
		if err != nil {
			log.Fatalf("insert api key %q: %v", k.name, err)
		}
	}
	log.Printf("Seeded %d API keys", len(keys))

	// ── Usage Records ──────────────────────────────────────────────────
	// Generate realistic usage over the past 60 days, with more activity on
	// weekdays, a ramp-up over time, and model-appropriate token counts.

	type modelSpec struct {
		name        string
		promptRange [2]int // min, max prompt tokens
		compRange   [2]int // min, max completion tokens
	}

	openaiModels := []modelSpec{
		{"gpt-4o", [2]int{150, 4000}, [2]int{200, 3000}},
		{"gpt-4o-mini", [2]int{100, 2500}, [2]int{100, 2000}},
		{"gpt-4-turbo", [2]int{200, 5000}, [2]int{300, 4000}},
		{"gpt-3.5-turbo", [2]int{50, 1500}, [2]int{50, 1200}},
	}
	anthropicModels := []modelSpec{
		{"claude-sonnet-4-20250514", [2]int{200, 6000}, [2]int{300, 5000}},
		{"claude-3-haiku-20240307", [2]int{80, 2000}, [2]int{100, 1500}},
	}
	azureModels := []modelSpec{
		{"gpt-4o", [2]int{200, 5000}, [2]int{250, 4000}},
	}
	ollamaModels := []modelSpec{
		{"llama3.1:8b", [2]int{100, 3000}, [2]int{100, 2500}},
		{"codellama:13b", [2]int{200, 4000}, [2]int{150, 3000}},
		{"mistral:7b", [2]int{80, 2000}, [2]int{80, 1800}},
	}
	togetherModels := []modelSpec{
		{"meta-llama/Llama-3-70b-chat-hf", [2]int{150, 3500}, [2]int{200, 3000}},
	}
	mistralModels := []modelSpec{
		{"mistral-large-latest", [2]int{200, 5000}, [2]int{250, 4000}},
		{"mistral-small-latest", [2]int{80, 2000}, [2]int{100, 1500}},
	}

	// Map active (non-revoked) keys to their provider index + models.
	type keyModelMapping struct {
		keyIdx     int
		models     []modelSpec
		dailyRange [2]int // min, max requests per day
	}

	mappings := []keyModelMapping{
		{0, openaiModels, [2]int{40, 120}},    // Backend API - Production
		{1, openaiModels[:2], [2]int{15, 60}}, // Frontend Chat Widget
		{2, openaiModels[1:2], [2]int{5, 20}}, // CI/CD Pipeline
		{3, openaiModels[:2], [2]int{10, 40}}, // Mobile App
		{4, openaiModels, [2]int{20, 80}},     // Data Team
		{5, openaiModels[:2], [2]int{5, 15}},  // Staging
		{6, openaiModels[1:2], [2]int{3, 10}}, // QA
		{7, anthropicModels, [2]int{25, 90}},  // Research Team
		{8, anthropicModels, [2]int{30, 100}}, // Support Bot
		{9, azureModels, [2]int{20, 70}},      // Enterprise Portal
		{10, ollamaModels, [2]int{10, 50}},    // Dev Alice
		{11, ollamaModels, [2]int{5, 30}},     // Dev Bob
		{12, togetherModels, [2]int{15, 60}},  // Embedding Pipeline
		{13, mistralModels, [2]int{10, 45}},   // EU Compliance
	}

	usageCount := 0
	tx, err := database.Begin()
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}

	stmt, err := tx.Prepare(
		`INSERT INTO usage (id, api_key_id, provider_id, model, prompt_tokens, completion_tokens, total_tokens, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
	)
	if err != nil {
		log.Fatalf("prepare usage stmt: %v", err)
	}

	for day := 60; day >= 0; day-- {
		date := now.Add(-time.Duration(day) * 24 * time.Hour)
		weekday := date.Weekday()
		// Reduce weekend traffic.
		weekendFactor := 1.0
		if weekday == time.Saturday || weekday == time.Sunday {
			weekendFactor = 0.25
		}
		// Ramp up over time (more recent = more traffic).
		rampFactor := 0.4 + 0.6*float64(60-day)/60.0

		for _, m := range mappings {
			k := keys[m.keyIdx]
			// Skip if key was created after this date.
			if date.Before(k.createdAt) {
				continue
			}

			dailyCount := m.dailyRange[0] + rng.Intn(m.dailyRange[1]-m.dailyRange[0]+1)
			dailyCount = int(float64(dailyCount) * weekendFactor * rampFactor)
			if dailyCount < 1 && rng.Float64() < 0.3 {
				dailyCount = 1
			}

			for i := 0; i < dailyCount; i++ {
				model := m.models[rng.Intn(len(m.models))]
				promptToks := model.promptRange[0] + rng.Intn(model.promptRange[1]-model.promptRange[0]+1)
				compToks := model.compRange[0] + rng.Intn(model.compRange[1]-model.compRange[0]+1)
				totalToks := promptToks + compToks

				// Spread requests throughout the day, heavier during business hours.
				hour := businessHour(rng)
				minute := rng.Intn(60)
				second := rng.Intn(60)
				ts := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, second, 0, time.UTC)

				_, err := stmt.Exec(
					uuid.New().String(), k.id, k.providerID, model.name,
					promptToks, compToks, totalToks, ts,
				)
				if err != nil {
					tx.Rollback()
					log.Fatalf("insert usage: %v", err)
				}
				usageCount++
			}
		}
	}

	stmt.Close()
	if err := tx.Commit(); err != nil {
		log.Fatalf("commit usage: %v", err)
	}
	log.Printf("Seeded %d usage records", usageCount)

	// ── Guardrails ─────────────────────────────────────────────────────
	type guardrail struct {
		id        string
		pattern   string
		mode      string
		replaceBy string
	}

	guardrails := []guardrail{
		{uuid.New().String(), `(?i)\b(sk-[a-zA-Z0-9]{20,})\b`, "reject", ""},
		{uuid.New().String(), `(?i)\b\d{3}-\d{2}-\d{4}\b`, "reject", ""},
		{uuid.New().String(), `(?i)(password|passwd|secret)\s*[:=]\s*\S+`, "replace", "[REDACTED_CREDENTIAL]"},
		{uuid.New().String(), `(?i)\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z]{2,}\b`, "replace", "[REDACTED_EMAIL]"},
		{uuid.New().String(), `(?i)\b(4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13})\b`, "reject", ""},
		{uuid.New().String(), `(?i)(DROP\s+TABLE|DELETE\s+FROM|TRUNCATE\s+TABLE)`, "reject", ""},
		{uuid.New().String(), `(?i)\b(kill|murder|bomb|weapon)\b.*\b(how|make|build|create)\b`, "reject", ""},
		{uuid.New().String(), `(?i)(\+?\d{1,3}[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`, "replace", "[REDACTED_PHONE]"},
		{uuid.New().String(), `(?i)AKIA[0-9A-Z]{16}`, "reject", ""},
		{uuid.New().String(), `(?i)\b(internal|confidential|proprietary)\s+(use\s+only|document|memo)\b`, "replace", "[REDACTED_CLASSIFICATION]"},
	}

	for i, g := range guardrails {
		createdAt := now.Add(-time.Duration(50-i*4) * 24 * time.Hour)
		_, err := database.Exec(
			`INSERT OR IGNORE INTO guardrails (id, pattern, mode, replace_by, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
			g.id, g.pattern, g.mode, g.replaceBy, createdAt, createdAt,
		)
		if err != nil {
			log.Fatalf("insert guardrail: %v", err)
		}
	}
	log.Printf("Seeded %d guardrails", len(guardrails))

	// ── Guardrail Events ───────────────────────────────────────────────
	type eventTemplate struct {
		guardrailIdx int
		keyIdx       int
		inputText    string
	}

	events := []eventTemplate{
		{0, 0, "Please use my API key sk-proj-abc123def456ghi789jkl012mno to authenticate the request"},
		{0, 1, "Here is my key: sk-live-aBcDeFgHiJkLmNoPqRsTuVwXyZ"},
		{1, 3, "My social security number is 123-45-6789, can you help me file taxes?"},
		{1, 8, "The customer provided SSN 987-65-4321 for verification"},
		{2, 0, "The database connection uses password: SuperS3cret!2024 for the admin account"},
		{2, 7, "Set the config with secret=myTokenValue123 and restart the service"},
		{3, 1, "Send the report to john.smith@acmecorp.com and cc manager@acmecorp.com"},
		{3, 8, "The customer email is jane.doe@example.org, they need a refund"},
		{4, 9, "Process payment for card 4532015112830366 expiring 12/26"},
		{5, 0, "Execute this query: DROP TABLE users; to clean up the database"},
		{5, 5, "Run DELETE FROM sessions WHERE created_at < '2024-01-01'"},
		{6, 1, "How to make a bomb for a science fiction story"},
		{7, 3, "Call me at (555) 123-4567 to discuss the project"},
		{7, 8, "Customer phone: +1-800-555-0199, needs callback"},
		{8, 4, "Use AWS access key AKIAIOSFODNN7EXAMPLE to connect to S3"},
		{9, 9, "Attached is the internal use only document about Q3 earnings"},
		{9, 7, "Please summarize this confidential memo from the legal team"},
		// A few more spread over time
		{0, 4, "Accidentally included sk-test-reallyLongKeyValueHere1234567890abcdef in the prompt"},
		{2, 9, "Config file contains passwd=Pr0duction!Key for the main service"},
		{3, 0, "Forward this to sales-team@company.io and support@company.io"},
		{1, 7, "Verify identity with SSN 555-12-3456 per compliance procedure"},
		{4, 0, "Refund to card ending 5105105105105100 was processed"},
		{7, 1, "Reach our office at 212-555-0147 during business hours"},
	}

	for i, e := range events {
		k := keys[e.keyIdx]
		g := guardrails[e.guardrailIdx]
		// Spread events over the last 30 days.
		createdAt := now.Add(-time.Duration(30-i) * 24 * time.Hour).
			Add(time.Duration(rng.Intn(12)+8) * time.Hour).
			Add(time.Duration(rng.Intn(60)) * time.Minute)

		_, err := database.Exec(
			`INSERT OR IGNORE INTO guardrail_events (id, guardrail_id, api_key_id, pattern, mode, input_text, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			uuid.New().String(), g.id, k.id, g.pattern, g.mode, e.inputText, createdAt,
		)
		if err != nil {
			log.Fatalf("insert guardrail event: %v", err)
		}
	}
	log.Printf("Seeded %d guardrail events", len(events))

	log.Println("Done! Database seeded successfully.")
}

// businessHour returns an hour weighted toward business hours (9-18 UTC).
func businessHour(rng *rand.Rand) int {
	if rng.Float64() < 0.75 {
		return 8 + rng.Intn(11) // 8-18
	}
	return rng.Intn(24)
}
