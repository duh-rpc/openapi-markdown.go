# Response Examples Implementation Plan

## Overview

Enhance the OpenAPI to Markdown converter to include JSON response examples in the generated documentation. Currently, the converter only outputs response descriptions, but the desired format (based on the Scout project template) includes JSON code blocks showing example response bodies.

**IMPORTANT ARCHITECTURAL DECISION**: This implementation uses the `openapi-proto.go` library's `ConvertToExamples()` function to generate examples from schemas. Inline schemas in responses are NOT allowed and will cause an error - all response schemas must use `$ref` to reference schemas defined in `components/schemas`.

## Current State Analysis

### What Exists Now:
- Basic OpenAPI 3.x to Markdown conversion in `convert.go`
- `renderResponses()` function at `convert.go:301-327` that only renders:
  - Response status code headers
  - Response descriptions
- Table of Contents generation
- Parameter tables (path, query, header)
- Tag-based organization
- Test suite in `convert_test.go`
- Example OpenAPI spec at `examples/openapi.yaml` with:
  - One endpoint using `$ref` schemas: `/v3/pets.delete` (lines 75-83)
  - Component schemas defined (lines 259-274)
- Makefile with standard Go targets: `test`, `lint`, `tidy`, `fmt`, `ci`, `coverage`
- Dependency: openapi-proto.go v0.5.0 (provides ConvertToExamples)

### What's Missing:
- Integration with openapi-proto.go `ConvertToExamples()` function
- Response content extraction and example prioritization logic
- Support for OpenAPI `example` and `examples` fields in MediaType
- Validation that response schemas use `$ref` (reject inline schemas)
- JSON code block rendering in markdown output
- Response content/schemas for most endpoints in examples/openapi.yaml

### Key Discoveries:
- openapi-proto.go:343 - `ConvertToExamples()` generates JSON from component schemas
- openapi-proto.go:50-55 - `ExampleOptions` with IncludeAll flag for all schemas
- openapi-proto.go:44-47 - `ExampleResult` returns map[string]json.RawMessage
- openapi-proto.go/internal/examplegenerator.go:257-276 - Handles arrays with minItems/maxItems
- openapi-proto.go/internal/examplegenerator.go:106-112 - Handles objects and arrays
- libopenapi Response.Content is orderedmap of MediaType objects
- libopenapi MediaType has Example, Examples, and Schema fields
- MediaType.Schema is a SchemaProxy that can be a $ref or inline schema
- Scout template at `../scout/README.md:56-87` shows desired format with JSON examples

## Desired End State

### Specification:
When the converter processes OpenAPI responses, the generated markdown should include JSON examples in code blocks. The priority order for examples is:

1. **First**: Use explicit `example` from MediaType if present
2. **Second**: Use first entry from `examples` collection if present
3. **Third**: If schema is a `$ref`, generate example using `ConvertToExamples()`
4. **Error**: If schema is inline (not a `$ref`), abort with error message
5. **Fallback**: No code block if no content defined (description only)

### Verification:
```bash
# Run the example generation
go run ./cmd/openapi-markdown/main.go \
  -input examples/openapi.yaml \
  -output examples/example.md \
  -title "Pet Store API" \
  -description "A comprehensive API for managing a pet store with users, pets, and orders"

# Verify output matches scout template format
grep -A 5 "##### 200 Response" examples/example.md | head -20
# Should show JSON code blocks, not just descriptions

# All tests pass
make test

# Code quality passes
make ci
```

## What We're NOT Doing

- NOT supporting inline schemas in response content (must use $ref)
- NOT generating examples for non-JSON response types (only application/json)
- NOT adding request body examples (only responses)
- NOT changing the overall markdown structure or organization
- NOT adding configuration options for example generation (use sensible defaults)
- NOT implementing custom example generation logic (use openapi-proto.go)

## Implementation Approach

