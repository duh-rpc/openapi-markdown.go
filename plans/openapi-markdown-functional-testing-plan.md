# OpenAPI to Markdown Converter Implementation Plan (Functional Testing Approach)

## Overview

This project creates a Go library that converts OpenAPI 3.x specifications to comprehensive markdown API documentation. Following **functional testing principles**, all tests interact exclusively through the public `Convert()` API, never calling internal functions directly.

## Functional Testing Compliance

This plan adheres to the functional-testing skill guidelines:

**Core Principles:**
- ✅ **Test public interface only**: All tests call `Convert()`, never internal functions
- ✅ **Incremental implementation**: Each phase extends `Convert()` with new functionality
- ✅ **Observability for internal behavior**: Debug fields in ConvertResult for testing internals
- ✅ **No internal function tests**: If a function can't be tested via public API, it's either removed or exposed through observability

**Reference**: See `functional-testing` skill and CLAUDE.md guidelines

## Current State Analysis

The openapi-markdown.go project has only a go.mod file. Reference projects:

**openapi-proto.go** (`/Users/thrawn/Development/openapi-proto.go`):
- Architecture pattern: Convert() function, internal packages, template generation
- Note: openapi-proto.go tests internal functions (violation of functional testing we'll avoid)

**scout** (`/Users/thrawn/Development/scout`):
- Target markdown format with TOC, tag sections, parameter tables, response examples

## Desired End State

A working library with:
```go
result, err := conv.Convert(openapi, conv.ConvertOptions{
    Title:       "My API",
    Description: "API documentation",
})
// result.Markdown contains generated markdown
// result.EndpointCount shows number of documented endpoints
// result.TagCount shows number of tag sections
```

**Validation:**
- Run: `make ci`
- Verify: All tests pass (testing ONLY via Convert())

## What We're NOT Doing

- Multi-file output
- Common Parameters section
- Request body documentation
- Authentication/security scheme documentation
- Testing internal functions directly (functional testing principle)

## Implementation Approach

**Incremental Public API Development:**
Each phase implements a working version of `Convert()` that can be tested through the public interface:
- Phase 1: Minimal Convert() - title and description only
- Phase 2: Basic Convert() - add table of contents and endpoint listings
- Phase 3: Complete Convert() - add parameter tables and response examples
- Phase 4: Production-ready Convert() - add observability and comprehensive validation

**Testing Strategy:**
All tests use the pattern:
```go
package conv_test

import (
    "testing"

    conv "github.com/duh-rpc/openapi-markdown.go"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConvertFeature(t *testing.T) {
    openapi := []byte(`openapi: 3.0.0...`)
    result, err := conv.Convert(openapi, conv.ConvertOptions{...})
    require.NoError(t, err)
    assert.Contains(t, string(result.Markdown), "expected content")
    assert.Equal(t, expectedCount, result.EndpointCount)
}
```

**Important:** All tests MUST use `package conv_test` (external test package) to ensure we only test through the public API.

## Phase 1: Minimal Working Convert() - Title and Description

### Overview
Create a minimal but **fully working** `Convert()` function that generates basic markdown (title + description only). This establishes the public interface that all subsequent phases will extend.

### Directory Structure
```
/Users/thrawn/Development/openapi-markdown.go/
├── go.mod
├── go.sum
├── convert.go          # Public API with working Convert()
├── convert_test.go     # Tests via Convert() only
├── Makefile
└── internal/
    └── parser/
        └── parser.go   # Internal implementation (not tested directly)
```

### Dependencies
```bash
go get github.com/pb33f/libopenapi@latest
go get github.com/stretchr/testify@latest
```

### Changes Required:

#### 1. Public API with Minimal Convert()
**File**: `convert.go`
**Changes**: Implement working Convert() that generates title + description

```go
package conv

// ConvertResult contains markdown output and generation metadata
type ConvertResult struct {
	Markdown      []byte
	EndpointCount int
	TagCount      int
}

// ConvertOptions configures markdown generation
type ConvertOptions struct {
	Title       string
	Description string
}

// Convert converts OpenAPI 3.x to markdown API documentation
func Convert(openapi []byte, opts ConvertOptions) (*ConvertResult, error)
```

**Implementation Requirements:**
- Validate inputs: non-empty openapi bytes, non-empty title
- Parse OpenAPI using libopenapi (wrap in internal/parser if needed, but don't test parser directly)
- Extract info section (title, description, version)
- Generate minimal markdown: `"# {title}\n\n{description}\n"`
- Return ConvertResult with Markdown populated, EndpointCount=0, TagCount=0

**Error Handling:**
- Return error for: empty openapi, empty title, invalid YAML/JSON, non-3.x OpenAPI
- Graceful degradation: empty description → use empty string in markdown

**Context for implementation:**
- Reference: openapi-proto.go/convert.go:73-150 for validation pattern
- Use libopenapi.NewDocument() and BuildV3Model() internally
- Keep internal/parser simple - it's just a wrapper, not tested directly

#### 2. Functional Tests via Public Interface
**File**: `convert_test.go`
**Changes**: Test Convert() behavior through public API only

```go
package conv_test

import (
    "testing"

    conv "github.com/duh-rpc/openapi-markdown.go"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConvertMinimalMarkdown(t *testing.T)
func TestConvertEmptyInput(t *testing.T)
func TestConvertEmptyTitle(t *testing.T)
func TestConvertInvalidOpenAPI(t *testing.T)
func TestConvertOpenAPI2Rejected(t *testing.T)
```

**Test Objectives:**
- Verify Convert() generates markdown with title
- Verify Convert() includes description when provided
- Verify Convert() rejects empty openapi bytes (separate test function)
- Verify Convert() rejects empty title (separate test function)
- Verify Convert() rejects invalid YAML/JSON (separate test function)
- Verify Convert() rejects OpenAPI 2.0 (separate test function)

**Note:** Create separate test functions for each error case (TestConvertEmptyInput, TestConvertEmptyTitle, etc.) to ensure clear, focused testing.

**Test Pattern (ALL tests use this approach):**
```go
func TestConvertMinimalMarkdown(t *testing.T) {
	for _, test := range []struct {
		name     string
		openapi  string
		opts     conv.ConvertOptions
		wantMd   string
		wantErr  string
	}{
		{
			name: "title and description",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  description: Test Description
  version: 1.0.0
paths: {}`,
			opts: conv.ConvertOptions{
				Title:       "Test API",
				Description: "Test Description",
			},
			wantMd: "# Test API\n\nTest Description\n",
		},
		{
			name: "title only",
			openapi: `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}`,
			opts: conv.ConvertOptions{
				Title: "Test API",
			},
			wantMd: "# Test API\n\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := conv.Convert([]byte(test.openapi), test.opts)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.wantMd, string(result.Markdown))
			assert.Equal(t, 0, result.EndpointCount)
			assert.Equal(t, 0, result.TagCount)
		})
	}
}
```

**Context for implementation:**
- Reference: openapi-proto.go/convert_test.go:27-119 for table-driven pattern
- Use testify/require for critical assertions
- Use testify/assert for non-critical checks
- NO tests for internal/parser functions

#### 3. Build Configuration
**File**: `Makefile`

```makefile
.PHONY: test lint tidy fmt coverage ci clean

