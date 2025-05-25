package parser

import (
	"os"
	"testing"
)

func TestExtractLinks(t *testing.T) {
	md := []byte(`This is a [link](https://example.com) and another [ref](ref.md).`)
	links := ExtractLinks(md)

	want := []string{"https://example.com", "ref.md"}
	if len(links) != len(want) {
		t.Fatalf("expected %d links, got %d", len(want), len(links))
	}
	for i, link := range want {
		if links[i] != link {
			t.Errorf("expected link %q, got %q", link, links[i])
		}
	}
}

func TestExtractLinksFromFile(t *testing.T) {
	const content = `# Title\n\nA [test link](https://golang.org) in markdown.`
	f, err := os.CreateTemp("", "test-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		t.Fatalf("failed to write to temp file: %v", err)
	}
	f.Close()

	links, err := ExtractLinksFromFile(f.Name())
	if err != nil {
		t.Fatalf("ExtractLinksFromFile error: %v", err)
	}
	if len(links) != 1 || links[0] != "https://golang.org" {
		t.Errorf("expected [https://golang.org], got %v", links)
	}
}
