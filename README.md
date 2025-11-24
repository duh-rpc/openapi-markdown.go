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
- **Request body documentation** with JSON examples and field definitions
- **Nested schema documentation** with hierarchical field definitions
- **Shared schema definitions** documented once, referenced across endpoints
- **Parameter documentation** with enum support (path/query use field definitions, headers use tables)
- **Rich response field documentation** for 2xx success responses
- Generates JSON examples from schemas (explicit, named, or schema-based)
- Validates that schemas use $ref (no inline schemas)
- Supports recursion detection and depth limiting

## Installation

```bash
go get github.com/duh-rpc/openapi-markdown.go
```

## Documentation Format

The converter generates comprehensive documentation with hierarchical field definitions for schemas.

### Request Bodies

Request bodies (for POST/PUT/PATCH/DELETE operations) include:
- JSON example generated from schema
- Field definitions with type information, required status, and descriptions
- Nested objects documented in separate sections
- Enum values shown inline

### Response Documentation

All responses include JSON examples. Success responses (2xx) also include detailed field definitions:
- Field type and required status
- Field descriptions
- Nested object documentation
- Array handling (primitives inline, objects as separate definitions)

Error responses (4xx/5xx) show JSON examples only.

### Example Generation

The converter automatically generates JSON examples using three priority levels:

1. **Explicit examples**: Uses `example` field from media type
2. **Named examples**: Uses first entry from `examples` collection
3. **Schema-based**: Generates from $ref schema using openapi-schema.go library

**Important**: Request and response schemas must use `$ref` to reference schemas in `components/schemas`. Inline schemas are not supported and will cause an error.

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

The new field definitions format provides comprehensive schema documentation:

```markdown
### Request

\`\`\`json
{
   "name": "Fluffy",
   "species": "cat"
}
\`\`\`

#### Field Definitions

**name** (string, required)
- The name of the pet

**species** (string, required)
- The species of the pet. Enums: `cat`, `dog`, `bird`

### Responses

#### 200 Response

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

#### Field Definitions

**pets** (array of objects)
- List of all pets in the store
- `id` (string): The unique identifier
- `name` (string): The name of the pet
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
- Request body documentation with field definitions
- Nested schema definitions with separate sections
- All parameter types with new field definitions format (path, query, headers)
- Parameter enums shown inline
- Multiple response codes (200, 201, 400, 403, 404)
- Response field definitions for 2xx success responses
- JSON examples with explicit and generated examples
- Schemas using $ref (required pattern)
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
