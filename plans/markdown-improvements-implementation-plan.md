# Markdown Improvements Implementation Plan

## Overview

This plan implements two improvements to the OpenAPI-to-markdown conversion:

1. **Respect schema/property examples**: Use explicit `example` or `examples` fields from schemas instead of always generating synthetic examples
2. **Show schema names for object types**: Display `*(Address)*` instead of `*(object)*` when a field references another schema

These changes span two repositories:
- `openapi-schema.go` - Example generation improvements
- `openapi-markdown.go` - Type display improvements

## Example Priority (Definitive Order)

When generating examples, the following priority applies (highest to lowest):

1. **Schema-level `example`** - Complete example for entire schema object
2. **Schema-level `examples[0]`** - OpenAPI 3.1 array format, use first entry
3. **Property-level `example`** - Individual field example (existing behavior)
4. **Property-level `examples[0]`** - OpenAPI 3.1 array format for individual field
5. **Generated value** - Current algorithmic generation (fallback)

**Note**: `example` (singular) takes precedence over `examples` (plural) for backward compatibility with OpenAPI 3.0 specs that only have `example`.

### Concrete Example

```yaml
components:
  schemas:
    # Schema-level example - entire object defined (highest priority)
    Transfer:
      type: object
      example:                           # <-- This is used for the whole object
        id: "xfer_complete_example"
        amount: 100
      properties:
        id:
          type: string
        amount:
          type: integer

    # Property-level examples - individual field examples
    Payment:
      type: object
      properties:
        id:
          type: string
          example: "pay_abc123"          # <-- Used for this field
        transactionId:
          type: string
          examples: ["txn_1a2b3c"]       # <-- Used when example is absent
        amount:
          type: integer                  # <-- Generated (no example)
```

## Current State Analysis

### Example Generation (`openapi-schema.go`)

The `internal/example/generator.go` currently:
- ✅ Respects `schema.Example` for **scalar properties only** (line 137-139 in `generateScalarValue`)
- ❌ Does NOT check `schema.Example` for objects/arrays at schema level
- ❌ Does NOT handle `schema.Examples` (plural, OpenAPI 3.1 array format)

### Type Display (`openapi-markdown.go`)

The `convert.go` currently:
- Stores schema name in `schemaField.nestedSchemaRef` (line 123)
- ❌ Always displays `*(object)*` instead of using the stored schema name (lines 1122-1124, 1355, 1531)

## Desired End State

After implementation:

1. **Example priority** (highest to lowest):
   - Schema-level `example` field (complete example for entire schema)
   - Property-level `example` fields (build object from individual property examples)
   - Property-level `examples[0]` (OpenAPI 3.1 array format, use first entry)
   - Generated examples (current behavior, fallback)

2. **Type display**: When a field is a `$ref` to another schema, show the schema name:
   ```markdown
   - `address` *(Address, required)* User's shipping address
   ```
   Instead of:
   ```markdown
   - `address` *(object, required)* User's shipping address
   ```

### Verification

- All existing tests pass
- New tests verify example priority behavior
- Golden file updated to reflect new type display format
- Manual verification with sample OpenAPI specs containing explicit examples

## What We're NOT Doing

- Not changing the example generation for scalars (already works)
- Not adding support for `examples` at the MediaType level (already handled in `openapi-markdown.go`)
- Not linking schema names to anchors (user confirmed just the name, not a link)

## Implementation Approach

Split into three phases:
1. **Phase 1**: Update `openapi-schema.go` to respect schema-level examples
2. **Phase 2**: Update `openapi-schema.go` to respect property-level `examples` (plural)
3. **Phase 3**: Update `openapi-markdown.go` to display schema names instead of "object"

---

## Phase 1: Schema-Level Example Support (`openapi-schema.go`)

### Overview

Add support for schema-level `example` field on objects and arrays. When a schema has an explicit `example`, use it directly instead of generating field-by-field.

### Changes Required

#### 1. Update `generateExample` function

**File**: `internal/example/generator.go`

**Changes**: Add schema-level example check before generating object/array examples

```go
// generateExample generates a JSON example for a single schema
func generateExample(name string, proxy *base.SchemaProxy, ctx *ExampleContext) (interface{}, error)
```

**Function Responsibilities:**
- Check for `schema.Example` at the beginning of the function (after nil/circular checks)
- If `schema.Example` is present, decode and return it directly using `decodeYAMLNode`
- If `schema.Examples` is present and non-empty, use `Examples[0]`
- Otherwise, fall through to existing generation logic

#### 2. Add `decodeYAMLNode` helper function

**File**: `internal/example/generator.go`

**Changes**: Add new helper to decode complex YAML nodes (objects/arrays), not just scalars

```go
// decodeYAMLNode recursively decodes a yaml.Node into a Go value
func decodeYAMLNode(node *yaml.Node) (interface{}, error)
```

