package parser

import (
	"os"
	"testing"
)

func TestExtractLinksFromHTML(t *testing.T) {
	htmlContent := []byte(`<html><body><a href="https://example.com">Example</a><a href='test.html'>Test</a></body></html>`)
	links := ExtractLinksFromHTML(htmlContent)

	want := []string{"https://example.com", "test.html"}
	if len(links) != len(want) {
		t.Fatalf("expected %d links, got %d", len(want), len(links))
	}
	for i, link := range want {
		if links[i] != link {
			t.Errorf("expected link %q, got %q", link, links[i])
		}
	}
}

func TestExtractLinksFromHTMLFile(t *testing.T) {
	const content = `<html><body><a href="https://golang.org">Go</a></body></html>`
	f, err := os.CreateTemp("", "test-*.html")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("failed to write to temp file: %v", err)
	}
	f.Close()

	links, err := ExtractLinksFromHTMLFile(f.Name())
	if err != nil {
		t.Fatalf("ExtractLinksFromHTMLFile error: %v", err)
	}
	if len(links) != 1 || links[0] != "https://golang.org" {
		t.Errorf("expected [https://golang.org], got %v", links)
	}
}
