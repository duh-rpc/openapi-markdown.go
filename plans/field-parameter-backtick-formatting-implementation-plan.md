# Field and Parameter Backtick Formatting Implementation Plan

## Overview

Change the markdown output format for path parameters, query parameters, and top-level field definitions from bold (`**name**`) to backtick (`` `name` ``) formatting. This creates visual consistency across all field documentation while maintaining the existing hierarchical nesting structure.

## Current State Analysis

The markdown generator currently uses two different formatting styles:

**Bold format (`**name**`)** - Used for:
- Path parameters (convert.go:617-619)
- Query parameters (convert.go:674-676)
- Top-level field definitions in requests/responses (convert.go:1152-1154)
- Top-level fields in shared schemas (convert.go:344-346)

**Backtick format (`` `name` ``)** - Used for:
- Nested field definitions within objects (convert.go:1560-1562)

**Key Discoveries:**
- All formatting logic uses `strings.Builder` for efficient string construction
- Description separators differ by nesting level:
  - Top-level: space separator at convert.go:1177 (` `)
  - Nested: colon separator at convert.go:1583 (`: `)
- Enum prefix separators also differ:
  - Top-level: space before "Enums:" at convert.go:1185
  - Nested: period before "Enums:" at convert.go:1587
- Parameters do not support object types with nested properties
- Nested field rendering happens in `renderSchemaDefinition()` at convert.go:1553-1604

**Important Note on Example Files:**
- `examples/example.md` - Target file to regenerate with backtick formatting
- `examples/updated-example.md` - Shows a DIFFERENT format pattern (bold schema headers with backtick fields) - NOT the target for this implementation
- The CLAUDE.md file references `updated-example.md` as showing "canonical format" but this plan implements a DIFFERENT format where ALL fields use backticks (no bold headers)

## Desired End State

All field names and parameter names will use backtick formatting (`` `name` ``):
- Path parameters: `` - `paramName` (type, required) Description ``
- Query parameters: `` - `paramName` (type) Description ``
- Top-level field definitions: `` - `fieldName` (type, required) Description ``
- Nested fields: `` - `fieldName` (type): Description `` (unchanged)

**Verification:**
1. Run: `go run /tmp/regenerate_example.go` to regenerate examples/example.md
2. Compare output format - all field/parameter names should use backticks
3. Run: `go test ./...` - all tests pass
4. Verify: Golden file matches new format

## What We're NOT Doing

- NOT adding object-type support for path/query parameters
- NOT changing description separators (space vs colon) - these remain based on nesting level
- NOT changing enum prefix logic (space vs period)
- NOT changing the hierarchical nesting structure
- NOT changing header parameter table format (headers stay as tables)

## Implementation Approach

This is a straightforward find-and-replace task with test updates. The changes are localized to four formatting functions in convert.go and their corresponding test assertions. No architectural changes are needed - we're simply swapping the wrapping characters from `**` to `` ` ``.

## Phase 1: Update Path Parameter Formatting

### Overview
Change path parameter rendering from bold to backtick format.

### Changes Required:

#### 1. Path Parameter Rendering Function
**File**: `convert.go`

**Function signature:**
```go
func renderPathParametersFieldDef(builder *strings.Builder, params []v3.Parameter)
```

**Changes at lines 617-619:**

Replace:
```go
builder.WriteString("- **")
builder.WriteString(param.Name)
builder.WriteString("** (")
```