### Strategy:
Leverage the existing openapi-proto.go library for robust example generation:
1. Generate all component schema examples once using `ConvertToExamples()`
2. Cache the examples for use during response rendering
3. Check for explicit examples in MediaType first (developer intent)
4. For $ref schemas, look up the generated example by schema name
5. Reject inline schemas with a clear error message
6. Use libopenapi's ordered map patterns for iteration

### Reasoning:
- DRY: reuse battle-tested example generation from openapi-proto.go
- Consistency: examples match what would be used in protobuf/struct generation
- Robustness: handles arrays, enums, formats, nested objects, circular refs
- Simplicity: no need to maintain duplicate generation logic
- Validation: enforces best practice of using component schemas

## Phase 1: Core Response Example Rendering with openapi-proto Integration

### Overview
Implement response example extraction and rendering using openapi-proto.go's ConvertToExamples function. This phase adds the core functionality to generate and display JSON examples for API responses.

### Testing Strategy

**IMPORTANT**: All tests call the public `Convert()` function. We test response example behavior by:
1. Creating minimal OpenAPI specs with specific response configurations
2. Calling `Convert(spec, opts)`
3. Verifying the markdown output contains expected JSON examples
4. All internal helper functions are implementation details - NOT tested directly

The internal functions (`extractResponseExample`, `getExampleFromMediaType`, etc.) are tested
ONLY through their effect on the `Convert()` output.

### Changes Required:

#### 1. Update Imports
**File**: `convert.go`
**Changes**: Add necessary imports for example generation

```go
import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/duh-rpc/openapi-proto.go/conv"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"gopkg.in/yaml.v3"
)
```

**Note**: Use `conv` alias for openapi-proto to avoid conflict with this package name.

#### 2. Update Convert Function to Generate Examples
**File**: `convert.go`
**Location**: Update `Convert()` function around line 39

```go
// Convert converts OpenAPI 3.x to markdown API documentation
func Convert(openapi []byte, opts ConvertOptions) (*ConvertResult, error)
```

