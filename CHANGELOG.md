# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Asynchronous link validation with configurable worker pool
- Debug mode for troubleshooting link processing
- Support for relative URL resolution on web pages
- Cross-platform CI/CD pipeline with GitHub Actions
- Comprehensive test suite and code quality checks
- Version command to display build information

### Changed
- Improved HTTP client with redirect following and fallback to GET requests
- Better error handling and status code reporting
- Enhanced CLI configuration display

### Fixed
- Relative URL resolution for web page link checking
- Timeout handling for HTTP requests
- Status code extraction and reporting

## [v1.0.0] - Initial Release

### Added
- Link validation for Markdown files
- Web page link checking
- Recursive directory scanning
- Configurable timeout and ignore patterns
- JSON and text output formats
- Support for both file paths and URLs as input
- Cross-platform builds (Linux, macOS, Windows)

### Features
- **Markdown Support**: Extract and validate links from `.md` and `.markdown` files
- **Web Page Checking**: Fetch web pages and validate all contained links
- **Flexible Input**: Accept file paths, directories, or URLs as arguments
- **Ignore Patterns**: Skip links matching specified regex patterns
- **Output Formats**: Choose between human-readable text or machine-readable JSON
- **Performance**: Configurable timeout and concurrent workers for fast validation
- **Cross-Platform**: Native binaries for Linux, macOS, and Windows

### Usage Examples
```bash
# Check markdown files
linkchecker README.md
linkchecker --recursive ./docs

# Check web pages
linkchecker https://example.com

# Advanced options
linkchecker --only-dead --format=json --workers=20 https://example.com
``` 