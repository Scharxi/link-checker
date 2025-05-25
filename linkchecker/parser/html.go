package parser

import (
	"bytes"
	"io"
	"os"

	"golang.org/x/net/html"
)

// ExtractLinksFromHTMLFile liest eine HTML-Datei ein und gibt alle gefundenen Links (href) zurück.
func ExtractLinksFromHTMLFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return ExtractLinksFromHTML(content), nil
}

// ExtractLinksFromHTML extrahiert alle Links (href) aus HTML-Content.
func ExtractLinksFromHTML(content []byte) []string {
	var links []string
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return links
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return links
}
