package validator

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LinkStatus struct {
	Link       string
	Valid      bool
	Reason     string
	StatusCode int
}

// ValidateLinks prüft, ob Links erreichbar sind (HTTP) oder existieren (Dateipfad).
func ValidateLinks(links []string, basePath string) []LinkStatus {
	return ValidateLinksWithTimeout(links, basePath, 30*time.Second)
}

// ValidateLinksWithTimeout prüft Links mit einem konfigurierbaren Timeout.
func ValidateLinksWithTimeout(links []string, basePath string, timeout time.Duration) []LinkStatus {
	return ValidateLinksAsync(links, basePath, timeout, 10) // Default 10 concurrent workers
}

// ValidateLinksAsync prüft Links asynchron mit konfigurierbarer Anzahl von Workern.
func ValidateLinksAsync(links []string, basePath string, timeout time.Duration, maxWorkers int) []LinkStatus {
	if len(links) == 0 {
		return []LinkStatus{}
	}

	// Channels für die Kommunikation
	linkChan := make(chan string, len(links))
	resultChan := make(chan LinkStatus, len(links))

	// Worker-Pool starten
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go worker(linkChan, resultChan, basePath, timeout, &wg)
	}

	// Links in den Channel senden
	go func() {
		for _, link := range links {
			linkChan <- link
		}
		close(linkChan)
	}()

	// Warten bis alle Worker fertig sind
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Ergebnisse sammeln
	results := make([]LinkStatus, 0, len(links))
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

func worker(linkChan <-chan string, resultChan chan<- LinkStatus, basePath string, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()

	for link := range linkChan {
		var status LinkStatus

		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			ok, reason, statusCode := checkHTTP(link, timeout)
			status = LinkStatus{
				Link:       link,
				Valid:      ok,
				Reason:     reason,
				StatusCode: statusCode,
			}
		} else {
			ok, reason := checkFile(basePath, link)
			status = LinkStatus{
				Link:       link,
				Valid:      ok,
				Reason:     reason,
				StatusCode: 0,
			}
		}

		resultChan <- status
	}
}

func checkHTTP(url string, timeout time.Duration) (bool, string, int) {
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Try HEAD request first (faster)
	resp, err := client.Head(url)
	if err != nil {
		// If HEAD fails, try GET request (some servers don't support HEAD)
		resp, err = client.Get(url)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				return false, "Request timeout", 0
			}
			return false, err.Error(), 0
		}
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx status codes as valid
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, "", resp.StatusCode
	}

	return false, resp.Status, resp.StatusCode
}

func checkFile(basePath, relPath string) (bool, string) {
	var fullPath string

	if filepath.IsAbs(relPath) {
		fullPath = relPath
	} else {
		fullPath = filepath.Join(basePath, relPath)
	}

	_, err := os.Stat(fullPath)
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}