test:
	go test -v ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy && git diff --exit-code

fmt:
	go fmt ./... && git diff --exit-code

ci: tidy fmt lint test
	@echo
	@echo "\033[32mEVERYTHING PASSED!\033[0m"

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	rm -f coverage.out coverage.html
	go clean
```

### Validation
- [ ] Run: `make test`
- [ ] Verify: All tests pass, testing ONLY via Convert()
- [ ] Verify: convert_test.go uses `package conv_test` (external test package)
- [ ] Verify: No imports of `internal/` packages in tests
- [ ] Verify: All test function names use camelCase (e.g., TestConvertMinimalMarkdown)
- [ ] Run: `make ci`
- [ ] Verify: All checks pass
- [ ] Manually verify: No internal function tests exist

## Phase 2: Basic Convert() - Table of Contents and Endpoint Listings

### Overview
Extend Convert() to extract operations from paths and generate a table of contents with basic endpoint listings. Still test ONLY through the public Convert() interface.

### Changes Required:

#### 1. Extend Convert() Implementation
**File**: `convert.go`
**Changes**: Add operation extraction and TOC generation to Convert()

**Implementation Requirements:**
- Extract paths and operations from OpenAPI document
- Group operations by tags (create "Default APIs" for untagged)
- Generate markdown with:
  - Title and description (from Phase 1)
  - Table of contents table with links
  - Tag section headings
  - Basic endpoint listings (method + path + summary)
- Update ConvertResult.EndpointCount and TagCount

**Internal Implementation Notes:**
- May create internal/builder.go for IR construction (not tested directly)
- May create helper functions for grouping by tags (not tested directly)
- All behavior verified through Convert() output

**Context for implementation:**
- Use PathItem.GetOperations().FromOldest() to preserve method order
- Handle multiple tags: duplicate endpoint in each tag section
- Preserve document order for paths and operations

#### 2. Functional Tests for TOC
**File**: `convert_test.go`
**Changes**: Add tests for TOC generation via Convert()

```go
func TestConvertTableOfContents(t *testing.T)
func TestConvertSingleEndpoint(t *testing.T)
func TestConvertMultipleEndpoints(t *testing.T)
func TestConvertUntaggedOperations(t *testing.T)
func TestConvertMultipleTagsPerOperation(t *testing.T)
```

**Test Objectives:**
- Verify TOC table generated with correct format
- Verify TOC links match endpoint anchors
- Verify endpoints listed under correct tag sections
- Verify untagged operations appear in "Default APIs"
- Verify operations with multiple tags appear in each section
- Verify EndpointCount and TagCount populated correctly

**Test Pattern:**
```go
func TestConvertTableOfContents(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Pet Store
  version: 1.0.0
paths:
  /pets:
    get:
      summary: List all pets
      tags:
        - pets
  /pets/{petId}:
    get:
      summary: Get a pet by ID
      tags:
        - pets
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Pet Store API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	// Verify title
	assert.Contains(t, markdown, "# Pet Store API")

	// Verify TOC exists
	assert.Contains(t, markdown, "## Table of Contents")
	assert.Contains(t, markdown, "HTTP Request | Description")

	// Verify TOC entries with links
	assert.Contains(t, markdown, "GET [/pets](#get-pets)")
	assert.Contains(t, markdown, "GET [/pets/{petId}](#get-petspetid)")

	// Verify tag section
	assert.Contains(t, markdown, "## pets")

	// Verify endpoints listed
	assert.Contains(t, markdown, "## GET /pets")
	assert.Contains(t, markdown, "List all pets")

	// Verify metadata
	assert.Equal(t, 2, result.EndpointCount)
	assert.Equal(t, 1, result.TagCount)
}
```

**Context for implementation:**
- All verification through markdown content and metadata
- No tests for internal groupByTags() or similar functions
- If internal behavior needs verification, use observability (see Phase 4)

### Validation
- [ ] Run: `make test`
- [ ] Verify: All tests pass, testing ONLY via Convert()
- [ ] Verify: TOC links work (manual check of generated markdown)
- [ ] Run: `make ci`
- [ ] Verify: All checks pass

## Phase 3: Complete Convert() - Parameter Tables and Response Examples

### Overview
Complete Convert() implementation by adding parameter tables and response examples. Continue testing ONLY through the public Convert() interface.

### Changes Required:

#### 1. Complete Convert() Implementation
**File**: `convert.go`
**Changes**: Add parameter and response extraction to Convert()

**Implementation Requirements:**
- Extract parameters (path, query, header) from operations
- Extract responses with status codes and schemas
- Generate markdown with full endpoint documentation:
  - Method and path heading
  - Description
  - Path parameters table (if any)
  - Query parameters table (if any)
  - Header parameters table (if any)
  - Response examples (if any)
- Parameter type mapping: string, integer, number, boolean, array, object, $ref
- Response schema extraction: prefer application/json, extract $ref names

**Internal Implementation Notes:**
- May create internal/generator.go for template rendering (not tested directly)
- May add helper functions for table formatting (not tested directly)
- May add makeAnchor() for link generation (not tested directly)

**Context for implementation:**
- Parameter type mapping rules from original plan
- Response schema extraction rules from original plan
- Anchor generation algorithm from original plan
- All verified through Convert() output

#### 2. Functional Tests for Complete Markdown
**File**: `convert_test.go`
**Changes**: Add tests for parameter tables and responses via Convert()

```go
func TestConvertPathParameters(t *testing.T)
func TestConvertQueryParameters(t *testing.T)
func TestConvertHeaderParameters(t *testing.T)
func TestConvertResponseExamples(t *testing.T)
func TestConvertCompleteEndpoint(t *testing.T)
```

**Test Objectives:**
- Verify path parameter tables generated correctly
- Verify query parameter tables generated correctly
- Verify header parameter tables generated correctly
- Verify response examples with status codes
- Verify complete endpoint with all sections
- Verify parameter type mapping correct
- Verify response schema extraction correct

**Test Pattern:**
```go
func TestConvertPathParameters(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets/{petId}:
    get:
      summary: Get a pet
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          description: ID of pet to return
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	// Verify path parameter table exists
	assert.Contains(t, markdown, "#### Path Parameters")
	assert.Contains(t, markdown, "Name | Description | Required | Type")
	assert.Contains(t, markdown, "petId | ID of pet to return | true | string")

	// Verify response section
	assert.Contains(t, markdown, "##### 200 Response")
	assert.Contains(t, markdown, "Successful response")
}
```

**Context for implementation:**
- Test markdown structure and content via assertions
- Test parameter type mapping through generated tables
- Test response extraction through generated examples
- NO tests for makeAnchor(), renderParamTable(), etc.

### Validation
- [ ] Run: `make test`
- [ ] Verify: All tests pass, testing ONLY via Convert()
- [ ] Manually review generated markdown for formatting
- [ ] Run: `make ci`
- [ ] Verify: All checks pass

## Phase 4: Production-Ready Convert() - Observability and Comprehensive Testing

### Overview
Add observability to ConvertResult for debugging/testing internal behavior without violating functional testing principles. Add comprehensive integration tests and documentation.

### Changes Required:

#### 1. Add Observability to ConvertResult
**File**: `convert.go`
**Changes**: Add optional debug information to ConvertResult

```go
// ConvertResult contains markdown output and generation metadata
type ConvertResult struct {
	Markdown      []byte
	EndpointCount int
	TagCount      int
	Debug         *DebugInfo  // Only populated if ConvertOptions.Debug = true
}