**Function Responsibilities:**
- Handle `yaml.MappingNode` (Kind=4, objects):
  - Iterate pairs: `node.Content[i]` = key, `node.Content[i+1]` = value
  - Recursively decode each value
  - Return `map[string]interface{}`
- Handle `yaml.SequenceNode` (Kind=2, arrays):
  - Iterate `node.Content`
  - Recursively decode each element
  - Return `[]interface{}`
- Handle `yaml.ScalarNode` (Kind=8):
  - Delegate to existing `extractYAMLNodeValue`
- Handle `nil` node:
  - Return `nil, nil`
- Return error for `yaml.DocumentNode`, `yaml.AliasNode`, or unknown types

**Error Handling:**
- If a nested decode fails, propagate the error up
- No recursion limit needed (YAML structure is finite from spec)
- Malformed nodes return descriptive errors: `"unsupported yaml node kind: %d"`

### Testing Requirements

**File**: `convert_examples_test.go` (package `schema_test`)

All tests go through the public `ConvertToExamples()` API.

```go
func TestConvertToExamples_SchemaLevelExample(t *testing.T)
func TestConvertToExamples_SchemaLevelExamplesArray(t *testing.T)
```

**Test Objectives:**
- Verify schema with `example` field uses that example verbatim (via `ConvertToExamples` output)
- Verify schema with `examples` array uses first entry (via `ConvertToExamples` output)
- Verify `example` takes precedence over `examples` when both present
- Verify fallback to generation when neither is present
- Internal helper `decodeYAMLNode` is NOT tested directly; its behavior is verified through public API output

**Edge Cases to Test:**
- Schema with empty `examples: []` falls back to generation
- Schema with `example: null` is treated as explicit null value
- Nested objects in schema-level example are decoded correctly (verified via JSON output)
- Arrays in schema-level example are decoded correctly (verified via JSON output)

### Validation
- [ ] Run: `go test ./...` in `openapi-schema.go` directory
- [ ] Verify: All existing tests pass
- [ ] Verify: New tests for schema-level examples pass

---

## Phase 2: Property-Level `examples` Support (`openapi-schema.go`)

### Overview

Add support for the OpenAPI 3.1 `examples` array field on properties. When `example` (singular) is not present but `examples` (plural) is, use the first entry.

### Changes Required

#### 1. Update `generateScalarValue` function

**File**: `internal/example/generator.go`

**Changes**: Add fallback to `schema.Examples[0]` when `schema.Example` is nil

```go
func generateScalarValue(fieldName string, schema *base.Schema, typ, format string, ctx *ExampleContext) (interface{}, error)
```

**Function Responsibilities:**
- Keep existing `schema.Example` check (line 137-139)
- Add new check after Example block:
  ```go
  // schema.Examples is []*yaml.Node per libopenapi
  if len(schema.Examples) > 0 && schema.Examples[0] != nil {
      return extractYAMLNodeValue(schema.Examples[0]), nil
  }
  ```
- Place this check after `Example` but before `Default`
- Priority order: Example → Examples[0] → Default → FieldOverrides → Generated

#### 2. Update `generatePropertyValue` function

**File**: `internal/example/generator.go`

**Changes**: Check for property-level example before delegating to type-specific generation

```go
func generatePropertyValue(propertyName string, propProxy *base.SchemaProxy, ctx *ExampleContext) (interface{}, error)
```

**Function Responsibilities:**
- After getting schema from proxy (line 404), add example checks before type-specific logic:
  ```go
  // Check for explicit example on this property (for non-scalar types)
  if schema.Example != nil {
      return decodeYAMLNode(schema.Example)
  }
  if len(schema.Examples) > 0 && schema.Examples[0] != nil {
      return decodeYAMLNode(schema.Examples[0])
  }
  ```
- Place this check AFTER the `$ref` handling block (lines 409-428)
- Place BEFORE the array/object type checks (lines 430-443)
- This allows explicit examples on object/array properties to be used

### Testing Requirements

**File**: `convert_examples_test.go` (package `schema_test`)

All tests go through the public `ConvertToExamples()` API.

```go
func TestConvertToExamples_PropertyLevelExamples(t *testing.T)
func TestConvertToExamples_PropertyLevelExampleObject(t *testing.T)
```

**Test Objectives:**
- Verify property with `examples: ["value1", "value2"]` uses "value1" (via JSON output)
- Verify `example` takes precedence over `examples`
- Verify object property with explicit `example` uses that example (via JSON output)
- Verify array property with explicit `example` uses that example (via JSON output)

**Edge Cases to Test:**
- Property with `examples: []` (empty array) falls back to generation
- Property with complex object in `examples[0]` is decoded correctly (via JSON output)
- Scalar property with `examples` uses first entry

