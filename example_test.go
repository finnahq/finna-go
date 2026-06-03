package finna_test

import (
	"context"
	"fmt"
	"log"

	"github.com/finnahq/finna-go"
)

func ptr[T any](v T) *T { return &v }

func Example() {
	ctx := context.Background()

	client := finna.New("sk_live_your_api_key_here",
		finna.WithBaseURL("http://localhost:7700"),
	)

	// Create an index
	idx, err := client.CreateIndex(ctx, finna.IndexSettings{
		Name:     "products",
		Language: "en",
		Fields: map[string]finna.FieldConfig{
			"title": {Type: finna.FieldTypeText, Searchable: ptr(true)},
			"price": {Type: finna.FieldTypeNumber, Filterable: ptr(true), Sortable: ptr(true)},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Ingest documents
	_, err = client.UpsertDocuments(ctx, idx.ID, []finna.Document{
		{"id": "1", "title": "Ergonomic Keyboard", "price": 149.99},
		{"id": "2", "title": "Mechanical Keyboard", "price": 89.99},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Search
	results, err := client.Search(ctx, idx.ID, finna.SearchRequest{
		Query: "keyboard",
		Limit: 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d hits\n", results.Total)
}
