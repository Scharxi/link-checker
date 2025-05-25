package parser

import (
	"io"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ExtractLinksFromFile liest eine Markdown-Datei ein und gibt alle gefundenen Links zurück.
func ExtractLinksFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return ExtractLinks(content), nil
}

// ExtractLinks extrahiert alle Links aus Markdown-Content.
func ExtractLinks(content []byte) []string {
	var links []string
	md := goldmark.New()
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if link, ok := n.(*ast.Link); ok && entering {
			links = append(links, string(link.Destination))
		}
		return ast.WalkContinue, nil
	})

	return links
}
