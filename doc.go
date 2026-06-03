// Package finna provides an idiomatic Go client for the Finna search API.
//
// Finna is a managed vector/hybrid search service. It supports keyword (BM25),
// vector, and hybrid retrieval over schematized indexes.
//
// # Quickstart
//
//	client := finna.New("sk_live_...")
//
//	// Create an index
//	idx, err := client.CreateIndex(ctx, finna.IndexSettings{
//	    Name:     "products",
//	    Language: "en",
//	    Fields: map[string]finna.FieldConfig{
//	        "title":     {Type: "text", Searchable: ptr(true)},
//	        "price":     {Type: "number", Filterable: ptr(true), Sortable: ptr(true)},
//	        "embedding": {Type: "vector", Dims: ptr(1536)},
//	    },
//	})
//
//	// Ingest documents
//	result, err := client.UpsertDocuments(ctx, idx.ID, []finna.Document{
//	    {"id": "1", "title": "Widget", "price": 9.99},
//	})
//
//	// Search
//	hits, err := client.Search(ctx, idx.ID, finna.SearchRequest{
//	    Query: "widget",
//	    Limit: 10,
//	})
//
// # Error handling
//
// All API errors are returned as *FinnaError, which carries the HTTP status code
// and the error message from the server.
//
//	hits, err := client.Search(ctx, indexID, req)
//	var fe *finna.FinnaError
//	if errors.As(err, &fe) {
//	    fmt.Printf("HTTP %d: %s\n", fe.Status, fe.Message)
//	}
package finna
