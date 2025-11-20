# OpenAPI to Markdown Examples

This directory contains example OpenAPI specifications and their generated markdown output.

## Files

- `openapi.yaml` - Complete OpenAPI 3.0 specification for a pet store API
- `example.md` - Generated markdown documentation from openapi.yaml

## Generating the Example

To regenerate `example.md`, use the regeneration script from the main README.md or use the library directly:

```go
package main

import (
    "os"
    conv "github.com/duh-rpc/openapi-markdown.go"
)

func main() {
    openapi, _ := os.ReadFile("openapi.yaml")
    result, _ := conv.Convert(openapi, conv.ConvertOptions{
        Title:       "Pet Store API",
        Description: "A comprehensive API for managing a pet store with users, pets, and orders",
    })
    os.WriteFile("example.md", result.Markdown, 0644)
}
```

## OpenAPI Spec Features

The example spec demonstrates:

- **Component Schemas**: All data types defined in `components/schemas`
- **$ref Usage**: All responses use `$ref` to reference schemas (required)
- **Explicit Examples**: Some endpoints have explicit `example` fields
- **Tag Organization**: Endpoints grouped by tags (admin, pets, users, orders)
- **Parameter Types**: Path, query, and header parameters
- **Multiple Response Codes**: Success (200, 201) and error responses (400, 403, 404)
- **Array Schemas**: Paginated list responses
- **Enum Values**: Status fields with enumerated values
- **Nested Objects**: Complex response structures with references

## Response Schema Requirements

**Important**: Response schemas MUST use `$ref` to reference schemas in `components/schemas`.
Inline schemas in responses are not supported.

✅ **Correct**:
```yaml
responses:
  '200':
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Pet'
```

❌ **Incorrect** (will cause error):
```yaml
responses:
  '200':
    content:
      application/json:
        schema:
          type: object
          properties:
            name:
              type: string
```

## Endpoints in Example

### Pets
- GET /v3/pets - List all pets (with pagination)
- POST /v3/pets - Create a new pet
- POST /v3/pets.delete - Delete a pet (admin only)
- GET /v3/pets/{petId} - Get pet by ID

### Users
- GET /v3/users - List all users
- POST /v3/users - Create a new user
- GET /v3/users/{userId} - Get user by ID
- GET /v3/users/{userId}/orders - Get user's orders

### Orders
- GET /v3/orders - List all orders (with pagination)
- POST /v3/orders - Create a new order
- GET /v3/orders/{orderId} - Get order by ID

### System
- GET /v3/health - Health check endpoint
- GET /v3/metrics - Get API metrics (admin only)