### Validation
- [ ] Run: `go test ./...` in `openapi-schema.go` directory
- [ ] Verify: All existing tests pass
- [ ] Verify: New tests for property-level examples pass

---

## Phase 3: Schema Name Display (`openapi-markdown.go`)

### Overview

Update the markdown renderer to display schema names instead of "object" when a field is a `$ref` to another schema.

### Changes Required

#### 1. Update `renderFieldDefinitionsContent` function

**File**: `convert.go`

**Changes**: Use `field.nestedSchemaRef` instead of hardcoded "object"

Current code (lines 1118-1124):
```go
} else if field.isArray && field.isObject {
    builder.WriteString("array of objects")
} else if field.isObject {
    builder.WriteString("object")
} else {
```

New logic:
```go
} else if field.isArray && field.isObject {
    if field.nestedSchemaRef != "" {
        builder.WriteString("array of ")
        builder.WriteString(field.nestedSchemaRef)
    } else {
        builder.WriteString("array of objects")
    }
} else if field.isObject {
    if field.nestedSchemaRef != "" {
        builder.WriteString(field.nestedSchemaRef)
    } else {
        builder.WriteString("object")
    }
} else {
```

#### 2. Update `renderSharedDefinitions` function

**File**: `convert.go`

**Changes**: Apply same pattern in shared definitions rendering (around line 349-358)

#### 3. Update `renderSchemaDefinition` function

**File**: `convert.go`

**Changes**: Apply same pattern in nested schema definition rendering (around line 1526-1531)

#### 4. Update version and regenerate golden files

**File**: `examples/example.md`, `testdata/golden/petstore-example.md`

**Changes**: Regenerate to reflect new format

### Testing Requirements

Existing tests that may require updates:
```go
func TestConvertFieldDefinitionsNested(t *testing.T)           // Update: "*(object, required)*" → "*(Metadata, required)*"
func TestConvertFieldDefinitionsArrayOfObjects(t *testing.T)   // Update: "*(array of objects)*" → "*(array of Item)*"
func TestCompleteExampleWithGoldenFile(t *testing.T)           // Update: regenerate golden file
```

**Test Objectives:**
- Verify `*(SchemaName)*` appears instead of `*(object)*`
- Verify `*(array of SchemaName)*` appears instead of `*(array of objects)*`
- Verify fallback to "object" when no schema ref is available (inline object)

**Edge Cases to Test:**
- Field with `$ref` shows schema name: `*(Address)*`
- Field with inline object (no `$ref`) shows: `*(object)*`
- Array of `$ref` objects shows: `*(array of Item)*`
- Array of inline objects shows: `*(array of objects)*`
- Nested schema refs maintain correct names at all levels

### Files to Regenerate

After all code changes are complete:
```bash
# Regenerate example markdown
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

# Copy to golden file
cp examples/example.md testdata/golden/petstore-example.md
```

### Validation
- [ ] Run: `go test ./...` in `openapi-markdown.go` directory
- [ ] Verify: All existing tests pass (after updating assertions)
- [ ] Verify: Golden file matches expected new format
- [ ] Run: `make ci` (if available) to run full CI checks

### Context for Implementation

- `schemaField.nestedSchemaRef` is already populated in `extractSchemaFields` and `extractSchemaFieldsFromProperties`
- The field stores the schema name extracted from the `$ref` path (e.g., "Address" from `#/components/schemas/Address`)
- No changes needed to field extraction logic, only to rendering
- When `nestedSchemaRef` is empty string, the field is an inline object without a `$ref`

---

## Dependency Order

```
Phase 1 (openapi-schema.go) ─┐
                             ├─→ Bump Version ─→ Phase 3 (openapi-markdown.go)
Phase 2 (openapi-schema.go) ─┘
```

- Phases 1 and 2 should be done together as a single release of `openapi-schema.go`
- Phase 3 depends on bumping the `openapi-schema.go` dependency version after Phases 1-2 are released
- **Important**: Phase 3 changes type display AND will benefit from the example improvements. Regenerate golden files only ONCE after all phases are complete to avoid double work.

## Version Bumping

After completing Phases 1-2:

```bash
# In openapi-schema.go directory:
1. Commit all changes
2. Tag new version (minor bump for new feature):
   git tag v0.8.0
   git push origin v0.8.0

# In openapi-markdown.go directory:
3. Update dependency:
   go get github.com/duh-rpc/openapi-schema.go@v0.8.0
   go mod tidy

4. Verify integration:
   go test -v -run TestCompleteExampleWithGoldenFile

5. Proceed with Phase 3
```

**Current Version Check:**
```bash
cd /Users/thrawn/Development/openapi-schema.go && git describe --tags --abbrev=0 2>/dev/null || echo "No tags yet"
```