**Function Responsibilities:**
- After building v3Model, call `generateComponentExamples(openapi)` to get example map
- Pass examples map to `generateMarkdown()`
- Handle errors from example generation gracefully (log but don't fail)
- Follow existing error handling pattern from convert.go:48-64

**New Helper Function**:
```go
// generateComponentExamples generates JSON examples for all component schemas
func generateComponentExamples(openapi []byte) (map[string]json.RawMessage, error)
```

**Function Responsibilities:**
- Call `conv.ConvertToExamples()` with `IncludeAll: true`
- Use MaxDepth of 5 (openapi-proto default)
- Use Seed of 42 for deterministic examples
- Return empty map on error (graceful degradation)
- Pattern: `conv.ConvertToExamples(openapi, conv.ExampleOptions{IncludeAll: true, MaxDepth: 5, Seed: 42})`

**Testing**: This function is NOT tested directly. It is tested through its effect on the `Convert()` output.

#### 2b. Optional: Add Debug Information for Testing

To make it easier to verify internal behavior through the public API, optionally add:

**File**: `convert.go`
**Location**: Around line 30

```go
type ConvertResult struct {
    Markdown []byte
    Debug    *DebugInfo `json:",omitempty"`
}

type DebugInfo struct {
    // Map of endpoint ID to example source: "explicit", "named", "generated", "none"
    ExampleSources map[string]string
    // Counts for quick verification
    ExplicitCount  int
    NamedCount     int
    GeneratedCount int
}
```

Then tests can verify priority order explicitly:

```go
func TestConvert_ExampleSourceTracking(t *testing.T) {
    result, err := Convert(spec, ConvertOptions{Debug: true})
    require.NoError(t, err)
    require.NotNil(t, result.Debug)

    // Verify GET /v3/pets used explicit example
    assert.Equal(t, "explicit", result.Debug.ExampleSources["GET /v3/pets"])

    // Verify counts
    assert.Equal(t, 1, result.Debug.ExplicitCount)
    assert.Equal(t, 15, result.Debug.GeneratedCount)
}
```

**Note**: This is optional but recommended as it makes testing the priority system much clearer.

#### 3. Update generateMarkdown Signature
**File**: `convert.go`
**Location**: Line 156

```go
func generateMarkdown(opts ConvertOptions, endpoints []endpoint, tagGroups map[string][]endpoint, examples map[string]json.RawMessage) string
```

**Function Responsibilities:**
- Accept examples map parameter
- Pass examples to `renderResponses()` when rendering each endpoint
- Follow existing pattern from convert.go:156-232

**Testing**: This function is NOT tested directly. It is tested through its effect on the `Convert()` output.

#### 4. Update renderResponses Function
**File**: `convert.go`
**Location**: Replace function at line 304

```go
// renderResponses generates markdown for operation responses with JSON examples
func renderResponses(builder *strings.Builder, op *v3.Operation, examples map[string]json.RawMessage)
```

**Function Responsibilities:**
- Extract and sort response codes (existing logic)
- For each response:
  - Render status code header: `##### {code} Response\n`
  - Render description if present
  - Call `extractResponseExample(resp, examples)` to get JSON
  - If JSON present, render code block: `\n```json\n{json}\n```\n`
- Pattern reference: Follow spacing from scout template (blank line before code block)

**Testing**: This function is NOT tested directly. It is tested through its effect on the `Convert()` output.

#### 5. Example Extraction Logic
**File**: `convert.go`

```go
// extractResponseExample extracts or generates a JSON example for a response
func extractResponseExample(resp *v3.Response, examples map[string]json.RawMessage) string
```

**Function Responsibilities:**
- Return empty string if response has no Content
- Iterate Content ordered map looking for "application/json" media type
- For application/json MediaType:
  - Try `getExampleFromMediaType(mt)` first (explicit examples)
  - If no explicit example, try `getExampleFromSchema(mt.Schema, examples)`
  - Return first non-empty result
- Pattern: Use ordered map iteration `for pair := resp.Content.First(); pair != nil; pair = pair.Next()`

**Testing**: This function is NOT tested directly. It is tested through its effect on the `Convert()` output.

```go
// getExampleFromMediaType extracts explicit example from MediaType
func getExampleFromMediaType(mt *v3.MediaType) string
```

**Function Responsibilities:**
- Check `mt.Example` field (YAML node), decode and marshal to indented JSON
- If no Example, check `mt.Examples` collection, use first entry
- Use 3-space indentation: `json.MarshalIndent(value, "", "   ")`
- Return empty string on any decode/marshal errors
- Pattern: `mt.Example.Decode(&value)` then `json.MarshalIndent(value, "", "   ")`

**Testing**: This function is NOT tested directly. It is tested through its effect on the `Convert()` output.

```go
// getExampleFromSchema generates example from schema using pre-generated examples
func getExampleFromSchema(schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage) (string, error)
```

**Function Responsibilities:**
- If schemaProxy is nil, return empty string and nil error
- Check if schema is a reference: `schemaProxy.IsReference()`
- If not a reference (inline schema), return error: "inline schemas not supported in responses, use $ref"
- If reference, extract schema name from $ref using `extractSchemaName(schemaProxy.GetReference())`
- Look up schema name in examples map
- If found, format with 3-space indentation and return
- If not found, return empty string and nil error (graceful degradation)

**Testing**: This function is NOT tested directly. It is tested through its effect on the `Convert()` output.

```go
// extractSchemaName extracts schema name from $ref (e.g., "#/components/schemas/Pet" -> "Pet")
func extractSchemaName(ref string) (string, error)
```

**Function Responsibilities:**
- Parse $ref string to extract schema name
- Handle format: `#/components/schemas/{SchemaName}`
- Return error if format doesn't match expected pattern
- Pattern: `strings.TrimPrefix()` and `strings.Split()`

**Testing**: This function is NOT tested directly. It is tested through its effect on the `Convert()` output.

**Context for Implementation:**
- libopenapi SchemaProxy.IsReference() returns bool
- SchemaProxy.GetReference() returns string with full $ref path
- openapi-proto examples map uses schema name as key (not full $ref path)
- json.RawMessage can be reformatted with json.Indent() or marshal/unmarshal
- yaml.Node.Decode(&value) populates value with decoded YAML content

### Testing Requirements:

#### Tests to Create:

All tests call `Convert()` and verify markdown output. Tests validate:

**1. Example Priority Order** (table-driven test)
```go
func TestConvert_ResponseExamplePriority(t *testing.T) {
	for _, test := range []struct {
		name     string
		spec     []byte
		wantJSON string // Expected JSON in markdown
	}{
		{
			name: "ExplicitExample",
			spec: []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
              example:
                id: "123"
                text: "explicit example"
components:
  schemas:
    Message:
      type: object
      properties:
        id:
          type: string
        text:
          type: string
`),
			wantJSON: `"id": "123"`,
		},
		{
			name: "NamedExamples",
			spec: []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
              examples:
                example1:
                  value:
                    id: "456"
                    text: "named example"
components:
  schemas:
    Message:
      type: object
      properties:
        id:
          type: string
        text:
          type: string
`),
			wantJSON: `"id": "456"`,
		},
		{
			name: "GeneratedFromRef",
			spec: []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
components:
  schemas:
    Message:
      type: object
      properties:
        text:
          type: string
`),
			wantJSON: `"text": ""`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := Convert(test.spec, ConvertOptions{})
			require.NoError(t, err)
			assert.Contains(t, string(result.Markdown), test.wantJSON)
		})
	}
}
```

**2. Error Handling** (table-driven test)
```go
func TestConvert_ResponseSchemaErrors(t *testing.T) {
	for _, test := range []struct {
		name    string
		spec    []byte
		wantErr string
	}{
		{
			name: "InlineSchemaNotAllowed",
			spec: []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          content:
            application/json:
              schema:
                type: object
                properties:
                  text:
                    type: string
`),
			wantErr: "inline schemas not supported in responses",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := Convert(test.spec, ConvertOptions{})
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}
```

**3. Content Type Handling** (table-driven test)
```go
func TestConvert_ResponseContentTypes(t *testing.T) {
	for _, test := range []struct {
		name         string
		spec         []byte
		wantContains string
		wantMissing  string
	}{
		{
			name: "NoContentDescriptionOnly",
			spec: []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success
`),
			wantContains: "Success",
			wantMissing:  "```json",
		},
		{
			name: "MultipleMediaTypesOnlyJSON",
			spec: []byte(`
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Message'
            text/html:
              schema:
                type: string
components:
  schemas:
    Message:
      type: object
      properties:
        text:
          type: string
`),
			wantContains: "```json",
			wantMissing:  "```html",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := Convert(test.spec, ConvertOptions{})
			require.NoError(t, err)
			markdown := string(result.Markdown)
			if test.wantContains != "" {
				assert.Contains(t, markdown, test.wantContains)
			}
			if test.wantMissing != "" {
				assert.NotContains(t, markdown, test.wantMissing)
			}
		})
	}
}
```

**Test Objectives:**
- Verify explicit examples prioritized over schema generation (through markdown output)
- Validate JSON formatting via 3-space indentation (check markdown contains properly formatted JSON)
- Ensure responses without content don't render code blocks (check markdown lacks ```json)
- Confirm inline schemas cause error from `Convert()`
- Test that only application/json media type appears in markdown
- All verification through public `Convert()` API