With:
```go
builder.WriteString("- `")
builder.WriteString(param.Name)
builder.WriteString("` (")
```

**Function responsibilities:**
- Iterate through path parameters
- Format each as: `` - `name` (type, required) Description ``
- Append inline enum values if present
- Maintain existing spacing and separator logic

#### 2. Update Path Parameter Tests
**File**: `convert_test.go`

**Test functions to update:**
```go
func TestConvertPathParameters(t *testing.T)
func TestConvertPathParametersFieldDef(t *testing.T)
```

**Test objectives:**
- Verify path parameters render with backtick format
- Check format: `` - `paramName` (type, required) Description ``
- Validate enum rendering still works correctly

**Context for implementation:**
- Find all assertions containing `**paramName**` and change to `` `paramName` ``
- Use `assert.Contains()` or `require.Contains()` for markdown string validation
- Pattern from existing tests: check for presence of formatted strings in result.Markdown

### Validation
- [ ] Run: `go test -run TestConvertPathParameters ./...`
- [ ] Verify: Test output shows backtick format for path parameters

## Phase 2: Update Query Parameter Formatting

### Overview
Change query parameter rendering from bold to backtick format.

### Changes Required:

#### 1. Query Parameter Rendering Function
**File**: `convert.go`

**Function signature:**
```go
func renderQueryParametersFieldDef(builder *strings.Builder, params []v3.Parameter)
```

**Changes at lines 674-676:**

Replace:
```go
builder.WriteString("- **")
builder.WriteString(param.Name)
builder.WriteString("** (")
```

With:
```go
builder.WriteString("- `")
builder.WriteString(param.Name)
builder.WriteString("` (")
```

**Function responsibilities:**
- Iterate through query parameters
- Format each as: `` - `name` (type, required) Description ``
- Append inline enum values if present
- Maintain existing spacing and separator logic

#### 2. Update Query Parameter Tests
**File**: `convert_test.go`

**Test functions to update:**
```go
func TestConvertQueryParameters(t *testing.T)
func TestConvertQueryParametersFieldDef(t *testing.T)
func TestConvertParameterWithEnums(t *testing.T)
func TestConvertParameterRequired(t *testing.T)
```

**Test objectives:**
- Verify query parameters render with backtick format
- Check format: `` - `paramName` (type) Description ``
- Validate enum rendering with backticks
- Verify required flag formatting

**Context for implementation:**
- Pattern to follow: Same approach as path parameter test updates
- All `**paramName**` assertions change to `` `paramName` ``
- Enum tests should verify backtick wrapping of values

### Validation
- [ ] Run: `go test -run TestConvertQueryParameters ./...`
- [ ] Run: `go test -run TestConvertParameter ./...`
- [ ] Verify: All query parameter tests pass with new format

## Phase 3: Update Top-Level Field Definition Formatting

### Overview
Change top-level field definitions in request/response bodies from bold to backtick format.

### Changes Required:

#### 1. Field Definitions Content Rendering
**File**: `convert.go`

**Function signature:**
```go
func renderFieldDefinitionsContent(builder *strings.Builder, schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage, sharedSchemas map[string]schemaUsage) error
```

**Changes at lines 1152-1154:**

Replace:
```go
builder.WriteString("- **")
builder.WriteString(field.name)
builder.WriteString("**")
```

With:
```go
builder.WriteString("- `")
builder.WriteString(field.name)
builder.WriteString("`")
```

**Function responsibilities:**
- Extract fields from schema proxy
- Format top-level fields as: `` - `name` (type, required) Description ``
- Render nested schema definitions recursively
- Handle arrays, objects, and primitive types
- Maintain space separator before description at line 1177 (DO NOT change to colon)
- Maintain space before "Enums:" at line 1185 (DO NOT change to period)

#### 2. Update Field Definition Tests
**File**: `convert_test.go`

**Test functions to update:**
```go
func TestConvertRequestBodyPOST(t *testing.T)
func TestConvertRequestBodyArrayField(t *testing.T)
func TestConvertRequestBodyEnumField(t *testing.T)
func TestConvertFieldDefinitionsNested(t *testing.T)
func TestConvertFieldDefinitionsDeeplyNested(t *testing.T)
func TestConvertFieldDefinitionsRecursive(t *testing.T)
func TestConvertFieldDefinitionsArrayOfObjects(t *testing.T)
func TestConvertFieldDefinitionsMaxDepth(t *testing.T)
func TestConvertFieldDefinitionsReferenceShared(t *testing.T)
func TestConvertResponseFieldDefinitions2xx(t *testing.T)
```

**Test objectives:**
- Verify top-level fields use backtick format
- Verify nested fields still use backtick format (unchanged)
- Validate hierarchical nesting structure preserved
- Check enum rendering correctness
- Verify array and object type formatting

**Context for implementation:**
- Change all top-level field assertions from `**fieldName**` to `` `fieldName` ``
- Nested field assertions already use backticks, should need no changes
- Pay attention to separator differences (space vs colon) in test assertions
- Follow pattern from convert_test.go:1656 and similar tests

### Validation
- [ ] Run: `go test -run TestConvertRequestBody ./...`
- [ ] Run: `go test -run TestConvertFieldDefinitions ./...`
- [ ] Run: `go test -run TestConvertResponseFieldDefinitions ./...`
- [ ] Verify: All field definition tests pass with new format

## Phase 4: Update Shared Schema Definition Formatting

### Overview
Change shared schema top-level field rendering from bold to backtick format.

### Changes Required:

#### 1. Shared Schema Rendering
**File**: `convert.go`

**Function signature:**
```go
func renderSharedDefinitions(builder *strings.Builder, sharedSchemas map[string]schemaUsage, model v3.Document, examples map[string]json.RawMessage) error
```

**Changes at lines 344-346:**

Replace:
```go
builder.WriteString("- **")
builder.WriteString(field.name)
builder.WriteString("**")
```

With:
```go
builder.WriteString("- `")
builder.WriteString(field.name)
builder.WriteString("`")
```

**Function responsibilities:**
- Render shared schemas used across multiple endpoints
- Format top-level fields with backtick wrapping
- Render nested definitions using existing `renderSchemaDefinition()`
- Maintain usage notes and anchor links
- Follow same formatting rules as request/response field definitions

**Context for implementation:**
- This code block is nearly identical to top-level field rendering in `renderFieldDefinitionsContent()`
- Both locations need the same bold-to-backtick change
- Nested schema rendering (via `renderSchemaDefinition()`) already uses backticks

#### 2. Update Shared Schema Tests
**File**: `convert_test.go`

**Test functions that may need updates:**
```go
func TestConvertFieldDefinitionsReferenceShared(t *testing.T)
```

**Test objectives:**
- Verify shared schema fields render with backtick format
- Validate "Used in:" notes still work correctly
- Check anchor links to shared schemas

**Context for implementation:**
- Look for test assertions that verify shared schema section format
- Change any `**fieldName**` to `` `fieldName` `` in shared schema context

### Validation
- [ ] Run: `go test -run TestConvertFieldDefinitionsReferenceShared ./...`
- [ ] Verify: Shared schema section renders correctly with backticks

## Phase 5: Regenerate Examples and Golden Files

### Overview
Update example output files and golden test files to reflect the new formatting.

### Changes Required:

#### 1. Regenerate Example Output
**Command to run:**
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

**Expected changes:**
- All path parameters change to backtick format
- All query parameters change to backtick format
- All top-level field definitions change to backtick format
- Nested fields remain unchanged (already use backticks)
- Headers remain in table format (unchanged)

#### 2. Update Golden File
**Command to run:**
```bash
cp examples/example.md testdata/golden/petstore-example.md
```

**Purpose:**
- Golden file is used by integration tests to verify complete conversion
- Must match the new output format for tests to pass

#### 3. Verify Golden File Test
**Test function:**
```go
func TestCompleteExampleWithGoldenFile(t *testing.T)
```

**Location:** convert_test.go:1336

**Context for implementation:**
- This test reads `testdata/golden/petstore-example.md` and compares against generated output
- Test will fail until golden file is updated with new format
- Regenerating golden file after code changes ensures this test passes

### Validation
- [ ] Run: `go test ./...`
- [ ] Verify: All tests pass including golden file comparisons
- [ ] Review: examples/example.md shows backtick format throughout

## Phase 6: Final Verification

### Overview
Run comprehensive tests and verify the complete formatting change.

### Validation Steps:

#### 1. Full Test Suite
```bash
go test ./...
```

**Expected outcome:**
- All tests pass
- No formatting regressions
- Golden file tests pass

#### 2. Manual Review of Generated Output
**File to review:** `examples/example.md`

**Check for:**
- Path parameters use backticks: `` - `userId` (string, required) ``
- Query parameters use backticks: `` - `status` (string) ``
- Top-level fields use backticks: `` - `petId` (string, required) ``
- Nested fields use backticks with proper indentation: `` - `nestedField` (type): Description ``
- Enum values wrapped in backticks: `` Enums: `val1`, `val2` ``
- Headers still use table format
- No stray bold formatting (`**`) for field/parameter names

#### 3. Verify Separator Logic Preserved
**Check in examples/example.md:**
- Top-level fields: space before description (no colon) - controlled by convert.go:1177
- Top-level fields: space before "Enums:" - controlled by convert.go:1185
- Nested fields: colon before description - controlled by convert.go:1583
- Nested fields: period before "Enums:" - controlled by convert.go:1587

**Critical:** The separator logic MUST remain unchanged. Only the field name wrapping changes from `**` to `` ` ``.

