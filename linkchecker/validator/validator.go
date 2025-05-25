package validator

import (
	"net/http"
	"os"
	"strings"
	"time"
)

type LinkStatus struct {
	Link   string
	Valid  bool
	Reason string
}

// ValidateLinks prüft, ob Links erreichbar sind (HTTP) oder existieren (Dateipfad).
func ValidateLinks(links []string, basePath string) []LinkStatus {
	var results []LinkStatus
	for _, link := range links {
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			ok, reason := checkHTTP(link)
			results = append(results, LinkStatus{Link: link, Valid: ok, Reason: reason})
		} else {
			ok, reason := checkFile(basePath, link)
			results = append(results, LinkStatus{Link: link, Valid: ok, Reason: reason})
		}
	}
	return results
}

func checkHTTP(url string) (bool, string) {
	client := http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Don't follow redirects - treat them as invalid
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Head(url)
	if err != nil {
		return false, err.Error()
	}
	defer resp.Body.Close()

	// Only consider 2xx status codes as valid
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true, ""
	}

	// Redirects (3xx) are considered invalid
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return false, "redirect not allowed"
	}

	return false, resp.Status
}

func checkFile(basePath, relPath string) (bool, string) {
	fullPath := relPath
	if !strings.HasPrefix(relPath, "/") {
		fullPath = basePath + "/" + relPath
	}
	_, err := os.Stat(fullPath)
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}