### Validation Commands:

```bash
# Build should succeed
go build ./...

# Tests should pass
make test

# Formatting check
make fmt

# Full CI validation
make ci
```

**Success Criteria:**
- No compilation errors
- All existing tests continue to pass
- New tests validate response examples through `Convert()` public API
- NO tests call internal functions (`extractResponseExample`, `getExampleFromMediaType`, etc.)
- Inline schemas in responses cause `Convert()` to return error
- Generated markdown contains JSON examples with 3-space indentation
- Code follows existing patterns (ordered map iteration, error handling)
- All tests in `package convert_test` (external test package)
- All tests follow naming pattern: `TestConvert_FeatureName`

---

## Phase 2: Update Example OpenAPI Spec with Response Schemas

### Overview
Enhance `examples/openapi.yaml` to include response content with $ref schemas for all JSON-returning endpoints, demonstrating the full capability of the response example rendering.

### Changes Required:

#### 1. Define Response Schemas in components/schemas
**File**: `examples/openapi.yaml`
**Location**: Expand components/schemas section (currently lines 259-274)

**New Schemas to Add:**
```yaml
components:
  schemas:
    # Existing schemas
    PetDeleteRequest:  # already exists
    PetDeleteResponse:  # already exists

    # New response schemas
    PetList:
      type: object
      properties:
        pets:
          type: array
          items:
            $ref: '#/components/schemas/Pet'
        cursor:
          type: string

    Pet:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        status:
          type: string
          enum: [available, pending, sold]
        tags:
          type: array
          items:
            type: string

    PetCreated:
      type: object
      properties:
        id:
          type: string
        message:
          type: string

    UserList:
      type: object
      properties:
        users:
          type: array
          items:
            $ref: '#/components/schemas/User'

    User:
      type: object
      properties:
        id:
          type: string
        username:
          type: string
        email:
          type: string
          format: email
        active:
          type: boolean

    UserCreated:
      type: object
      properties:
        id:
          type: string
        message:
          type: string

    OrderList:
      type: object
      properties:
        orders:
          type: array
          items:
            $ref: '#/components/schemas/Order'
        cursor:
          type: string

    Order:
      type: object
      properties:
        id:
          type: string
        userId:
          type: string
        petId:
          type: string
        status:
          type: string
          enum: [placed, approved, delivered]
        quantity:
          type: integer

    OrderCreated:
      type: object
      properties:
        id:
          type: string
        message:
          type: string

    HealthStatus:
      type: object
      properties:
        status:
          type: string
          enum: [healthy, degraded, down]
        timestamp:
          type: string
          format: date-time

    Metrics:
      type: object
      properties:
        requestsTotal:
          type: integer
        requestsPerSecond:
          type: number
        uptime:
          type: integer

    ErrorResponse:
      type: object
      properties:
        error:
          type: string
        code:
          type: string
```

