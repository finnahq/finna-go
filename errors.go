package finna

import "fmt"

// FinnaError is returned whenever the Finna API responds with a non-2xx status
// code. It carries the HTTP status and the human-readable message from the
// response body.
type FinnaError struct {
	Status  int
	Message string
}

func (e *FinnaError) Error() string {
	return fmt.Sprintf("finna: HTTP %d: %s", e.Status, e.Message)
}

// errBody is the JSON shape of every Finna error response: {"error": "..."}
type errBody struct {
	Error string `json:"error"`
}
