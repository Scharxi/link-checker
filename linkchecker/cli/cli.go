package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"bxfferoverflow.me/link-checker/linkchecker/parser"
	"bxfferoverflow.me/link-checker/linkchecker/validator"
	"github.com/spf13/cobra"
)

// Config holds all CLI configuration options
type Config struct {
	Recursive   bool
	IgnoreList  []string
	Timeout     time.Duration
	OnlyDead    bool
	Format      string
	InputPaths  []string
	InputURLs   []string
	IgnoreRegex []*regexp.Regexp
	Workers     int
	Debug       bool
}

// Result represents a link check result
type Result struct {
	URL        string `json:"url"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	Error      string `json:"error,omitempty"`
	Source     string `json:"source"`
	Line       int    `json:"line,omitempty"`
}

// Output represents the final output structure
type Output struct {
	Summary struct {
		Total    int    `json:"total"`
		Valid    int    `json:"valid"`
		Invalid  int    `json:"invalid"`
		Duration string `json:"duration"`
	} `json:"summary"`
	Results []Result `json:"results"`
}

var (
	config  Config
	rootCmd = &cobra.Command{
		Use:   "linkchecker [paths or URLs...]",
		Short: "A fast and reliable link checker for markdown files and web pages",
		Long: `Link Checker is a command-line tool that validates links in markdown files
and checks web pages for dead links. You can provide file paths, directories,
or direct URLs to check.`,
		Example: `  # Check markdown files
  linkchecker README.md
  linkchecker --recursive ./docs
  
  # Check web pages for dead links
  linkchecker https://example.com
  linkchecker https://github.com/user/repo
  
  # Mixed usage
  linkchecker README.md https://example.com ./docs
  
  # Advanced options
  linkchecker --ignore="example.com,test.local" --timeout=10s https://mysite.com
  linkchecker --only-dead --format=json https://example.com`,
		RunE: runLinkChecker,
	}
)

func init() {
	// Define flags
	rootCmd.Flags().BoolVarP(&config.Recursive, "recursive", "r", false,
		"Recursively scan directories for markdown files")

	rootCmd.Flags().StringSliceVar(&config.IgnoreList, "ignore", []string{},
		"Comma-separated list of domains or regex patterns to ignore (e.g., 'example.com,*.test.local')")

	rootCmd.Flags().DurationVar(&config.Timeout, "timeout", 30*time.Second,
		"HTTP request timeout (e.g., 10s, 1m, 30s)")

	rootCmd.Flags().BoolVar(&config.OnlyDead, "only-dead", false,
		"Only show dead/broken links in output")

	rootCmd.Flags().StringVar(&config.Format, "format", "text",
		"Output format: 'text' or 'json'")

	rootCmd.Flags().IntVar(&config.Workers, "workers", 10,
		"Number of concurrent workers for link validation (default: 10)")

	rootCmd.Flags().BoolVar(&config.Debug, "debug", false,
		"Enable debug output")

	// Add help examples
	rootCmd.SetHelpTemplate(getHelpTemplate())
}

func runLinkChecker(cmd *cobra.Command, args []string) error {
	// Set default if no arguments provided
	if len(args) == 0 {
		config.InputPaths = []string{"."}
	} else {
		// Separate URLs from file paths
		config.InputPaths = []string{}
		config.InputURLs = []string{}

		for _, arg := range args {
			if isURL(arg) {
				config.InputURLs = append(config.InputURLs, arg)
			} else {
				config.InputPaths = append(config.InputPaths, arg)
			}
		}
	}

	// Validate format
	if config.Format != "text" && config.Format != "json" {
		return fmt.Errorf("invalid format '%s': must be 'text' or 'json'", config.Format)
	}

	// Compile ignore patterns into regex
	if err := compileIgnorePatterns(); err != nil {
		return fmt.Errorf("error compiling ignore patterns: %w", err)
	}

	// Print configuration in verbose mode
	if config.Format == "text" {
		printConfig()
	}

	// Run actual link checking
	return runRealLinkChecker()
}

func isURL(input string) bool {
	u, err := url.Parse(input)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func compileIgnorePatterns() error {
	config.IgnoreRegex = make([]*regexp.Regexp, 0, len(config.IgnoreList))

	for _, pattern := range config.IgnoreList {
		// Convert glob-like patterns to regex
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "?", ".")

		// If it doesn't look like a regex, treat it as a domain
		if !strings.ContainsAny(pattern, ".*+?^${}[]|()\\") {
			regexPattern = fmt.Sprintf(".*%s.*", regexp.QuoteMeta(pattern))
		}

		regex, err := regexp.Compile(regexPattern)
		if err != nil {
			return fmt.Errorf("invalid ignore pattern '%s': %w", pattern, err)
		}

		config.IgnoreRegex = append(config.IgnoreRegex, regex)
	}

	return nil
}