#### 2. Add Response Content to Endpoints
**File**: `examples/openapi.yaml`
**Changes**: Add content sections to responses for each endpoint

**Endpoints to Update:**

**GET /v3/pets** (lines 10-37):
```yaml
responses:
  '200':
    description: Successful response with pet list
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/PetList'
        example:  # Explicit example to demonstrate priority
          pets:
            - id: "pet-123"
              name: "Fluffy"
              status: "available"
              tags: ["cat", "friendly"]
          cursor: "eyJpZCI6InBldC0xMjMifQ=="
  '400':
    description: Invalid request parameters
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/ErrorResponse'
```

**POST /v3/pets** (lines 39-57):
```yaml
responses:
  '201':
    description: Pet created successfully
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/PetCreated'
  '400':
    description: Invalid pet data
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/ErrorResponse'
```

**POST /v3/pets.delete** (already has content, lines 77-87) - no changes needed

**GET /v3/pets/{petId}** (lines 89-107):
```yaml
responses:
  '200':
    description: Successful response
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Pet'
  '404':
    description: Pet not found
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/ErrorResponse'
```

**Similar patterns for**:
- GET /v3/users → UserList
- POST /v3/users → UserCreated
- GET /v3/users/{userId} → User
- GET /v3/users/{userId}/orders → OrderList
- GET /v3/orders → OrderList
- POST /v3/orders → OrderCreated
- GET /v3/orders/{orderId} → Order
- GET /v3/health → HealthStatus
- GET /v3/metrics → Metrics

