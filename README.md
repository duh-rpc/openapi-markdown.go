# OpenAPI to Markdown Converter

A Go library that converts OpenAPI 3.x specifications to comprehensive markdown API documentation.

## Installation

```bash
go get github.com/duh-rpc/openapi-markdown.go
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "os"

    conv "github.com/duh-rpc/openapi-markdown.go"
)

func main() {
    openapi, err := os.ReadFile("api.yaml")
    if err != nil {
        panic(err)
    }

    result, err := conv.Convert(openapi, conv.ConvertOptions{
        Title:       "My API",
        Description: "API documentation",
    })
    if err != nil {
        panic(err)
    }

    err = os.WriteFile("API.md", result.Markdown, 0644)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Generated documentation for %d endpoints in %d sections\n",
        result.EndpointCount, result.TagCount)
}
```

### Debug Mode

```go
result, err := conv.Convert(openapi, conv.ConvertOptions{
    Title: "My API",
    Debug: true,  // Enable debug information
})

fmt.Printf("Parsed %d paths\n", result.Debug.ParsedPaths)
fmt.Printf("Extracted %d operations\n", result.Debug.ExtractedOps)
```

## Testing Philosophy

This library follows **functional testing principles**:
- All tests interact through the public `Convert()` API
- Internal functions are never tested directly
- Internal behavior is verified through observability (Debug field)
- Tests validate generated markdown content and metadata

## Development

### Running Tests
```bash
make test
```

### Running Full CI Suite
```bash
make ci
```

### Test Coverage
```bash
make coverage
```