func printConfig() {
	fmt.Printf("Link Checker Configuration:\n")
	if len(config.InputPaths) > 0 {
		fmt.Printf("  File Paths: %v\n", config.InputPaths)
	}
	if len(config.InputURLs) > 0 {
		fmt.Printf("  URLs to Check: %v\n", config.InputURLs)
	}
	fmt.Printf("  Recursive: %v\n", config.Recursive)
	fmt.Printf("  Timeout: %v\n", config.Timeout)
	fmt.Printf("  Only Dead Links: %v\n", config.OnlyDead)
	fmt.Printf("  Output Format: %s\n", config.Format)
	fmt.Printf("  Workers: %d\n", config.Workers)
	if len(config.IgnoreList) > 0 {
		fmt.Printf("  Ignore Patterns: %v\n", config.IgnoreList)
	}
	fmt.Println()
}

func runRealLinkChecker() error {
	start := time.Now()
	results := []Result{}

	// Process file paths
	for _, inputPath := range config.InputPaths {
		fileResults, err := processPath(inputPath)
		if err != nil {
			return fmt.Errorf("error processing path '%s': %w", inputPath, err)
		}
		results = append(results, fileResults...)
	}

	// Process URLs
	for _, inputURL := range config.InputURLs {
		urlResults, err := processURL(inputURL)
		if err != nil {
			return fmt.Errorf("error processing URL '%s': %w", inputURL, err)
		}
		results = append(results, urlResults...)
	}

	// Filter results if only-dead is enabled
	if config.OnlyDead {
		filteredResults := make([]Result, 0)
		for _, result := range results {
			if result.Status == "invalid" {
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	// Create output
	output := Output{
		Results: results,
	}

	// Calculate summary
	valid := 0
	invalid := 0
	for _, result := range results {
		if result.Status == "valid" {
			valid++
		} else {
			invalid++
		}
	}

	output.Summary.Total = len(results)
	output.Summary.Valid = valid
	output.Summary.Invalid = invalid
	output.Summary.Duration = time.Since(start).String()

	// Output results
	if config.Format == "json" {
		return outputJSON(output)
	}

	return outputText(output)
}

func processPath(inputPath string) ([]Result, error) {
	var results []Result

	// Check if path exists
	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", inputPath)
	}

	if info.IsDir() {
		// Process directory
		err := filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && isMarkdownFile(path) {
				fileResults, err := processMarkdownFile(path)
				if err != nil {
					return err
				}
				results = append(results, fileResults...)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Process single file
		if isMarkdownFile(inputPath) {
			fileResults, err := processMarkdownFile(inputPath)
			if err != nil {
				return nil, err
			}
			results = append(results, fileResults...)
		}
	}

	return results, nil
}

func processMarkdownFile(filePath string) ([]Result, error) {
	links, err := parser.ExtractLinksFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error extracting links from %s: %w", filePath, err)
	}

	// Filter ignored links
	var filteredLinks []string
	for _, link := range links {
		if !IsURLIgnored(link) {
			filteredLinks = append(filteredLinks, link)
		}
	}

	// Validate links
	linkStatuses := validator.ValidateLinksAsync(filteredLinks, filepath.Dir(filePath), config.Timeout, config.Workers)

	var results []Result
	for _, status := range linkStatuses {
		result := Result{
			URL:    status.Link,
			Source: filePath,
		}

		if status.Valid {
			result.Status = "valid"
			result.StatusCode = status.StatusCode
		} else {
			result.Status = "invalid"
			result.Error = status.Reason
			result.StatusCode = status.StatusCode
		}

		results = append(results, result)
	}

	return results, nil
}

func processURL(inputURL string) ([]Result, error) {
	// Fetch the web page
	client := &http.Client{
		Timeout: config.Timeout,
	}

	resp, err := client.Get(inputURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching URL %s: %w", inputURL, err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response from %s: %w", inputURL, err)
	}

	// Extract links from HTML content
	links := parser.ExtractLinksFromHTML(content)

	if config.Debug {
		fmt.Printf("Debug: Found %d raw links in HTML\n", len(links))
	}

	// Convert relative URLs to absolute URLs
	baseURL, err := url.Parse(inputURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL %s: %w", inputURL, err)
	}

	var absoluteLinks []string
	for _, link := range links {
		// Skip empty links, anchors, and javascript/mailto links
		if link == "" ||
			strings.HasPrefix(link, "#") ||
			strings.HasPrefix(link, "javascript:") ||
			strings.HasPrefix(link, "mailto:") ||
			strings.HasPrefix(link, "tel:") {
			if config.Debug && link != "" {
				fmt.Printf("Debug: Skipping link: %s (anchor/javascript/mailto/tel)\n", link)
			}
			continue
		}

		// Parse the link URL
		linkURL, err := url.Parse(link)
		if err != nil {
			if config.Debug {
				fmt.Printf("Debug: Failed to parse link: %s (error: %v)\n", link, err)
			}
			continue // Skip invalid URLs
		}

		// Resolve relative URLs to absolute URLs
		absoluteURL := baseURL.ResolveReference(linkURL)

		if config.Debug {
			fmt.Printf("Debug: %s -> %s\n", link, absoluteURL.String())
		}

		// Only include HTTP/HTTPS URLs for validation
		if absoluteURL.Scheme == "http" || absoluteURL.Scheme == "https" {
			absoluteLinks = append(absoluteLinks, absoluteURL.String())
		} else if config.Debug {
			fmt.Printf("Debug: Skipping non-HTTP(S) URL: %s (scheme: %s)\n", absoluteURL.String(), absoluteURL.Scheme)
		}
	}

	if config.Debug {
		fmt.Printf("Debug: %d links will be validated\n", len(absoluteLinks))
	}

	// Filter ignored links
	var filteredLinks []string
	for _, link := range absoluteLinks {
		if !IsURLIgnored(link) {
			filteredLinks = append(filteredLinks, link)
		}
	}

	// Validate links
	linkStatuses := validator.ValidateLinksAsync(filteredLinks, "", config.Timeout, config.Workers)

	var results []Result
	for _, status := range linkStatuses {
		result := Result{
			URL:    status.Link,
			Source: inputURL,
		}

		if status.Valid {
			result.Status = "valid"
			result.StatusCode = status.StatusCode
		} else {
			result.Status = "invalid"
			result.Error = status.Reason
			result.StatusCode = status.StatusCode
		}

		results = append(results, result)
	}

	return results, nil
}

func isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

func outputJSON(output Output) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputText(output Output) error {
	fmt.Printf("Link Check Results\n")
	fmt.Printf("==================\n\n")

	// Group results by source
	sourceGroups := make(map[string][]Result)
	for _, result := range output.Results {
		sourceGroups[result.Source] = append(sourceGroups[result.Source], result)
	}

	for source, results := range sourceGroups {
		if isURL(source) {
			fmt.Printf("🌐 Checking web page: %s\n", source)
		} else {
			fmt.Printf("📄 Checking file: %s\n", source)
		}
		fmt.Println(strings.Repeat("-", len(source)+20))

		for _, result := range results {
			status := "✓"
			if result.Status == "invalid" {
				status = "✗"
			}

			fmt.Printf("%s %s\n", status, result.URL)
			if result.Line > 0 {
				fmt.Printf("  Line: %d\n", result.Line)
			}

			if result.StatusCode > 0 {
				fmt.Printf("  Status: %d\n", result.StatusCode)
			}
			if result.Error != "" {
				fmt.Printf("  Error: %s\n", result.Error)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	fmt.Printf("Summary:\n")
	fmt.Printf("  Total Links: %d\n", output.Summary.Total)
	fmt.Printf("  Valid: %d\n", output.Summary.Valid)
	fmt.Printf("  Invalid: %d\n", output.Summary.Invalid)
	fmt.Printf("  Duration: %s\n", output.Summary.Duration)

	return nil
}

func getHelpTemplate() string {
	return `{{.Short}}

{{.Long}}

Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
}

// Execute runs the CLI
func Execute() error {
	return rootCmd.Execute()
}

// GetConfig returns the current CLI configuration
// This can be used by other modules to access CLI settings
func GetConfig() Config {
	return config
}

// IsURLIgnored checks if a URL should be ignored based on the ignore patterns
func IsURLIgnored(url string) bool {
	for _, regex := range config.IgnoreRegex {
		if regex.MatchString(url) {
			return true
		}
	}
	return false
}