**Context for Implementation:**
- Follow existing pattern from /v3/pets.delete (lines 75-87)
- Use $ref for all schemas (no inline schemas)
- Add explicit example to one endpoint (GET /v3/pets) to demonstrate priority
- Use examples collection on another endpoint to show that feature
- Let others use schema-based generation

### Testing Requirements:

#### Existing Tests to Verify:

All existing tests should continue to pass with richer output:
```go
func TestConvertPathParameters(t *testing.T)      // Now shows JSON examples
func TestConvertQueryParameters(t *testing.T)     // Now shows JSON examples
func TestConvertMultipleResponses(t *testing.T)   // Shows JSON for each response
func TestConvertHeaderParameters(t *testing.T)    // Now shows JSON examples
```

**Test Objectives:**
- Existing tests continue to pass (backward compatibility)
- Generated markdown now includes JSON examples for all responses with content
- No breaking changes to markdown structure
- Table of contents still works
- Tag grouping still works

### Validation Commands:

```bash
# Validate YAML syntax
yamllint examples/openapi.yaml

# Regenerate the example markdown
go run ./cmd/openapi-markdown/main.go \
  -input examples/openapi.yaml \
  -output examples/example.md \
  -title "Pet Store API" \
  -description "A comprehensive API for managing a pet store with users, pets, and orders"

# Count JSON code blocks (should be ~30+ for all success/error responses)
grep -c '```json' examples/example.md

# Verify explicit example is used for GET /v3/pets
grep -A 15 "GET /v3/pets" examples/example.md | grep -A 5 "##### 200 Response"
# Should show the explicit example with "Fluffy"

# All tests pass
make test
```

**Success Criteria:**
- examples/example.md contains JSON code blocks for all responses with content
- Format matches scout template (description, then ```json block)
- Explicit example is used where provided
- All response types are represented (lists, single objects, errors)
- Generated examples are valid JSON with proper indentation

---

## Phase 3: Add Golden File Test for Complete Example

### Overview
Add a comprehensive test that validates the complete output format against a golden file, ensuring generated markdown matches the scout template format and providing regression protection.

### Changes Required:

#### 1. Create Golden File
**File**: `testdata/golden/petstore-example.md`
**Changes**: Create expected output file

**Creation Process:**
1. Run Phase 1 and Phase 2 first (golden file depends on working implementation)
2. Generate initial version:
   ```bash
   go run ./cmd/openapi-markdown/main.go \
     -input examples/openapi.yaml \
     -output testdata/golden/petstore-example.md \
     -title "Pet Store API" \
     -description "A comprehensive API for managing a pet store with users, pets, and orders"
   ```
3. Manually verify format matches scout template:
   - Check JSON indentation (3 spaces)
   - Verify response sections have proper spacing
   - Confirm table of contents works
   - Validate all examples are present

**Golden File Structure:**
- Title and description
- Table of contents with all endpoints
- Tag sections (admin, orders, pets, users, Default APIs)
- Each endpoint with parameters and responses
- JSON examples for all responses that have content
- Matches scout template format exactly

#### 2. Add Golden File Test
**File**: `convert_test.go`

```go
// TestCompleteExampleWithGoldenFile validates complete conversion against golden file
func TestCompleteExampleWithGoldenFile(t *testing.T)
```

**Function Responsibilities:**
- Read examples/openapi.yaml using os.ReadFile
- Convert with full title and description
- Read golden file from testdata/golden/petstore-example.md
- Compare generated output with golden file using bytes.Equal
- On mismatch, write actual output to testdata/golden/petstore-example.actual.md
- Provide helpful error message with diff command

**Test Implementation Pattern:**
```go
func TestCompleteExampleWithGoldenFile(t *testing.T) {
	// Read input spec
	spec, err := os.ReadFile("examples/openapi.yaml")
	require.NoError(t, err)

	// Convert
	result, err := Convert(spec, ConvertOptions{
		Title:       "Pet Store API",
		Description: "A comprehensive API for managing a pet store with users, pets, and orders",
	})
	require.NoError(t, err)

	// Read golden file
	golden, err := os.ReadFile("testdata/golden/petstore-example.md")
	require.NoError(t, err)

	// Compare
	if !bytes.Equal(result.Markdown, golden) {
		// Write actual for debugging
		actualPath := "testdata/golden/petstore-example.actual.md"
		err := os.WriteFile(actualPath, result.Markdown, 0644)
		require.NoError(t, err)

		t.Fatalf("Output doesn't match golden file\nActual written to: %s\nRun: diff %s testdata/golden/petstore-example.md",
			actualPath, actualPath)
	}
}
```