// DebugInfo provides visibility into internal conversion process for testing
type DebugInfo struct {
	ParsedPaths      int
	ExtractedOps     int
	TagsFound        []string
	UntaggedOps      int
	ParameterCounts  map[string]int  // "path", "query", "header" -> count
	ResponseCounts   map[string]int  // status code -> count
}

// ConvertOptions configures markdown generation
type ConvertOptions struct {
	Title       string
	Description string
	Debug       bool  // Enable debug information in result
}
```

**Implementation Requirements:**
- Populate DebugInfo when opts.Debug = true
- Track internal metrics during conversion
- Use DebugInfo for testing internal behavior without testing internal functions directly

**Context for implementation:**
- This allows verifying internal logic through the public interface
- Example: Test that parameter extraction works by checking DebugInfo.ParameterCounts
- Still testing via Convert(), just with more observability

#### 2. Tests Using Debug Observability
**File**: `convert_test.go`
**Changes**: Add tests that verify internal behavior via Debug field

```go
func TestConvertDebugParameterExtraction(t *testing.T)
func TestConvertDebugResponseExtraction(t *testing.T)
func TestConvertDebugTagGrouping(t *testing.T)
```

**Test Pattern:**
```go
func TestConvertDebugParameterExtraction(t *testing.T) {
	openapi := []byte(`openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
        - name: limit
          in: query
          schema:
            type: integer
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Test API",
		Debug: true,  // Enable debug info
	})
	require.NoError(t, err)
	require.NotNil(t, result.Debug)

	// Verify internal behavior through observability
	assert.Equal(t, 1, result.Debug.ParameterCounts["path"])
	assert.Equal(t, 1, result.Debug.ParameterCounts["query"])
	assert.Equal(t, 0, result.Debug.ParameterCounts["header"])
}
```

#### 3. Comprehensive Integration Tests
**File**: `convert_test.go`
**Changes**: Add integration tests with complete OpenAPI specs

```go
func TestConvertIntegrationPetStore(t *testing.T)
func TestConvertIntegrationComplexAPI(t *testing.T)
```

**Test Objectives:**
- Convert complete OpenAPI spec with multiple operations, tags, parameters
- Verify all markdown sections present
- Verify metadata accurate
- Verify debug information correct
- Use golden file pattern for markdown verification

**Test Pattern:**
```go
func TestConvertIntegrationPetStore(t *testing.T) {
	// Inline OpenAPI spec for test reliability (no external file dependencies)
	openapi := []byte(`openapi: 3.0.0
info:
  title: Pet Store API
  description: A sample API for managing pets
  version: 1.0.0
paths:
  /pets:
    get:
      summary: List all pets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: Maximum number of pets to return
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
  /pets/{petId}:
    get:
      summary: Get a pet by ID
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          description: ID of pet to return
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: Pet not found
`)

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title: "Pet Store API",
		Debug: true,
	})
	require.NoError(t, err)

	markdown := string(result.Markdown)

	// Verify complete structure
	assert.Contains(t, markdown, "# Pet Store API")
	assert.Contains(t, markdown, "## Table of Contents")
	assert.Contains(t, markdown, "## pets")

	// Verify metadata
	assert.Equal(t, 2, result.EndpointCount)
	assert.Equal(t, 1, result.TagCount)

	// Verify debug info
	require.NotNil(t, result.Debug)
	assert.Equal(t, 2, result.Debug.ParsedPaths)

	// Optional: Golden file testing
	// golden := filepath.Join("testdata", "petstore.golden.md")
	// if *update {
	//     os.WriteFile(golden, result.Markdown, 0644)
	// }
	// expected, _ := os.ReadFile(golden)
	// assert.Equal(t, string(expected), markdown)
}
```

**Note:** Inline OpenAPI specs in tests for reliability. For large specs, consider using `testdata/` directory with the golden file pattern shown above.

#### 4. Optional: Golden File Testing
**Directory**: `testdata/`

For comprehensive markdown verification, consider the golden file pattern:

```bash
# Update golden files when output format changes
go test -update

# Tests compare generated output against golden files
go test
```

**Example Test Data Structure:**
```
testdata/
├── petstore.yaml           # Input OpenAPI spec
└── petstore.golden.md      # Expected markdown output
```

#### 5. Example OpenAPI Specification (Optional)
**File**: `testdata/petstore.yaml`

```yaml
openapi: 3.0.0
info:
  title: Pet Store API
  description: A sample API for managing pets
  version: 1.0.0
paths:
  /pets:
    get:
      summary: List all pets
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: Maximum number of pets to return
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
  /pets/{petId}:
    get:
      summary: Get a pet by ID
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          description: ID of pet to return
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: Pet not found
```

#### 6. README Documentation
**File**: `README.md`

```markdown
# OpenAPI to Markdown Converter

A Go library that converts OpenAPI 3.x specifications to comprehensive markdown API documentation.

## Installation

​```bash
go get github.com/duh-rpc/openapi-markdown.go
​```

## Usage

### Basic Example

​```go
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
​```

### Debug Mode

​```go
result, err := conv.Convert(openapi, conv.ConvertOptions{
    Title: "My API",
    Debug: true,  // Enable debug information
})

fmt.Printf("Parsed %d paths\n", result.Debug.ParsedPaths)
fmt.Printf("Extracted %d operations\n", result.Debug.ExtractedOps)
​```

## Testing Philosophy

This library follows **functional testing principles**:
- All tests interact through the public `Convert()` API
- Internal functions are never tested directly
- Internal behavior is verified through observability (Debug field)
- Tests validate generated markdown content and metadata

## Development

### Running Tests
​```bash
make test
​```

### Running Full CI Suite
​```bash
make ci
​```

### Test Coverage
​```bash
make coverage
​```
```

### Validation
- [ ] Run: `make test`
- [ ] Verify: All tests pass, ALL via Convert()
- [ ] Verify: All test functions use camelCase naming (e.g., TestConvertIntegrationPetStore)
- [ ] Verify: Debug information populated correctly in DebugInfo fields
- [ ] Run: `make coverage`
- [ ] Verify: Coverage > 80%
- [ ] Run: `make ci`
- [ ] Verify: All checks pass
- [ ] Manually verify: Generated markdown readable and complete (check test output or golden files)
- [ ] Confirm: Zero tests of internal functions exist
- [ ] Confirm: All tests use `package conv_test`

## Phase 5: Complete Example with Golden File Verification

### Overview
Create a comprehensive example demonstrating the full capabilities of the converter with a complex OpenAPI specification and exact expected output. This serves as both documentation and a rigorous end-to-end test.

### Directory Structure
```
examples/
├── openapi.yaml    # Complex OpenAPI spec with all features
└── example.md      # Exact expected markdown output
```

### Changes Required:

#### 1. Complex OpenAPI Specification
**File**: `examples/openapi.yaml`

Create a comprehensive OpenAPI spec demonstrating all converter features:
- Multiple tags (pets, users, orders, admin)
- Mix of tagged and untagged operations
- Operations with multiple tags
- All parameter types: path, query, header
- All parameter data types: string, integer, boolean, array
- Required and optional parameters
- Multiple response codes (200, 201, 400, 404, 500)
- Rich descriptions and summaries
- Approximately 10-15 endpoints

**Complete Example:**
```yaml
openapi: 3.0.0
info:
  title: Pet Store API
  description: A comprehensive API for managing a pet store with users, pets, and orders
  version: 2.0.0

paths:
  /pets:
    get:
      summary: List all pets
      description: Returns a paginated list of all pets in the store
      tags:
        - pets
      parameters:
        - name: limit
          in: query
          description: Maximum number of pets to return
          required: false
          schema:
            type: integer
        - name: offset
          in: query
          description: Number of pets to skip
          required: false
          schema:
            type: integer
        - name: tag
          in: query
          description: Filter by tag
          required: false
          schema:
            type: string
      responses:
        '200':
          description: Successful response with pet list
        '400':
          description: Invalid request parameters

    post:
      summary: Create a new pet
      description: Adds a new pet to the store inventory
      tags:
        - pets
      parameters:
        - name: X-Request-ID
          in: header
          description: Unique request identifier
          required: true
          schema:
            type: string
      responses:
        '201':
          description: Pet created successfully
        '400':
          description: Invalid pet data

  /pets/{petId}:
    get:
      summary: Get a pet by ID
      description: Returns detailed information about a specific pet
      tags:
        - pets
      parameters:
        - name: petId
          in: path
          description: ID of pet to return
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: Pet not found

    delete:
      summary: Delete a pet
      description: Removes a pet from the store inventory
      tags:
        - pets
        - admin
      parameters:
        - name: petId
          in: path
          description: ID of pet to delete
          required: true
          schema:
            type: string
        - name: X-Admin-Token
          in: header
          description: Admin authorization token
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Pet deleted successfully
        '404':
          description: Pet not found
        '403':
          description: Unauthorized

  /users:
    get:
      summary: List all users
      description: Returns a list of registered users
      tags:
        - users
      parameters:
        - name: active
          in: query
          description: Filter by active status
          required: false
          schema:
            type: boolean
      responses:
        '200':
          description: Successful response

    post:
      summary: Create a new user
      description: Register a new user account
      tags:
        - users
      responses:
        '201':
          description: User created successfully
        '400':
          description: Invalid user data

  /users/{userId}:
    get:
      summary: Get user by ID
      description: Returns detailed user information
      tags:
        - users
      parameters:
        - name: userId
          in: path
          description: User identifier
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: User not found

  /users/{userId}/orders:
    get:
      summary: Get user orders
      description: Returns all orders placed by a specific user
      tags:
        - users
        - orders
      parameters:
        - name: userId
          in: path
          description: User identifier
          required: true
          schema:
            type: string
        - name: status
          in: query
          description: Filter by order status
          required: false
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: User not found

  /orders:
    get:
      summary: List all orders
      description: Returns a list of all orders in the system
      tags:
        - orders
      parameters:
        - name: limit
          in: query
          description: Maximum number of orders to return
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: Successful response

    post:
      summary: Create a new order
      description: Place a new order for pets
      tags:
        - orders
      responses:
        '201':
          description: Order created successfully
        '400':
          description: Invalid order data

  /orders/{orderId}:
    get:
      summary: Get order by ID
      description: Returns detailed order information
      tags:
        - orders
      parameters:
        - name: orderId
          in: path
          description: Order identifier
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
        '404':
          description: Order not found

  /health:
    get:
      summary: Health check endpoint
      description: Returns the health status of the API
      responses:
        '200':
          description: Service is healthy

  /metrics:
    get:
      summary: Get API metrics
      description: Returns usage metrics and statistics
      parameters:
        - name: X-Admin-Token
          in: header
          description: Admin authorization token
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Metrics data
        '403':
          description: Unauthorized
```

#### 2. Expected Markdown Output
**File**: `examples/example.md`

The complete expected markdown output (this will be quite long - showing exact format):

```markdown
# Pet Store API

A comprehensive API for managing a pet store with users, pets, and orders

## Table of Contents

HTTP Request | Description
-------------|------------
GET [/pets](#get-pets) | List all pets
POST [/pets](#post-pets) | Create a new pet
GET [/pets/{petId}](#get-petspetid) | Get a pet by ID
DELETE [/pets/{petId}](#delete-petspetid) | Delete a pet
GET [/users](#get-users) | List all users
POST [/users](#post-users) | Create a new user
GET [/users/{userId}](#get-usersuserid) | Get user by ID
GET [/users/{userId}/orders](#get-usersuseridorders) | Get user orders
GET [/orders](#get-orders) | List all orders
POST [/orders](#post-orders) | Create a new order
GET [/orders/{orderId}](#get-ordersorderid) | Get order by ID
GET [/health](#get-health) | Health check endpoint
GET [/metrics](#get-metrics) | Get API metrics

## pets

### GET /pets

List all pets

Returns a paginated list of all pets in the store

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
limit | Maximum number of pets to return | false | integer
offset | Number of pets to skip | false | integer
tag | Filter by tag | false | string

##### 200 Response

Successful response with pet list

##### 400 Response

Invalid request parameters

### POST /pets

Create a new pet

Adds a new pet to the store inventory

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Request-ID | Unique request identifier | true | string

##### 201 Response

Pet created successfully

##### 400 Response

Invalid pet data

### GET /pets/{petId}

Get a pet by ID

Returns detailed information about a specific pet

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to return | true | string

##### 200 Response

Successful response

##### 404 Response

Pet not found

### DELETE /pets/{petId}

Delete a pet

Removes a pet from the store inventory

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to delete | true | string

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Pet deleted successfully

##### 404 Response

Pet not found

##### 403 Response

Unauthorized

## users

### GET /users

List all users

Returns a list of registered users

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
active | Filter by active status | false | boolean

##### 200 Response

Successful response

### POST /users

Create a new user

Register a new user account

##### 201 Response

User created successfully

##### 400 Response

Invalid user data

### GET /users/{userId}

Get user by ID

Returns detailed user information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
userId | User identifier | true | string

##### 200 Response

Successful response

##### 404 Response

User not found

### GET /users/{userId}/orders

Get user orders

Returns all orders placed by a specific user

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
userId | User identifier | true | string

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
status | Filter by order status | false | string

##### 200 Response

Successful response

##### 404 Response

User not found

## orders

### GET /orders

List all orders

Returns a list of all orders in the system

#### Query Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
limit | Maximum number of orders to return | false | integer

##### 200 Response

Successful response

### POST /orders

Create a new order

Place a new order for pets

##### 201 Response

Order created successfully

##### 400 Response

Invalid order data

### GET /orders/{orderId}

Get order by ID

Returns detailed order information

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
orderId | Order identifier | true | string

##### 200 Response

Successful response

##### 404 Response

Order not found

## admin

### DELETE /pets/{petId}

Delete a pet

Removes a pet from the store inventory

#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
petId | ID of pet to delete | true | string

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Pet deleted successfully

##### 404 Response

Pet not found

##### 403 Response

Unauthorized

## Default APIs

### GET /health

Health check endpoint

Returns the health status of the API

##### 200 Response

Service is healthy

### GET /metrics

Get API metrics

Returns usage metrics and statistics

#### Header Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
X-Admin-Token | Admin authorization token | true | string

##### 200 Response

Metrics data

##### 403 Response

Unauthorized
```

#### 3. Golden File Verification Test
**File**: `convert_test.go`
**Changes**: Add test that verifies exact output match

```go
func TestConvertCompleteExample(t *testing.T) {
	// Load the example OpenAPI spec
	openapi, err := os.ReadFile("examples/openapi.yaml")
	require.NoError(t, err)

	// Load the expected markdown output
	expected, err := os.ReadFile("examples/example.md")
	require.NoError(t, err)

	// Convert the OpenAPI spec
	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title:       "Pet Store API",
		Description: "A comprehensive API for managing a pet store with users, pets, and orders",
	})
	require.NoError(t, err)

	// Assert the entire markdown output matches exactly
	assert.Equal(t, string(expected), string(result.Markdown))

	// Also verify metadata for documentation
	assert.Equal(t, 13, result.EndpointCount)
	assert.Equal(t, 4, result.TagCount) // pets, users, orders, admin (+ Default APIs)
}
```

**Test Objectives:**
- Verify complete end-to-end conversion with complex input
- Ensure exact markdown output matches expected format
- Serve as comprehensive documentation of converter behavior
- Catch any formatting regressions immediately

**Note:** This test provides a single source of truth for expected output. When the markdown format changes intentionally, update `examples/example.md` to reflect the new format.

### Validation
- [ ] Run: `make test`
- [ ] Verify: TestConvertCompleteExample passes with exact match
- [ ] Manually review: `examples/openapi.yaml` demonstrates all converter features
- [ ] Manually review: `examples/example.md` is correctly formatted and readable
- [ ] Verify: Test fails if even minor formatting changes occur (validates strict checking)
- [ ] Document: How to regenerate example.md when format changes are intentional

### Regenerating Expected Output

When markdown format changes are intentional (e.g., improving table formatting):

```bash
# Run the converter manually and save output
go run main.go examples/openapi.yaml > examples/example.md

# Or use a temporary test program
cat > /tmp/generate.go <<'EOF'
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
go run /tmp/generate.go
```

## Functional Testing Checklist

Before considering this plan complete, verify:

- [ ] **All tests use public API**: Every test calls `Convert()`, zero tests call internal functions
- [ ] **External test package**: All tests use `package conv_test`, not `package conv`
- [ ] **No internal imports**: Tests never import `internal/` packages
- [ ] **CamelCase test names**: All test functions use camelCase (e.g., TestConvertMinimalMarkdown, not TestConvert_MinimalMarkdown)
- [ ] **Internal behavior observable**: DebugInfo provides visibility without exposing internals
- [ ] **Incremental implementation**: Each phase has working `Convert()` that can be tested
- [ ] **TDD compatible**: Tests can be written before implementation
- [ ] **No test duplication**: Don't test same behavior at multiple levels
- [ ] **Inline test data**: Use inline OpenAPI specs in tests for reliability, `testdata/` for golden files only
- [ ] **Documentation mentions testing approach**: README explains functional testing philosophy
