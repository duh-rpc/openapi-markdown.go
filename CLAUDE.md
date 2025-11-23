# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library that converts OpenAPI 3.x specifications to markdown documentation. The library has a single public
API (`Convert()`) and follows functional testing principles where all tests interact through this public interface.

## Architecture

### Core Components

- **`Convert()`**: Single public API that takes OpenAPI bytes and options, returns markdown with metadata
- **`ConvertOptions`**: Configuration struct with Title, Description, and Debug flag
- **`ConvertResult`**: Output struct containing Markdown bytes, counts, and optional Debug info
- **`DebugInfo`**: Observability structure exposing internal conversion metrics for testing

### Example Generation

The converter uses a three-tier priority system for generating JSON examples for both request bodies and responses:

1. **Explicit examples**: Uses `example` field from media type
2. **Named examples**: Uses first entry from `examples` collection
3. **Schema-based**: Generates from $ref using `openapi-schema.go` library

**Critical constraint**: Request and response schemas MUST use `$ref` to reference schemas in `components/schemas`. Inline schemas are rejected with an error.

### Internal Flow

1. Parse OpenAPI document with `libopenapi`
2. Generate component examples using `openapi-schema.go`
3. Extract endpoints from paths
4. Identify shared schemas (used across multiple endpoints)
5. Group endpoints by tags (untagged operations go to "Default APIs")
6. Generate markdown with TOC, shared schema definitions, tag sections, and endpoints
7. Collect debug info if requested

### Tag Grouping Behavior

- Operations with multiple tags appear in each tag section
- Untagged operations grouped under "Default APIs"
- When only one tag exists, tag sections are omitted (endpoints rendered at top level)
- "Default APIs" section is always sorted last

### Field Definitions Format

The converter renders comprehensive field definitions for request bodies, responses, and parameters using a hierarchical format:

**Request Bodies** (POST/PUT/PATCH/DELETE):
- JSON example generated from schema
- Field definitions section documenting each field
- Top-level objects use bold headers: `**fieldName**`
- Nested fields use bulleted list with backticks: `` - `fieldName` (type, required): Description ``
- Nested objects get separate definition sections

**Responses**:
- All responses show JSON examples
- Field definitions only for 2xx success responses
- 4xx/5xx error responses show JSON only (no field docs)
- Responses use H4 headings: `#### 200 Response`

**Parameters**:
- Path and Query parameters use field definitions format: `**paramName** (type, required)`
- Headers remain in table format for compactness
- Enum values shown inline with field: `` Enums: `VAL1`, `VAL2` ``

**Shared Schemas**:
- Schemas used across multiple endpoints documented once in "Shared Schema Definitions" section
- Endpoints reference shared schemas instead of duplicating documentation
- Same-endpoint reuse (e.g., request + response) does NOT create shared schema

**Nesting and Recursion**:
- Maximum nesting depth: 10 levels
- Recursive schemas capped at depth 1 (marked as "(recursive)")
- Arrays of primitives shown inline with enums
- Arrays of objects expect named component schemas

Reference: `examples/updated-example.md` shows canonical format with nested objects, arrays, and enums.

## Regenerating Example Output

When markdown format changes are intentional:

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

Also regenerate the golden file:

```bash
cp examples/example.md testdata/golden/petstore-example.md
```

## Testing Philosophy

This library follows **functional testing principles**:

- All tests use the public `Convert()` API exclusively
- Internal functions (`extractEndpoints`, `groupByTags`, etc.) are NEVER tested directly
- Internal behavior is verified through the `Debug` field observability API
- Tests validate both generated markdown content and metadata (counts, debug info)

**When adding features**: Expose observability through `DebugInfo` rather than testing internal functions directly.