#### 3. Add .gitignore Entry
**File**: `.gitignore`
**Changes**: Add pattern to ignore .actual.md files

```
testdata/golden/*.actual.md
```

**Context for Implementation:**
- Use os.ReadFile and os.WriteFile from standard library
- bytes.Equal for exact comparison
- Provide helpful error messages for debugging
- Golden file should be committed to repo
- Actual files should be ignored (for local debugging only)

### Testing Requirements:

#### New Test:

```go
func TestCompleteExampleWithGoldenFile(t *testing.T)
```

**Test Objectives:**
- Validate complete end-to-end conversion
- Ensure format consistency over time
- Catch unintended changes to output format
- Provide regression protection
- Make it easy to review changes (diff command in error message)

**Testing the Test:**
```bash
# Initial run - will create .actual.md file
go test -v -run TestCompleteExampleWithGoldenFile

# If output is correct, copy to golden:
cp testdata/golden/petstore-example.actual.md testdata/golden/petstore-example.md

# Verify test passes
go test -v -run TestCompleteExampleWithGoldenFile

# When making intentional changes to format:
# 1. Make changes
# 2. Run test (fails, creates .actual.md)
# 3. Review diff: diff testdata/golden/petstore-example.{actual,}.md
# 4. If correct: cp testdata/golden/petstore-example.actual.md testdata/golden/petstore-example.md
# 5. Commit updated golden file
```

### Validation Commands:

```bash
# Run all tests including golden file test
make test

# Run only golden file test
go test -v -run TestCompleteExampleWithGoldenFile

# Create testdata directory if needed
mkdir -p testdata/golden

# Verify golden file is tracked
git status testdata/

# Check that .actual.md is ignored
touch testdata/golden/test.actual.md
git status testdata/
# Should not show test.actual.md

# Full CI validation
make ci
```

**Success Criteria:**
- Golden file test passes
- Generated markdown matches scout template format
- Test provides clear error messages on mismatch
- Actual output file helps debug differences
- All existing tests continue to pass
- Golden file is committed to repository
- .actual.md files are gitignored

---

## Phase 4: Documentation and Examples Update

### Overview
Update project documentation to reflect the new response example capabilities, the $ref requirement, and provide clear guidance on how to use the feature.

### Changes Required:

#### 1. Update README
**File**: `README.md`

**Sections to Add/Update:**

```markdown
## Features

- Converts OpenAPI 3.x specifications to clean, readable Markdown
- Generates Table of Contents with anchor links
- Organizes endpoints by tags
- Renders parameter tables (path, query, header)
- **Generates JSON response examples from schemas**
- Supports explicit examples and schema-based generation
- Validates that response schemas use $ref (no inline schemas)

## Response Examples

The converter automatically generates JSON examples for API responses using three priority levels:

1. **Explicit examples**: Uses `example` field from response media type
2. **Named examples**: Uses first entry from `examples` collection
3. **Schema-based**: Generates from $ref schema using openapi-proto.go library

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
```

### Generated Markdown

```markdown
##### 200 Response

List of pets

```json
{
   "pets": [
      {
         "id": "123",
         "name": "Fluffy"
      }
   ]
}
```
```

## Requirements

- Go 1.25.4 or later
- Dependencies:
  - github.com/pb33f/libopenapi v0.28.2 (OpenAPI parsing)
  - github.com/duh-rpc/openapi-proto.go v0.5.0 (Example generation)
  - github.com/stretchr/testify v1.11.1 (Testing)

