package finna

// FieldType enumerates the supported field types.
type FieldType string

const (
	FieldTypeText   FieldType = "text"
	FieldTypeNumber FieldType = "number"
	FieldTypeVector FieldType = "vector"
)

// FilterOp enumerates the supported filter operators.
type FilterOp string

const (
	FilterOpEq  FilterOp = "eq"
	FilterOpLt  FilterOp = "lt"
	FilterOpLte FilterOp = "lte"
	FilterOpGt  FilterOp = "gt"
	FilterOpGte FilterOp = "gte"
)

// SearchMode enumerates the retrieval modes.
type SearchMode string

const (
	SearchModeText   SearchMode = "text"
	SearchModeVector SearchMode = "vector"
	SearchModeHybrid SearchMode = "hybrid"
)

// TotalType describes whether the total hit count is exact or an approximation.
type TotalType string

const (
	TotalTypeExact       TotalType = "exact"
	TotalTypeApproximate TotalType = "approximate"
)

// FieldConfig describes the schema for a single index field.
type FieldConfig struct {
	// Type is required: "text", "number", or "vector".
	Type FieldType `json:"type"`
	// Searchable enables BM25 full-text search (text fields only).
	Searchable *bool `json:"searchable,omitempty"`
	// Filterable enables exact/range filtering (text and number fields).
	Filterable *bool `json:"filterable,omitempty"`
	// Sortable enables sorting (number fields only).
	Sortable *bool `json:"sortable,omitempty"`
	// Dims is the vector dimensionality (vector fields only; required for vector).
	Dims *int `json:"dims,omitempty"`
	// Metric is the similarity metric (vector fields only; default: "cosine").
	Metric *string `json:"metric,omitempty"`
}

// IndexSettings is the request body for CreateIndex.
type IndexSettings struct {
	// Name is the display label (1–64 chars).
	Name string `json:"name"`
	// Language is the fixed language code (e.g. "en", "de") or "auto" (default).
	Language string `json:"language,omitempty"`
	// Decompound splits German compounds at index + query time (default true).
	Decompound *bool `json:"decompound,omitempty"`
	// Fields is the field schema map.
	Fields map[string]FieldConfig `json:"fields,omitempty"`
}

// Index is an IndexSettings with an assigned server-side ID.
type Index struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Language   string                 `json:"language,omitempty"`
	Decompound *bool                  `json:"decompound,omitempty"`
	Fields     map[string]FieldConfig `json:"fields,omitempty"`
}

// IndexStats holds document count and storage usage for an index.
type IndexStats struct {
	Documents    int64 `json:"documents"`
	StorageBytes int64 `json:"storage_bytes"`
}

// Document is a free-form map that must contain a string "id" key.
// Values may be strings (text fields), floats (number fields), or []float64
// (vector fields).
type Document map[string]any

// FilterCondition is a single predicate in a search filter.
type FilterCondition struct {
	Field string   `json:"field"`
	Op    FilterOp `json:"op"`
	Value any      `json:"value"`
}

// SearchRequest is the payload for the Search method.
type SearchRequest struct {
	// Query is the text query string. Empty + no Vector = match-all.
	Query string `json:"query,omitempty"`
	// Vector is the query embedding for vector/hybrid search.
	Vector []float64 `json:"vector,omitempty"`
	// VectorField names the vector field to search (omit if there is only one).
	VectorField string `json:"vector_field,omitempty"`
	// Mode is "text", "vector", or "hybrid". Inferred from inputs when omitted.
	Mode SearchMode `json:"mode,omitempty"`
	// Filter is a list of filter conditions (all must match).
	Filter []FilterCondition `json:"filter,omitempty"`
	// Sort is a sortable number field name, optionally prefixed with "-" for descending.
	Sort string `json:"sort,omitempty"`
	// Limit is the page size (default 10, max 200).
	Limit int `json:"limit,omitempty"`
	// Offset is the page offset (default 0).
	Offset int `json:"offset,omitempty"`
	// Prefix enables search-as-you-type (matches last token as a prefix).
	Prefix bool `json:"prefix,omitempty"`
	// Fuzzy enables typo tolerance (Levenshtein on query tokens).
	Fuzzy bool `json:"fuzzy,omitempty"`
	// Highlight returns per-field HTML snippets in _highlights.
	Highlight bool `json:"highlight,omitempty"`
}

// Hit is a matched document returned by Search.
// It always contains "_score". When Highlight was requested it also contains
// "_highlights" with per-field HTML strings. Vector fields are omitted.
type Hit map[string]any

// SearchResults is returned by the Search method.
type SearchResults struct {
	Hits      []Hit     `json:"hits"`
	Total     int       `json:"total"`
	TotalType TotalType `json:"total_type"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// CreateIndexResponse is the response from CreateIndex.
type CreateIndexResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UpsertDocumentsResponse is the response from UpsertDocuments.
type UpsertDocumentsResponse struct {
	Indexed int `json:"indexed"`
}

// PatchDocumentResponse is the response from PatchDocument.
type PatchDocumentResponse struct {
	Ok       bool     `json:"ok"`
	Document Document `json:"document"`
}

// AnalyzeResponse is the response from Analyze and AnalyzeIndex.
type AnalyzeResponse struct {
	Tokens []string `json:"tokens"`
}
