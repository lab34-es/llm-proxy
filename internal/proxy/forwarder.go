package proxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/lab34/llm-proxy/internal/models"
	"github.com/lab34/llm-proxy/internal/store"
)

// Forwarder proxies OpenAI-compatible requests to an upstream provider.
type Forwarder struct {
	client          *http.Client
	usage           *store.UsageStore
	guardrails      *store.GuardrailStore
	guardrailEvents *store.GuardrailEventStore

	// Guardrail cache with TTL.
	cacheMu      sync.RWMutex
	cachedRules  []models.Guardrail
	cacheExpires time.Time
	cacheTTL     time.Duration
}

func NewForwarder(usage *store.UsageStore, guardrails *store.GuardrailStore, guardrailEvents *store.GuardrailEventStore) *Forwarder {
	return &Forwarder{
		client:          &http.Client{Timeout: 5 * time.Minute},
		usage:           usage,
		guardrails:      guardrails,
		guardrailEvents: guardrailEvents,
		cacheTTL:        30 * time.Second,
	}
}

// chatRequest is the minimal structure we need to inspect from the incoming body.
type chatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// GuardrailRejectedError is returned when a reject guardrail matches.
type GuardrailRejectedError struct {
	Pattern string
}

func (e *GuardrailRejectedError) Error() string {
	return fmt.Sprintf("request blocked by guardrail: pattern %q matched", e.Pattern)
}

// loadGuardrails returns cached guardrails, refreshing if the cache has expired.
func (f *Forwarder) loadGuardrails() ([]models.Guardrail, error) {
	f.cacheMu.RLock()
	if time.Now().Before(f.cacheExpires) && f.cachedRules != nil {
		rules := f.cachedRules
		f.cacheMu.RUnlock()
		return rules, nil
	}
	f.cacheMu.RUnlock()

	f.cacheMu.Lock()
	defer f.cacheMu.Unlock()

	// Double-check after acquiring write lock.
	if time.Now().Before(f.cacheExpires) && f.cachedRules != nil {
		return f.cachedRules, nil
	}

	rules, err := f.guardrails.List()
	if err != nil {
		return nil, fmt.Errorf("load guardrails: %w", err)
	}
	if rules == nil {
		rules = []models.Guardrail{}
	}
	f.cachedRules = rules
	f.cacheExpires = time.Now().Add(f.cacheTTL)
	return rules, nil
}

// applyGuardrails checks all messages against guardrail rules.
// Returns the (possibly modified) messages, or an error if a reject rule matches.
func (f *Forwarder) applyGuardrails(messages []chatMessage, apiKeyID string) ([]chatMessage, error) {
	rules, err := f.loadGuardrails()
	if err != nil {
		return messages, err
	}
	if len(rules) == 0 {
		return messages, nil
	}

	// Compile all patterns.
	type compiledRule struct {
		guardrail models.Guardrail
		re        *regexp.Regexp
	}
	var compiled []compiledRule
	for _, g := range rules {
		re, err := regexp.Compile(g.Pattern)
		if err != nil {
			continue // skip invalid patterns (shouldn't happen since we validate on create)
		}
		compiled = append(compiled, compiledRule{guardrail: g, re: re})
	}

	// Check each message against all rules.
	result := make([]chatMessage, len(messages))
	copy(result, messages)

	for i, msg := range result {
		for _, cr := range compiled {
			if !cr.re.MatchString(msg.Content) {
				continue
			}

			switch cr.guardrail.Mode {
			case "reject":
				// Record the event.
				if f.guardrailEvents != nil {
					_ = f.recordGuardrailEvent(cr.guardrail, apiKeyID, msg.Content)
				}
				return nil, &GuardrailRejectedError{Pattern: cr.guardrail.Pattern}

			case "replace":
				result[i].Content = cr.re.ReplaceAllString(result[i].Content, cr.guardrail.ReplaceBy)
			}
		}
	}

	return result, nil
}

func (f *Forwarder) recordGuardrailEvent(g models.Guardrail, apiKeyID, inputText string) error {
	_, err := f.guardrailEvents.Record(g.ID, apiKeyID, g.Pattern, g.Mode, inputText)
	return err
}

// usagePayload is the token-usage portion of a non-streaming response.
type usagePayload struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type completionResponse struct {
	Usage *usagePayload `json:"usage"`
}

// streamChunkUsage represents the usage field that may appear in a streaming chunk.
type streamChunkUsage struct {
	Usage *usagePayload `json:"usage"`
}

