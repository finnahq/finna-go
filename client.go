package finna

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.finna.sh"
	defaultTimeout = 30 * time.Second
	maxRetries     = 2
)

// Client is the Finna API client. Create one with New.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// Option is a functional option for New.
type Option func(*Client)

// WithBaseURL overrides the default API base URL (https://api.finna.sh).
// Useful for local development: finna.WithBaseURL("http://localhost:7700").
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient replaces the default *http.Client. Timeout on this client
// takes precedence over WithTimeout.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithTimeout sets the HTTP client timeout (default 30s).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.httpClient.Timeout = d }
}

// New creates a Finna API client. apiKey is required (e.g. "sk_live_…").
func New(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: defaultTimeout},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// ---------------------------------------------------------------------------
// Indexes
// ---------------------------------------------------------------------------

// ListIndexes returns all indexes visible to this API key.
func (c *Client) ListIndexes(ctx context.Context) ([]Index, error) {
	var result struct {
		Indexes []Index `json:"indexes"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/indexes", nil, &result); err != nil {
		return nil, err
	}
	return result.Indexes, nil
}

// CreateIndex creates a new index with the given settings.
func (c *Client) CreateIndex(ctx context.Context, settings IndexSettings) (CreateIndexResponse, error) {
	var result CreateIndexResponse
	if err := c.do(ctx, http.MethodPost, "/v1/indexes", settings, &result); err != nil {
		return CreateIndexResponse{}, err
	}
	return result, nil
}

// GetIndex fetches a single index by ID.
func (c *Client) GetIndex(ctx context.Context, indexID string) (Index, error) {
	var result Index
	if err := c.do(ctx, http.MethodGet, "/v1/indexes/"+indexID, nil, &result); err != nil {
		return Index{}, err
	}
	return result, nil
}

// DeleteIndex deletes an index and all its documents.
func (c *Client) DeleteIndex(ctx context.Context, indexID string) error {
	return c.do(ctx, http.MethodDelete, "/v1/indexes/"+indexID, nil, nil)
}

// GetIndexStats returns document count and storage size for an index.
func (c *Client) GetIndexStats(ctx context.Context, indexID string) (IndexStats, error) {
	var result IndexStats
	if err := c.do(ctx, http.MethodGet, "/v1/indexes/"+indexID+"/stats", nil, &result); err != nil {
		return IndexStats{}, err
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Documents
// ---------------------------------------------------------------------------

// UpsertDocuments inserts or replaces documents by their "id" field.
func (c *Client) UpsertDocuments(ctx context.Context, indexID string, docs []Document) (UpsertDocumentsResponse, error) {
	body := map[string]any{"documents": docs}
	var result UpsertDocumentsResponse
	if err := c.do(ctx, http.MethodPost, "/v1/indexes/"+indexID+"/documents", body, &result); err != nil {
		return UpsertDocumentsResponse{}, err
	}
	return result, nil
}

// PatchDocument partially updates a document (shallow merge). A null value in
// patch clears the corresponding field; the "id" field is immutable.
func (c *Client) PatchDocument(ctx context.Context, indexID, documentID string, patch map[string]any) (PatchDocumentResponse, error) {
	var result PatchDocumentResponse
	path := fmt.Sprintf("/v1/indexes/%s/documents/%s", indexID, documentID)
	if err := c.do(ctx, http.MethodPatch, path, patch, &result); err != nil {
		return PatchDocumentResponse{}, err
	}
	return result, nil
}

// DeleteDocument deletes a single document by ID.
func (c *Client) DeleteDocument(ctx context.Context, indexID, documentID string) error {
	path := fmt.Sprintf("/v1/indexes/%s/documents/%s", indexID, documentID)
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// ClearDocuments deletes all documents in an index (the index itself remains).
func (c *Client) ClearDocuments(ctx context.Context, indexID string) error {
	return c.do(ctx, http.MethodDelete, "/v1/indexes/"+indexID+"/documents", nil, nil)
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// Search executes a text, vector, or hybrid search.
func (c *Client) Search(ctx context.Context, indexID string, req SearchRequest) (SearchResults, error) {
	var result SearchResults
	if err := c.do(ctx, http.MethodPost, "/v1/indexes/"+indexID+"/search", req, &result); err != nil {
		return SearchResults{}, err
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Analyze
// ---------------------------------------------------------------------------

// Analyze tokenizes text using the given language's analyzer.
// Pass language "" or "auto" to let the server detect the language.
func (c *Client) Analyze(ctx context.Context, text, language string) ([]string, error) {
	body := map[string]string{"text": text}
	if language != "" {
		body["language"] = language
	}
	var result AnalyzeResponse
	if err := c.do(ctx, http.MethodPost, "/v1/_analyze", body, &result); err != nil {
		return nil, err
	}
	return result.Tokens, nil
}

// AnalyzeIndex tokenizes text using the analyzer configured for the given index.
func (c *Client) AnalyzeIndex(ctx context.Context, indexID, text string) ([]string, error) {
	body := map[string]string{"text": text}
	var result AnalyzeResponse
	if err := c.do(ctx, http.MethodPost, "/v1/indexes/"+indexID+"/_analyze", body, &result); err != nil {
		return nil, err
	}
	return result.Tokens, nil
}

// ---------------------------------------------------------------------------
// Internal HTTP transport
// ---------------------------------------------------------------------------

// do executes an HTTP request, decoding the JSON response into out (may be nil).
// It retries up to maxRetries times on network errors, 429, and 5xx responses.
func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 500ms, 1000ms
			wait := time.Duration(math.Pow(2, float64(attempt-1))*500) * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
		}

		var reqBody io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return fmt.Errorf("finna: marshal request: %w", err)
			}
			reqBody = bytes.NewReader(b)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
		if err != nil {
			return fmt.Errorf("finna: build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("finna: network error: %w", err)
			continue // retry on network error
		}

		// Decide whether to retry before consuming the body.
		shouldRetry := resp.StatusCode == http.StatusTooManyRequests ||
			resp.StatusCode >= http.StatusInternalServerError

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if out != nil {
				if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
					resp.Body.Close()
					return fmt.Errorf("finna: decode response: %w", err)
				}
			}
			resp.Body.Close()
			return nil
		}

		// Non-2xx: parse the error body.
		var eb errBody
		_ = json.NewDecoder(resp.Body).Decode(&eb)
		resp.Body.Close()

		fe := &FinnaError{Status: resp.StatusCode, Message: eb.Error}
		if shouldRetry && attempt < maxRetries {
			lastErr = fe
			continue
		}
		return fe
	}
	return lastErr
}
