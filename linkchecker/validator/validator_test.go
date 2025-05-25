package validator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateLinks_HTTP(t *testing.T) {
	// Simple HTTP server that properly handles HEAD and GET requests
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	links := []string{
		ts.URL + "/ok",
		ts.URL + "/notfound",
	}
	results := ValidateLinks(links, "")

	// Debug output
	for i, result := range results {
		t.Logf("Link %d: %s -> Valid: %v, StatusCode: %d, Reason: %s",
			i, result.Link, result.Valid, result.StatusCode, result.Reason)
	}

	// Create a map for easier testing since async results may be in different order
	resultMap := make(map[string]LinkStatus)
	for _, result := range results {
		resultMap[result.Link] = result
	}

	// Test that 200 OK is valid
	okResult := resultMap[ts.URL+"/ok"]
	if !okResult.Valid {
		t.Errorf("expected %s to be valid, got: %s (status: %d)", ts.URL+"/ok", okResult.Reason, okResult.StatusCode)
	}

	// Test that 404 is invalid
	notFoundResult := resultMap[ts.URL+"/notfound"]
	if notFoundResult.Valid {
		t.Errorf("expected %s to be invalid (not found), got: %s (status: %d)", ts.URL+"/notfound", notFoundResult.Reason, notFoundResult.StatusCode)
	}
}

func TestValidateLinks_File(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.md")
	if err := os.WriteFile(file, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	links := []string{
		"test.md",     // existiert
		"notfound.md", // existiert nicht
	}
	results := ValidateLinks(links, dir)

	// Create a map for easier testing since async results may be in different order
	resultMap := make(map[string]LinkStatus)
	for _, result := range results {
		resultMap[result.Link] = result
	}

	// Test that existing file is valid
	validResult := resultMap["test.md"]
	if !validResult.Valid {
		t.Errorf("expected test.md to be valid, got: %s", validResult.Reason)
	}

	// Test that non-existing file is invalid
	invalidResult := resultMap["notfound.md"]
	if invalidResult.Valid {
		t.Errorf("expected notfound.md to be invalid, got: %s", invalidResult.Reason)
	}
}
