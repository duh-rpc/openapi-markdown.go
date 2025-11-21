# OpenAPI to Markdown Converter

A Go library that converts OpenAPI 3.x specifications to comprehensive markdown API documentation.

[![Go Version](https://img.shields.io/github/go-mod/go-version/duh-rpc/openapi-markdown.go)](https://golang.org/dl/)
[![CI Status](https://github.com/duh-rpc/openapi-markdown.go/workflows/CI/badge.svg)](https://github.com/duh-rpc/openapi-markdown.go/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/duh-rpc/openapi-markdown.go)](https://goreportcard.com/report/github.com/duh-rpc/openapi-markdown.go)

## Features

- Converts OpenAPI 3.x specifications to clean, readable Markdown
- Generates Table of Contents with anchor links
- Organizes endpoints by tags
- Renders parameter tables (path, query, header)
- **Generates JSON response examples from schemas**
- Supports explicit examples and schema-based generation
- Validates that response schemas use $ref (no inline schemas)

## Installation

```bash
go get github.com/duh-rpc/openapi-markdown.go
```

## Response Examples

The converter automatically generates JSON examples for API responses using three priority levels:

1. **Explicit examples**: Uses `example` field from response media type
2. **Named examples**: Uses first entry from `examples` collection
3. **Schema-based**: Generates from $ref schema using openapi-schema.go library

**Important**: Response schemas must use `$ref` to reference schemas in `components/schemas`. Inline schemas in responses are not supported and will cause an error.

### Example OpenAPI Spec

```yaml
paths:
  /pets:
    get:
      responses:
        '200':
          description: List of pets
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PetList'  # Must use $ref
              example:  # Optional explicit example
                pets:
                  - id: "123"
                    name: "Fluffy"
components:
  schemas:
    PetList:
      type: object
      properties:
        pets:
          type: array
          items:
            $ref: '#/components/schemas/Pet'
    Pet:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
```

### Generated Markdown

```markdown
##### 200 Response

List of pets

\`\`\`json
{
   "pets": [
      {
         "id": "123",
         "name": "Fluffy"
      }
   ]
}
\`\`\`
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

## Requirements

- Go 1.25.4 or later
- Dependencies:
  - github.com/pb33f/libopenapi v0.28.2 (OpenAPI parsing)
  - github.com/duh-rpc/openapi-schema.go v0.7.0 (Example generation)
  - github.com/stretchr/testify v1.11.1 (Testing)

## Complete Example

A comprehensive example demonstrating all converter features is available in the `examples/` directory:

- `examples/openapi.yaml` - Complete OpenAPI specification with multiple tags, parameters, and responses
- `examples/example.md` - Expected markdown output demonstrating all features
- `examples/README.md` - Detailed documentation about the example spec

The example demonstrates:
- Multiple tags (pets, users, orders, admin)
- Untagged operations (Default APIs section)
- Operations with multiple tags appearing in each section
- All parameter types: path, query, header
- All parameter data types: string, integer, boolean
- Multiple response codes (200, 201, 400, 403, 404)
- **JSON response examples with explicit and generated examples**
- **Response schemas using $ref (required pattern)**
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