#### 4. Build and Lint
```bash
go build ./...
go vet ./...
```

**Expected outcome:**
- Clean build with no errors
- No linter warnings

### Validation
- [ ] Run: `go test ./...` - All tests pass
- [ ] Run: `go build ./...` - Clean build
- [ ] Review: examples/example.md format is correct
- [ ] Verify: Golden file matches new format

## Testing Strategy

**Approach:**
- Follow TDD principles: update tests first, then implementation
- Run tests incrementally after each phase
- Use `go test -run <pattern>` to focus on specific test groups
- Validate golden file at the end to ensure complete integration

**Test Coverage Areas:**
- Path parameter rendering
- Query parameter rendering
- Request body field definitions
- Response field definitions
- Nested object rendering
- Array field rendering
- Enum value rendering
- Shared schema definitions
- Golden file comparison

## Notes

**Consistency with CLAUDE.md:**
- Tests use `assert` and `require` from `github.com/stretchr/testify`
- No descriptive messages in assertions
- Tests are in `conv_test` package (separate from `conv`)
- Use table-driven tests where appropriate

**Git Status:**
- Working on branch: `thrawn/refactor-markdown`
- Main branch: `main`
- Current uncommitted changes: `M examples/example.md`

**Reference Files:**
- Current implementation: convert.go:595-1604
- Test patterns: convert_test.go
- Golden file test: TestCompleteExampleWithGoldenFile at convert_test.go:1336
- Example output: examples/example.md (currently partially edited to show desired format)
- Golden file: testdata/golden/petstore-example.md
- Separator logic:
  - Top-level description: convert.go:1177
  - Top-level enums: convert.go:1185
  - Nested description: convert.go:1583
  - Nested enums: convert.go:1587
