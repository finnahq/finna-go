// Command quickstart demonstrates: create index → ingest → search.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/finnahq/finna-go"
)

func ptr[T any](v T) *T { return &v }

func main() {
	apiKey := os.Getenv("FINNA_API_KEY")
	if apiKey == "" {
		log.Fatal("FINNA_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("FINNA_BASE_URL") // optional; defaults to https://api.finna.sh
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opts := []finna.Option{}
	if baseURL != "" {
		opts = append(opts, finna.WithBaseURL(baseURL))
	}
	client := finna.New(apiKey, opts...)

	// ── 1. Create index ─────────────────────────────────────────────────────
	fmt.Println("Creating index...")
	idx, err := client.CreateIndex(ctx, finna.IndexSettings{
		Name:     "books",
		Language: "en",
		Fields: map[string]finna.FieldConfig{
			"title":  {Type: finna.FieldTypeText, Searchable: ptr(true), Filterable: ptr(true)},
			"author": {Type: finna.FieldTypeText, Searchable: ptr(true)},
			"year":   {Type: finna.FieldTypeNumber, Filterable: ptr(true), Sortable: ptr(true)},
		},
	})
	if err != nil {
		var fe *finna.FinnaError
		if errors.As(err, &fe) {
			log.Fatalf("API error %d: %s", fe.Status, fe.Message)
		}
		log.Fatal(err)
	}
	fmt.Printf("Created index %q (id: %s)\n", idx.Name, idx.ID)

	// ── 2. Ingest documents ─────────────────────────────────────────────────
	fmt.Println("Ingesting documents...")
	upsert, err := client.UpsertDocuments(ctx, idx.ID, []finna.Document{
		{"id": "1", "title": "The Go Programming Language", "author": "Donovan & Kernighan", "year": 2015},
		{"id": "2", "title": "Designing Data-Intensive Applications", "author": "Martin Kleppmann", "year": 2017},
		{"id": "3", "title": "Clean Code", "author": "Robert C. Martin", "year": 2008},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Indexed %d documents\n", upsert.Indexed)

	// ── 3. Search ───────────────────────────────────────────────────────────
	fmt.Println("Searching for \"programming\"...")
	results, err := client.Search(ctx, idx.ID, finna.SearchRequest{
		Query:     "programming",
		Limit:     5,
		Highlight: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d hits (total_type: %s)\n", results.Total, results.TotalType)
	for i, hit := range results.Hits {
		b, _ := json.MarshalIndent(hit, "  ", "  ")
		fmt.Printf("  [%d] %s\n", i+1, b)
	}

	// ── 4. Stats ────────────────────────────────────────────────────────────
	stats, err := client.GetIndexStats(ctx, idx.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Index stats: %d documents, %d bytes\n", stats.Documents, stats.StorageBytes)

	// ── 5. Cleanup ──────────────────────────────────────────────────────────
	fmt.Println("Deleting index...")
	if err := client.DeleteIndex(ctx, idx.ID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Done.")
}
