# Link Checker

A fast and reliable link checker for markdown files and web pages written in Go.

## Features

- ✅ **Recursive scanning** - Scan directories recursively for markdown files
- ✅ **Direct URL checking** - Check web pages directly for dead links
- ✅ **Flexible ignore patterns** - Ignore specific domains or regex patterns
- ✅ **Configurable timeout** - Set custom HTTP request timeouts
- ✅ **Dead link filtering** - Show only broken links
- ✅ **Multiple output formats** - Text and JSON output formats
- ✅ **Detailed reporting** - Shows source file, line numbers, and error details
- ✅ **Mixed input support** - Combine files, directories, and URLs in one command

## Installation

```bash
go build -o linkchecker main.go
```

## Usage

### Basic Usage

```bash
# Check links in current directory
./linkchecker

# Check specific files
./linkchecker README.md docs/guide.md

# Check specific directory
./linkchecker ./docs

# Check web pages for dead links
./linkchecker https://example.com
./linkchecker https://github.com/user/repo

# Mixed usage - files and URLs
./linkchecker README.md https://example.com ./docs
```

### Available Flags

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--recursive` | `-r` | Recursively scan directories for markdown files | `--recursive` |
| `--ignore` | | Comma-separated list of domains or regex patterns to ignore | `--ignore="example.com,*.test.local"` |
| `--timeout` | | HTTP request timeout | `--timeout=10s` |
| `--only-dead` | | Only show dead/broken links in output | `--only-dead` |
| `--format` | | Output format: 'text' or 'json' | `--format=json` |

### Examples

#### Recursive scan with custom timeout
```bash
./linkchecker --recursive --timeout=10s ./docs
```

#### Check web page for dead links
```bash
./linkchecker https://example.com
./linkchecker --only-dead --format=json https://github.com/user/repo
```

#### Show only broken links in JSON format
```bash
./linkchecker --only-dead --format=json ./
```

#### Ignore specific domains
```bash
./linkchecker --ignore="example.com,localhost,*.test.local" ./docs
```

#### Check URL with custom settings
```bash
./linkchecker --ignore="ads.example.com,*.tracking.com" --timeout=10s https://mywebsite.com
```

#### Complex example with mixed inputs
```bash
./linkchecker \
  --recursive \
  --ignore="example.com,*.test.local" \
  --timeout=15s \
  --only-dead \
  --format=json \
  ./docs https://mysite.com README.md
```

## Output Formats

### Text Output (Default)

```
Link Checker Configuration:
  File Paths: [./docs]
  URLs to Check: [https://example.com]
  Recursive: true
  Timeout: 30s
  Only Dead Links: false
  Output Format: text

Link Check Results
==================

📄 Checking file: README.md
------------------------------------
✓ https://example.com
  Line: 10
  Status: 200

✗ https://broken-link.example
  Line: 25
  Status: 404
  Error: 404 Not Found

🌐 Checking web page: https://example.com
------------------------------------------
✓ https://external-link.com
  Status: 200

✗ https://dead-external-link.com
  Status: 404
  Error: 404 Not Found

Summary:
  Total Links: 5
  Valid: 2
  Invalid: 3
  Duration: 1.234s
```

### JSON Output

```json
{
  "summary": {
    "total": 2,
    "valid": 1,
    "invalid": 1,
    "duration": "1.234s"
  },
  "results": [
    {
      "url": "https://example.com",
      "status": "valid",
      "status_code": 200,
      "source": "README.md",
      "line": 10
    },
    {
      "url": "https://broken-link.example",
      "status": "invalid",
      "status_code": 404,
      "error": "404 Not Found",
      "source": "docs/guide.md",
      "line": 25
    }
  ]
}
```

## Ignore Patterns

The `--ignore` flag supports both simple domain matching and regex patterns:

- **Simple domains**: `example.com` - matches any URL containing "example.com"
- **Wildcards**: `*.test.local` - matches any subdomain of test.local
- **Regex patterns**: Use full regex syntax for complex patterns

Examples:
```bash
# Ignore specific domains
--ignore="example.com,localhost"

# Ignore with wildcards
--ignore="*.test.local,*.example.org"

# Mix of patterns
--ignore="example.com,*.test.local,localhost:*"
```

## Development

### Project Structure

```
link-checker/
├── linkchecker/
│   ├── cli/          # CLI implementation
│   ├── parser/       # Markdown parsing logic
│   └── validator/    # Link validation logic
├── go.mod
├── go.sum
├── main.go
└── README.md
```

### Building

```bash
go build -o linkchecker main.go
```

### Testing

```bash
go test ./...
```

## License

MIT License 