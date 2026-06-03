# finna-go

Official Go SDK for the [Finna](https://finna.sh) managed search API — keyword, vector, and hybrid retrieval over EU-sovereign infrastructure.

## Install

```sh
go get github.com/finnahq/finna-go
```

Requires Go 1.21+. No external dependencies — standard library only.

## Quickstart

```go
client := finna.New("sk_live_...")

// Create an index
idx, err := client.CreateIndex(ctx, finna.IndexSettings{
    Name:     "products",
    Language: "en",
    Fields: map[string]finna.FieldConfig{
        "title": {Type: finna.FieldTypeText, Searchable: ptr(true)},
        "price": {Type: finna.FieldTypeNumber, Filterable: ptr(true), Sortable: ptr(true)},
    },
})

// Ingest documents
client.UpsertDocuments(ctx, idx.ID, []finna.Document{
    {"id": "1", "title": "Widget", "price": 9.99},
})

// Search
results, err := client.Search(ctx, idx.ID, finna.SearchRequest{
    Query: "widget", Limit: 10,
})
fmt.Printf("%d hits\n", results.Total)
```

## Client options

```go
client := finna.New(apiKey,
    finna.WithBaseURL("http://localhost:7700"), // local dev
    finna.WithTimeout(10*time.Second),
    finna.WithHTTPClient(myHTTPClient),
)
```

## API methods

| Method | Description |
|---|---|
| `ListIndexes(ctx)` | List all indexes |
| `CreateIndex(ctx, settings)` | Create an index |
| `GetIndex(ctx, indexID)` | Get an index by ID |
| `DeleteIndex(ctx, indexID)` | Delete an index and all documents |
| `GetIndexStats(ctx, indexID)` | Document count and storage bytes |
| `UpsertDocuments(ctx, indexID, docs)` | Insert or replace documents |
| `PatchDocument(ctx, indexID, docID, patch)` | Shallow-merge fields into a document |
| `DeleteDocument(ctx, indexID, docID)` | Delete one document |
| `ClearDocuments(ctx, indexID)` | Delete all documents (keep index) |
| `Search(ctx, indexID, req)` | Text / vector / hybrid search |
| `Analyze(ctx, text, language)` | Tokenize text with a language analyzer |
| `AnalyzeIndex(ctx, indexID, text)` | Tokenize text with an index's analyzer |

## Error handling

All API errors are returned as `*FinnaError`:

```go
results, err := client.Search(ctx, indexID, req)
var fe *finna.FinnaError
if errors.As(err, &fe) {
    fmt.Printf("HTTP %d: %s\n", fe.Status, fe.Message)
}
```

The client automatically retries network errors, HTTP 429, and HTTP 5xx responses (2 retries, 500ms / 1s backoff).

## Vector search example

```go
results, err := client.Search(ctx, indexID, finna.SearchRequest{
    Vector:      []float64{0.12, 0.98, /* ... 1536 dims */},
    VectorField: "embedding",
    Mode:        finna.SearchModeVector,
    Limit:       20,
})
```

## Hybrid search example

```go
results, err := client.Search(ctx, indexID, finna.SearchRequest{
    Query:       "fast database",
    Vector:      queryEmbedding,
    VectorField: "embedding",
    Mode:        finna.SearchModeHybrid,
    Filter: []finna.FilterCondition{
        {Field: "year", Op: finna.FilterOpGte, Value: 2020},
    },
    Highlight: true,
    Limit:     10,
})
```

## Running the quickstart example

```sh
export FINNA_API_KEY=sk_live_...
export FINNA_BASE_URL=http://localhost:7700  # optional
go run ./examples/quickstart
```

## License

Proprietary — see [LICENSE](LICENSE).
