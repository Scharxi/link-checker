package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

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
	IgnoreRegex []*regexp.Regexp
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
		Use:   "linkchecker [paths...]",
		Short: "A fast and reliable link checker for markdown files",
		Long: `Link Checker is a command-line tool that validates links in markdown files.
It can recursively scan directories, ignore specific domains or patterns,
and output results in various formats.`,
		Example: `  linkchecker README.md
  linkchecker --recursive ./docs
  linkchecker --ignore="example.com,test.local" --timeout=10s ./
  linkchecker --only-dead --format=json ./docs`,
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

	// Add help examples
	rootCmd.SetHelpTemplate(getHelpTemplate())
}

func runLinkChecker(cmd *cobra.Command, args []string) error {
	// Set input paths
	if len(args) == 0 {
		config.InputPaths = []string{"."}
	} else {
		config.InputPaths = args
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

	// TODO: Implement actual link checking logic
	// For now, return a placeholder implementation
	return runPlaceholderChecker()
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
	fmt.Printf("  Paths: %v\n", config.InputPaths)
	fmt.Printf("  Recursive: %v\n", config.Recursive)
	fmt.Printf("  Timeout: %v\n", config.Timeout)
	fmt.Printf("  Only Dead Links: %v\n", config.OnlyDead)
	fmt.Printf("  Output Format: %s\n", config.Format)
	if len(config.IgnoreList) > 0 {
		fmt.Printf("  Ignore Patterns: %v\n", config.IgnoreList)
	}
	fmt.Println()
}

func runPlaceholderChecker() error {
	start := time.Now()

	// Placeholder results for demonstration
	results := []Result{
		{
			URL:        "https://example.com",
			Status:     "valid",
			StatusCode: 200,
			Source:     "README.md",
			Line:       10,
		},
		{
			URL:        "https://broken-link.example",
			Status:     "invalid",
			StatusCode: 404,
			Error:      "404 Not Found",
			Source:     "docs/guide.md",
			Line:       25,
		},
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
	output.Summary.Total = 2
	output.Summary.Valid = 1
	output.Summary.Invalid = 1
	output.Summary.Duration = time.Since(start).String()

	// Output results
	if config.Format == "json" {
		return outputJSON(output)
	}

	return outputText(output)
}

func outputJSON(output Output) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func outputText(output Output) error {
	fmt.Printf("Link Check Results\n")
	fmt.Printf("==================\n\n")

	for _, result := range output.Results {
		status := "✓"
		if result.Status == "invalid" {
			status = "✗"
		}

		fmt.Printf("%s %s\n", status, result.URL)
		fmt.Printf("  Source: %s", result.Source)
		if result.Line > 0 {
			fmt.Printf(":%d", result.Line)
		}
		fmt.Println()

		if result.StatusCode > 0 {
			fmt.Printf("  Status: %d\n", result.StatusCode)
		}
		if result.Error != "" {
			fmt.Printf("  Error: %s\n", result.Error)
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