## Usage

\`\`\`bash
# Basic usage
go run ./cmd/openapi-markdown/main.go \\
  -input openapi.yaml \\
  -output api-docs.md \\
  -title "My API" \\
  -description "API Documentation"

# Or as a library
import conv "github.com/duh-rpc/openapi-markdown.go"

result, err := conv.Convert(specBytes, conv.ConvertOptions{
    Title:       "My API",
    Description: "API Documentation",
})
```

#### 2. Create CHANGELOG
**File**: `CHANGELOG.md`

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- JSON response examples in generated markdown documentation
- Support for OpenAPI `example` and `examples` fields in response media types
- Automatic example generation from response schemas using openapi-proto.go
- Validation that response schemas use $ref (rejects inline schemas)
- Three-tier priority system for example selection (explicit → named → generated)
- Golden file test for regression protection
- Comprehensive test suite for response example handling

### Changed
- `renderResponses()` now generates JSON code blocks for responses with content
- `generateMarkdown()` signature includes examples map parameter

### Dependencies
- Added: github.com/duh-rpc/openapi-proto.go v0.5.0 for example generation
```

#### 3. Create Examples README
**File**: `examples/README.md`

```markdown
# OpenAPI to Markdown Examples

This directory contains example OpenAPI specifications and their generated markdown output.

## Files

- `openapi.yaml` - Complete OpenAPI 3.0 specification for a pet store API
- `example.md` - Generated markdown documentation from openapi.yaml

## Generating the Example

To regenerate `example.md`:

```bash
go run ../cmd/openapi-markdown/main.go \
  -input openapi.yaml \
  -output example.md \
  -title "Pet Store API" \
  -description "A comprehensive API for managing a pet store with users, pets, and orders"
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
```

**Context for Implementation:**
- README.md likely exists, add new sections for response examples
- CHANGELOG.md should follow keepachangelog.com format
- examples/README.md is new, provides context for the example spec
- Document the $ref requirement prominently (it's a breaking constraint)

### Testing Requirements:

**No new tests required** - documentation only phase

**Manual Validation:**
- Read through documentation for clarity
- Verify code examples are correct and runnable
- Check markdown rendering in GitHub preview
- Ensure all commands in documentation work as shown
- Validate links and references

### Validation Commands:

```bash
# Test documented commands actually work
go run ./cmd/openapi-markdown/main.go \
  -input examples/openapi.yaml \
  -output /tmp/test-output.md \
  -title "Pet Store API" \
  -description "Test generation"

# Verify output matches documented format
cat /tmp/test-output.md | grep -A 10 "##### 200 Response" | head -20

# Check for broken markdown links (if markdown-link-check is installed)
npx markdown-link-check README.md CHANGELOG.md examples/README.md

# Verify YAML examples are valid
python3 -c "import yaml; yaml.safe_load(open('README.md').read().split('```yaml')[1].split('```')[0])"

# Build and test to ensure everything still works
make ci
```

**Success Criteria:**
- Documentation clearly explains response example feature
- $ref requirement is prominently documented
- Examples are accurate and work as shown
- Users understand how to provide examples in their OpenAPI specs
- Priority order is clearly explained
- All links and references are valid
- Markdown renders correctly on GitHub
- No broken code examples or commands

---

## Summary

This plan implements JSON response example generation in four focused phases:

1. **Phase 1**: Core implementation using openapi-proto.go with helper functions and tests
2. **Phase 2**: Enhanced example OpenAPI spec with $ref schemas demonstrating all features
3. **Phase 3**: Golden file test for regression protection and format validation
4. **Phase 4**: Complete documentation for users including $ref requirement

Each phase is independently testable and adds tangible value. The implementation leverages the existing openapi-proto.go library for robust example generation and enforces best practices by requiring $ref usage.

**Key Architectural Decision**: Using openapi-proto.go's ConvertToExamples ensures consistency with the proto/struct generation features while providing robust handling of arrays, enums, formats, nested objects, and circular references.
