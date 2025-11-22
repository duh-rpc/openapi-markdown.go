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

### Response Example Generation

The converter uses a three-tier priority system for generating JSON examples:

1. **Explicit examples**: Uses `example` field from response media type
2. **Named examples**: Uses first entry from `examples` collection
3. **Schema-based**: Generates from $ref using `openapi-schema.go` library

**Critical constraint**: Response schemas MUST use `$ref` to reference schemas in `components/schemas`. Inline schemas in responses are rejected with an error.

### Internal Flow

1. Parse OpenAPI document with `libopenapi`
2. Generate component examples using `openapi-schema.go`
3. Extract endpoints from paths
4. Group endpoints by tags (untagged operations go to "Default APIs")
5. Generate markdown with TOC, tag sections, and endpoints
6. Collect debug info if requested

### Tag Grouping Behavior

- Operations with multiple tags appear in each tag section
- Untagged operations grouped under "Default APIs"
- When only one tag exists, tag sections are omitted (endpoints rendered at top level)
- "Default APIs" section is always sorted last

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
