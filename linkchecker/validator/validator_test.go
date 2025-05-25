package validator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateLinks_HTTP(t *testing.T) {
	// HTTP-Server für verschiedene Status
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/redirect":
			http.Redirect(w, r, "/ok", http.StatusFound)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	links := []string{
		ts.URL + "/ok",
		ts.URL + "/redirect",
		ts.URL + "/notfound",
	}
	results := ValidateLinks(links, "")

	if !results[0].Valid {
		t.Errorf("expected %s to be valid", links[0])
	}
	if results[1].Valid {
		t.Errorf("expected %s to be invalid (redirect)", links[1])
	}
	if results[2].Valid {
		t.Errorf("expected %s to be invalid (not found)", links[2])
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

	if !results[0].Valid {
		t.Errorf("expected %s to be valid", links[0])
	}
	if results[1].Valid {
		t.Errorf("expected %s to be invalid", links[1])
	}
}