// ForwardChatCompletion handles both streaming and non-streaming chat completions.
func (f *Forwarder) ForwardChatCompletion(
	w http.ResponseWriter,
	r *http.Request,
	provider *models.Provider,
	apiKey *models.APIKey,
) error {
	// Read and parse the request body to inspect model/stream.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	defer r.Body.Close()

	var req chatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return fmt.Errorf("parse request: %w", err)
	}

	// Apply guardrails to all message contents.
	if f.guardrails != nil {
		filtered, err := f.applyGuardrails(req.Messages, apiKey.ID)
		if err != nil {
			return err // will be *GuardrailRejectedError for reject mode
		}
		// If messages were modified (replace mode), re-serialize the body.
		req.Messages = filtered
		// Re-marshal the full request body with updated messages.
		var rawBody map[string]json.RawMessage
		if err := json.Unmarshal(body, &rawBody); err == nil {
			if updatedMessages, err := json.Marshal(filtered); err == nil {
				rawBody["messages"] = updatedMessages
				if newBody, err := json.Marshal(rawBody); err == nil {
					body = newBody
				}
			}
		}
	}

	// Build upstream request.
	upstreamURL := strings.TrimRight(provider.BaseURL, "/") + "/v1/chat/completions"
	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create upstream request: %w", err)
	}
	upReq.Header.Set("Content-Type", "application/json")
	upReq.Header.Set("Authorization", "Bearer "+provider.APIKey)

	// Copy Accept header if present.
	if accept := r.Header.Get("Accept"); accept != "" {
		upReq.Header.Set("Accept", accept)
	}

	resp, err := f.client.Do(upReq)
	if err != nil {
		return fmt.Errorf("upstream request: %w", err)
	}
	defer resp.Body.Close()

	if req.Stream {
		return f.handleStreaming(w, resp, req.Model, provider, apiKey)
	}
	return f.handleNonStreaming(w, resp, req.Model, provider, apiKey)
}

func (f *Forwarder) handleNonStreaming(
	w http.ResponseWriter,
	resp *http.Response,
	model string,
	provider *models.Provider,
	apiKey *models.APIKey,
) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read upstream response: %w", err)
	}

	// Copy headers.
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(body)

	// Track usage (best-effort, don't fail the request).
	if resp.StatusCode == http.StatusOK {
		var cr completionResponse
		if json.Unmarshal(body, &cr) == nil && cr.Usage != nil {
			_ = f.usage.Record(
				apiKey.ID, provider.ID, model,
				cr.Usage.PromptTokens, cr.Usage.CompletionTokens, cr.Usage.TotalTokens,
			)
		}
	}

	return nil
}

func (f *Forwarder) handleStreaming(
	w http.ResponseWriter,
	resp *http.Response,
	model string,
	provider *models.Provider,
	apiKey *models.APIKey,
) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported by response writer")
	}

	// Copy headers.
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(resp.StatusCode)

	var totalUsage usagePayload

	scanner := bufio.NewScanner(resp.Body)
	// Increase scanner buffer for large chunks.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		// Try to extract usage from data chunks.
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data != "[DONE]" {
				var chunk streamChunkUsage
				if json.Unmarshal([]byte(data), &chunk) == nil && chunk.Usage != nil {
					totalUsage = *chunk.Usage
				}
			}
		}

		fmt.Fprintf(w, "%s\n", line)
		flusher.Flush()
	}

	// Record usage if we got any.
	if totalUsage.TotalTokens > 0 {
		_ = f.usage.Record(
			apiKey.ID, provider.ID, model,
			totalUsage.PromptTokens, totalUsage.CompletionTokens, totalUsage.TotalTokens,
		)
	}

	return scanner.Err()
}

// ForwardListModels proxies a GET /v1/models request to the upstream provider.
func (f *Forwarder) ForwardListModels(
	w http.ResponseWriter,
	r *http.Request,
	provider *models.Provider,
) error {
	upstreamURL := strings.TrimRight(provider.BaseURL, "/") + "/v1/models"
	upReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, upstreamURL, nil)
	if err != nil {
		return fmt.Errorf("create upstream request: %w", err)
	}
	upReq.Header.Set("Authorization", "Bearer "+provider.APIKey)

	resp, err := f.client.Do(upReq)
	if err != nil {
		return fmt.Errorf("upstream request: %w", err)
	}
	defer resp.Body.Close()

	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}
