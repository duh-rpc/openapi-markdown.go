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

## Complete Example

A comprehensive example demonstrating all converter features is available in the `examples/` directory:

- `examples/openapi.yaml` - Complete OpenAPI specification with multiple tags, parameters, and responses
- `examples/example.md` - Expected markdown output demonstrating all features

The example demonstrates:
- Multiple tags (pets, users, orders, admin)
- Untagged operations (Default APIs section)
- Operations with multiple tags appearing in each section
- All parameter types: path, query, header
- All parameter data types: string, integer, boolean
- Multiple response codes (200, 201, 400, 403, 404)
- Rich descriptions and summaries

### Regenerating Example Output

When markdown format changes are intentional, regenerate the expected output:

```bash
cat > /tmp/regenerate_example.go <<'EOF'
package main

import (
    "os"
    conv "github.com/duh-rpc/openapi-markdown.go"
)

func main() {
    openapi, _ := os.ReadFile("examples/openapi.yaml")
    result, _ := conv.Convert(openapi, conv.ConvertOptions{
        Title: "Pet Store API",
        Description: "A comprehensive API for managing a pet store with users, pets, and orders",
    })
    os.WriteFile("examples/example.md", result.Markdown, 0644)
}
EOF
go run /tmp/regenerate_example.go
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
