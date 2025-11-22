# Field Definitions Format Implementation Plan

## Overview

This plan converts the openapi-markdown.go library to use a new Field Definitions format for documenting request bodies, responses, and parameters. The new format provides richer schema documentation including nested objects, enums, required fields, and shared schema definitions.

## Current State Analysis

### Existing Functionality
- **Document structure**: Title, description, table of contents, tag-based grouping (`convert.go:168-275`)
- **Parameters**: Rendered as tables for path, query, and header parameters (`convert.go:277-342`)
- **Responses**: Show status code, description, and JSON example (`convert.go:344-380`)
- **Request bodies**: NOT IMPLEMENTED - completely ignored
- **Schema documentation**: NOT IMPLEMENTED - only JSON examples shown, no field explanations

### Key Discoveries
- Single file implementation in `convert.go` with all markdown generation logic
- Uses `openapi-schema.go` library for generating component examples (`convert.go:424-439`)
- Three-tier priority for response examples: explicit > named > schema-generated (`convert.go:519-552`)
- Functional testing pattern: all tests use public `Convert()` API (`convert_test.go`)
- Response schemas MUST use `$ref` - inline schemas rejected (`convert.go:462-464`)

## Desired End State

### Updated Output Format

**Document structure** (unchanged):
- Title and description
- Table of contents
- Tag-based grouping

**Endpoint format** (new):
```markdown
### POST /v1/resource

Endpoint description

#### Headers

Name | Description | Required | Type
-----|-------------|----------|-----
X-Token | Auth token | true | string

#### Path Parameters

**resourceId** (string, required)
- The unique identifier for the resource

#### Query Parameters

**limit** (integer)
- Maximum number of items to return

### Request

```json
{ "field": "value" }
```

#### Field Definitions

**field** (string, required)
- Field description. Enums: `VALUE1`, `VALUE2`

### Responses

#### 200 Response

Success description

```json
{ "result": "value" }
```

#### Field Definitions

**result** (string)
- Result description

#### 400 Response

```json
{ "error": { "code": "INVALID" } }
```
```

**Shared schemas** (new):
- "Shared Schema Definitions" section at document level
- Referenced from endpoints to avoid duplication

**Verification**:
```bash
go run /tmp/regenerate_example.go
diff examples/example.md examples/updated-example.md
go test ./...
```

### Field Definition Format Rules

The following rules determine how to format field definitions:

**1. Top-level object fields** (object properties at schema root):
- Format: Bold header `**fieldName**` followed by bulleted properties
- Children: Bulleted list with backtick-wrapped names, type, required status, description
- Example from `examples/updated-example.md:46-48`:
  ```markdown
  **principal**
  - `type` (string, required): The type of principal. Enums: `API_KEY`, `USER`
  - `id` (string, required): The identifier for the principal
  ```

**2. Nested object references** (when a field's type is object):
- In parent: `- 'fieldName' (object, required): Description of what this object represents`
- Separate section: Create bold header section for the nested schema with its own properties
- Example from `examples/updated-example.md:50-57`:
  ```markdown
  **resource**
  - `nested` (object, required): A description of the nested thing

  **nested**
  - `type` (string, required): A description of this field
  - `name` (string): A description of this field
  ```

**3. Array fields**:
- **Arrays of primitives**: `**fieldName** (type array, required) Description. Enums: 'VAL1', 'VAL2'`
- **Arrays of objects**: `**fieldName** (array of objects)` with bullet list showing object properties inline
- Example (primitives) from `examples/updated-example.md:59`:
  ```markdown
  **actions** (string array, required) A description. Enums: `VAULTS_DETAILS_VIEW`, `WITHDRAW_INITIATE`
  ```
- Example (objects) from `examples/updated-example.md:89-92`:
  ```markdown
  **results** (array of objects)
  - One entry per requested action, in the same order as the request
  - `action` (string): The action that was checked
  - `allowed` (boolean): Whether this specific action is allowed
  ```

**4. Primitive fields in nested sections** (non-object properties):
- Format: `- 'fieldName' (type, required): Description. Enums: 'VAL1', 'VAL2'`
- Used within bold header sections to document individual properties
- Example from `examples/updated-example.md:47-48, 56-57`

**5. Parameter fields** (path, query):
- Format: `**paramName** (type, required)` with description as bullet
- Include enums if present
- Example:
  ```markdown
  **limit** (integer)
  - Maximum number of items to return. Enums: `10`, `25`, `50`, `100`
  ```

### Key Requirements

1. **Request bodies**: Render for POST/PUT/PATCH/DELETE with RequestBody (not GET) with JSON example and field definitions
2. **Field definitions**: Hierarchical structure with separate sections for nested objects (follow format rules above)
3. **Parameters**: Path and Query use field definitions format; Headers remain table format
4. **Responses**: Add "Responses" section header; field definitions only for 2xx responses
5. **Shared schemas**: Document once at document level if used across multiple endpoints (not same-endpoint reuse)
6. **Nested objects**: Create separate definition section for each nested schema (max depth: 10)
7. **Arrays**: Objects in arrays should be named component schemas; primitives shown inline with enums
8. **Recursion**: Cap recursive schema traversal at depth 1 (same schema in ancestry chain)

## What We're NOT Doing

- No backward compatibility mode - complete format replacement
- No support for OpenAPI 2.0 (already rejected)
- No changes to tag grouping behavior
- No changes to table of contents generation
- No changes to document title/description rendering

## Implementation Approach

### Strategy

Follow TDD approach:
1. Write tests for new behavior first
2. Implement functionality to make tests pass
3. Refactor while keeping tests green
4. Update integration tests and golden files last

### Technical Considerations

- Extend existing schema traversal using `libopenapi` base.Schema access patterns
- Reuse `generateComponentExamples()` for request body JSON generation
- Build schema definition tree to identify shared schemas across endpoints
- Track visited schemas with depth counter to prevent infinite recursion
- Use `strings.Builder` pattern for markdown construction (existing pattern)

## Phase 1: Request Body Rendering Infrastructure

### Overview
Add capability to render request body JSON examples and basic field definitions for POST/PUT/PATCH endpoints.

### Changes Required

#### 1. Schema Field Extraction
**File**: `convert.go`
**Changes**: Add schema field extraction types and functions

```go
// Schema field information for documentation
type schemaField struct {
	name            string
	typeStr         string
	required        bool
	description     string
	enum            []interface{}
	isArray         bool
	isObject        bool
	nestedSchemaRef string
}

// schemaDefinition represents a complete schema with all fields
type schemaDefinition struct {
	name        string
	description string
	fields      []schemaField
}
```

**Function Responsibilities**:
- Define data structures for schema field information
- Support type identification (string, integer, boolean, object, array)
- Track required status, enums, descriptions
- Mark nested object references for separate sections

#### 2. Request Body JSON Generation
**File**: `convert.go`
**Changes**: Add request body example extraction

```go
// extractRequestExample extracts or generates JSON example for request body
func extractRequestExample(op *v3.Operation, examples map[string]json.RawMessage) (string, error)
```

**Function Responsibilities**:
- Check if operation has requestBody: `if op.RequestBody == nil || op.RequestBody.Content == nil`
- Iterate through content types to find `application/json`: use `.First()` and `.Next()` pattern
- Follow same three-tier priority as responses: explicit > named > schema-generated
- Use pre-generated component examples from map
- Return formatted JSON with 3-space indentation
- Pattern reference: Follow `extractResponseExample()` at `convert.go:519-552`

**Access Pattern** (similar to responses):
```go
// 1. Check request body exists
if op.RequestBody == nil || op.RequestBody.Content == nil {
    return "", nil
}

// 2. Find application/json media type
for pair := op.RequestBody.Content.First(); pair != nil; pair = pair.Next() {
    mediaType := pair.Key()
    if mediaType != "application/json" {
        continue
    }

    mt := pair.Value()
    if mt == nil {
        continue
    }

    // 3. Priority 1: Explicit example from media type
    if explicit := getExampleFromMediaType(mt); explicit != "" {
        return explicit, nil
    }

    // 4. Priority 2-3: Generate from schema $ref
    if mt.Schema != nil {
        generated, err := getExampleFromSchema(mt.Schema, examples)
        if err != nil {
            return "", err
        }
        if generated != "" {
            return generated, nil
        }
    }
}

return "", nil
```

#### 3. Basic Field Definitions Rendering
**File**: `convert.go`
**Changes**: Add field definitions markdown generation

```go
// renderFieldDefinitions renders field definitions section for a schema
func renderFieldDefinitions(builder *strings.Builder, schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage, prefix string, visited map[string]int, maxDepth int) error
```

**Function Responsibilities**:
- Extract fields from schema using libopenapi Schema access
- Render top-level fields with bold headers
- Format: `**fieldName** (type, required)`
- Include description and enums if present
- Handle nested objects (defer to Phase 2)
- Business logic: Skip if schema is nil or not a reference

#### 4. Request Section Rendering
**File**: `convert.go`
**Changes**: Add request body rendering function

```go
// renderRequestBody renders request section with JSON and field definitions
func renderRequestBody(builder *strings.Builder, op *v3.Operation, examples map[string]json.RawMessage) error
```

**Function Responsibilities**:
- Check operation method (POST/PUT/PATCH only, not GET)
- Write "### Request" header
- Call `extractRequestExample()` for JSON
- Write JSON in code fence
- Call `renderFieldDefinitions()` for schema documentation
- Pattern reference: Similar to `renderResponses()` at `convert.go:344-380`

#### 5. Integration into Markdown Generation
**File**: `convert.go`
**Changes**: Call `renderRequestBody()` in endpoint rendering

Update `generateMarkdown()` at line 242 (multi-tag mode) and line 266 (single-tag mode) to call:
```go
if err := renderRequestBody(&builder, e.operation, examples); err != nil {
    return "", err
}
```

**Context for implementation**:
- Insert after `renderParameters()` call
- Before `renderResponses()` call
- Both tag modes need same update

### Testing Requirements

```go
// Test request body rendering for POST endpoint
func TestConvertRequestBodyPOST(t *testing.T)

// Test request body not shown for GET endpoint
func TestConvertRequestBodyGETSkipped(t *testing.T)

// Test request body with nested object schema
func TestConvertRequestBodyNestedSchema(t *testing.T)

// Test request body with array field
func TestConvertRequestBodyArrayField(t *testing.T)

// Test request body with enum field
func TestConvertRequestBodyEnumField(t *testing.T)
```

**Test Objectives**:
- Verify "### Request" section appears for POST/PUT/PATCH
- Verify request section NOT shown for GET
- Verify JSON example rendered correctly
- Verify basic field definitions rendered
- Verify required fields marked correctly
- Verify enum values displayed inline

**Context for implementation**:
- Follow pattern from `convert_test.go:485-524` (TestConvertResponseExamples)
- Use table-driven tests for different HTTP methods
- Assert markdown contains expected sections and field formats

### Validation
- [ ] Run: `go test -v -run TestConvertRequestBody`
- [ ] Verify: All request body tests pass
- [ ] Verify: Existing tests still pass (no regressions)

## Phase 2: Nested Schema Definitions

### Overview
Extend field definitions to handle nested objects with separate definition sections, following hierarchical pattern from updated example.

### Changes Required

#### 1. Schema Traversal with Recursion Tracking
**File**: `convert.go`
**Changes**: Enhance schema field extraction to handle nesting

```go
// extractSchemaFields recursively extracts field information from schema
func extractSchemaFields(schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage, visited map[string]int, maxDepth int) ([]schemaField, []schemaDefinition, error)
```

**Function Responsibilities**:
- Traverse schema properties using `schema.Properties` iteration
- For each property: extract name, type, description, required status, enums
- Identify nested objects: check if `propSchema.Type[0] == "object"`
- For nested objects: extract schema reference name, recursively process
- Track visited schemas with depth counter: `visited[schemaName]++`
- Cap recursion: if `visited[schemaName] > 1`, mark as "(recursive)" and stop
- Cap depth: if `maxDepth` reached, stop traversal
- Return both flat field list and nested schema definitions
- Pattern reference: Use schema access pattern from `renderParamTable()` at `convert.go:332-337`

#### 2. Nested Definition Section Rendering
**File**: `convert.go`
**Changes**: Render separate sections for each nested schema

```go
// renderSchemaDefinition renders a single schema definition section
func renderSchemaDefinition(builder *strings.Builder, def schemaDefinition) error
```

**Function Responsibilities**:
- Write bold header: `**schemaName**`
- For each field: write bullet point with type and description
- Format: `- 'fieldName' (type, required): Description. Enums: 'VAL1', 'VAL2'`
- Handle arrays: format as "(array of objects)" or "(string array)"
- Handle nested objects: mark as "(object, required): Description" with reference
- Business logic: Separate section means easier to reference from multiple places

#### 3. Update Field Definitions Function
**File**: `convert.go`
**Changes**: Update `renderFieldDefinitions()` to use traversal

Enhance function from Phase 1:
- Call `extractSchemaFields()` to get field tree
- Render top-level fields first
- Then render each nested schema definition in separate section
- Pass `visited` map and `maxDepth=10` to traversal
- Pattern reference: Follow hierarchical structure from `examples/updated-example.md:44-59`

**Context for implementation**:
- Top-level fields: lines 46-48 (principal), 50-53 (resource), 59 (actions)
- Nested definitions: lines 55-57 (nested schema)
- Each nested schema gets own bold header section

### Testing Requirements

```go
// Test nested object creates separate definition section
func TestConvertFieldDefinitionsNested(t *testing.T)

// Test deeply nested objects (3+ levels)
func TestConvertFieldDefinitionsDeeplyNested(t *testing.T)

// Test recursive schema capped at depth 1
func TestConvertFieldDefinitionsRecursive(t *testing.T)

// Test array of objects with nested schema
func TestConvertFieldDefinitionsArrayOfObjects(t *testing.T)

// Test maximum depth limit (10 levels)
func TestConvertFieldDefinitionsMaxDepth(t *testing.T)
```

**Test Objectives**:
- Verify nested objects get separate bold header sections
- Verify parent field shows "(object, required): Description"
- Verify child section shows all nested fields
- Verify recursive schemas marked and stopped
- Verify depth limit enforced at 10 levels
- Verify arrays of objects show item fields inline

**Context for implementation**:
- Create OpenAPI specs with `$ref` to component schemas
- Test schema like `examples/updated-example.md` nested example (lines 32-35, 55-57)
- Assert separate sections appear in correct order
- Assert recursive references don't cause infinite loops

### Validation
- [ ] Run: `go test -v -run TestConvertFieldDefinitions`
- [ ] Verify: All nested schema tests pass
- [ ] Verify: Phase 1 tests still pass

## Phase 3: Response Field Definitions

### Overview
Add field definitions to 2xx responses and update response section structure with "Responses" header.

### Changes Required

#### 1. Update Response Rendering Structure
**File**: `convert.go`
**Changes**: Modify `renderResponses()` function

Update function at `convert.go:344-380`:
- Add "### Responses" section header at start
- Change response code headers from "##### {code} Response" to "#### {code} Response"
- After JSON example: call `renderFieldDefinitions()` for 2xx responses only
- Skip field definitions for 4xx and 5xx responses

**Function Responsibilities**:
- Write "### Responses" as H3 header before processing response codes
- For each response code:
  - Write "#### {code} Response" (H4, not H5)
  - Write description
  - Write JSON example
  - If code starts with "2": call `renderFieldDefinitions()`
- Business logic: Only success responses need detailed field docs

**Context for implementation**:
- Current heading at line 358: `builder.WriteString("##### ")`
- Change to: `builder.WriteString("#### ")`
- Add field definitions call after line 376 (after JSON block)
- Check code: `if strings.HasPrefix(code, "2") { ... }`

#### 2. Shared Response Schema Detection
**File**: `convert.go`
**Changes**: Track when same schema used in multiple responses

```go
// identifySharedResponseSchemas finds schemas used in multiple responses within endpoint
func identifySharedResponseSchemas(op *v3.Operation) map[string][]string
```

**Function Responsibilities**:
- Iterate through all response codes
- Extract schema reference from each response
- Build map of schemaName -> []responseCodes
- Return schemas that appear in 2+ response codes
- Pattern reference: Similar to `groupByTags()` at `convert.go:142-156`

#### 3. Consolidated Field Definitions for Shared Responses
**File**: `convert.go`
**Changes**: Render field definitions once for multiple 2xx responses

Update `renderResponses()`:
- Call `identifySharedResponseSchemas()` first
- Group responses with same schema
- Render field definitions once with note: "#### Field Definitions (applies to 200, 201 responses)"
- Skip individual field definitions for grouped responses

**Function Responsibilities**:
- Detect when 200 and 201 both use same schema
- Render JSON for both
- Render field definitions once with combined note
- Business logic: Reduces duplication when error schemas reused

### Testing Requirements

```go
// Test response field definitions for 200 response
func TestConvertResponseFieldDefinitions2xx(t *testing.T)

// Test no field definitions for 400 response
func TestConvertResponseFieldDefinitions4xx(t *testing.T)

// Test "Responses" section header
func TestConvertResponsesSectionHeader(t *testing.T)

// Test shared schema in multiple 2xx responses
func TestConvertResponseSharedSchema(t *testing.T)

// Test response heading level is H4
func TestConvertResponseHeadingLevel(t *testing.T)
```

**Test Objectives**:
- Verify "### Responses" header appears
- Verify "#### 200 Response" uses H4 heading
- Verify field definitions appear after 2xx response JSON
- Verify field definitions NOT shown for 4xx/5xx
- Verify shared schemas documented once with combined note
- Verify nested response schemas create separate sections

**Context for implementation**:
- Follow pattern from `convert_test.go:485-524` (TestConvertResponseExamples)
- Assert markdown contains "### Responses"
- Assert markdown contains "#### 200 Response" (not "#####")
- Assert field definitions appear for success codes only

### Validation
- [ ] Run: `go test -v -run TestConvertResponse`
- [ ] Verify: All response field definition tests pass
- [ ] Verify: Existing response tests updated and passing

## Phase 4: Parameter Field Definitions Format

### Overview
Convert path and query parameters from table format to field definitions format, while keeping headers as tables.

### Changes Required

#### 1. Path Parameters Field Definitions
**File**: `convert.go`
**Changes**: Add new path parameter rendering function

```go
// renderPathParametersFieldDef renders path parameters in field definitions format
func renderPathParametersFieldDef(builder *strings.Builder, params []v3.Parameter) error
```

**Function Responsibilities**:
- Write "#### Path Parameters" header
- For each parameter: write bold header with type: `**paramName** (type, required)`
- Write description as bullet point
- Include enums if present in schema: `Enums: 'VAL1', 'VAL2'`
- Pattern reference: Convert table format from `renderParamTable()` at `convert.go:305-342`

**Context for implementation**:
- Parameter schema access: `param.Schema.Schema()` at line 332
- Type extraction: `schema.Type[0]` at line 335
- Required check: `param.Required != nil && *param.Required` at line 325
- Enum extraction: `schema.Enum` (new - not in current code)

**Enum Access Pattern**:
```go
if param.Schema != nil && param.Schema.Schema() != nil {
    schema := param.Schema.Schema()

    // Get type
    typeStr := ""
    if len(schema.Type) > 0 {
        typeStr = schema.Type[0]
    }

    // Get required status
    required := param.Required != nil && *param.Required

    // Format: **paramName** (type, required)
    builder.WriteString("**")
    builder.WriteString(param.Name)
    builder.WriteString("** (")
    builder.WriteString(typeStr)
    if required {
        builder.WriteString(", required")
    }
    builder.WriteString(")\n")

    // Description as bullet
    if param.Description != "" {
        builder.WriteString("- ")
        builder.WriteString(param.Description)

        // Add enums if present
        if schema.Enum != nil && len(schema.Enum) > 0 {
            builder.WriteString(". Enums: ")
            for i, enumVal := range schema.Enum {
                if i > 0 {
                    builder.WriteString(", ")
                }
                builder.WriteString("`")
                builder.WriteString(fmt.Sprintf("%v", enumVal))
                builder.WriteString("`")
            }
        }
        builder.WriteString("\n")
    }
    builder.WriteString("\n")
}
```

#### 2. Query Parameters Field Definitions
**File**: `convert.go`
**Changes**: Add new query parameter rendering function

```go
// renderQueryParametersFieldDef renders query parameters in field definitions format
func renderQueryParametersFieldDef(builder *strings.Builder, params []v3.Parameter) error
```

**Function Responsibilities**:
- Write "#### Query Parameters" header
- Same format as path parameters: `**paramName** (type)`
- Include enums from schema
- Business logic: Query params often have enums for filtering, new format allows documenting them

#### 3. Keep Headers as Table Format
**File**: `convert.go`
**Changes**: Extract header rendering to separate function

```go
// renderHeaders renders header parameters in table format
func renderHeaders(builder *strings.Builder, params []v3.Parameter) error
```

**Function Responsibilities**:
- Write "#### Headers" header (not "#### Header Parameters")
- Render as table (keep existing format)
- Pattern reference: Reuse logic from `renderParamTable()` at `convert.go:305-342`

#### 4. Update Parameter Rendering
**File**: `convert.go`
**Changes**: Replace `renderParameters()` calls

Update `renderParameters()` at `convert.go:277-303`:
- Replace table rendering with new functions
- Call `renderPathParametersFieldDef()` for path params
- Call `renderQueryParametersFieldDef()` for query params
- Call `renderHeaders()` for header params

**Context for implementation**:
- Line 300: replace `renderParamTable(builder, "Path", pathParams)`
- Line 301: replace `renderParamTable(builder, "Query", queryParams)`
- Line 302: replace `renderParamTable(builder, "Header", headerParams)`

### Testing Requirements

```go
// Test path parameters use field definitions format
func TestConvertPathParametersFieldDef(t *testing.T)

// Test query parameters use field definitions format
func TestConvertQueryParametersFieldDef(t *testing.T)

// Test headers remain in table format
func TestConvertHeadersTableFormat(t *testing.T)

// Test parameter with enum values
func TestConvertParameterWithEnums(t *testing.T)

// Test parameter required vs optional
func TestConvertParameterRequired(t *testing.T)
```

**Test Objectives**:
- Verify path params show `**paramName** (type, required)` format
- Verify query params show field definitions format
- Verify headers still use table format with "#### Headers" title
- Verify enums appear inline for parameters
- Verify required status shown correctly
- Verify parameter descriptions included

**Context for implementation**:
- Update existing tests: `convert_test.go:364-399` (TestConvertPathParameters)
- Update existing tests: `convert_test.go:401-441` (TestConvertQueryParameters)
- Update existing tests: `convert_test.go:443-483` (TestConvertHeaderParameters)
- Change assertions from table format to field definitions format

### Validation
- [ ] Run: `go test -v -run TestConvertPath`
- [ ] Run: `go test -v -run TestConvertQuery`
- [ ] Run: `go test -v -run TestConvertHeader`
- [ ] Verify: All parameter tests pass with new format

## Phase 5: Shared Schema Definitions

### Overview
Add document-level "Shared Schema Definitions" section for schemas used across multiple endpoints to reduce duplication.

### Changes Required

#### 1. Schema Usage Tracking
**File**: `convert.go`
**Changes**: Add schema usage analysis across all endpoints

```go
// schemaUsage tracks where schemas are used
type schemaUsage struct {
	schemaName string
	endpoints  []string  // e.g., ["POST /users", "PUT /users/{id}"]
	contexts   []string  // e.g., ["request", "response:200"]
}

// identifySharedSchemas finds schemas used in multiple endpoints
func identifySharedSchemas(endpoints []endpoint, examples map[string]json.RawMessage) map[string]schemaUsage
```

**Function Responsibilities**:
- Iterate through all endpoints
- For each endpoint: extract schema references from request body and responses
- Track schema name and where it's used (endpoint + context)
- Return schemas that appear across multiple endpoints (not same-endpoint request/response reuse)
- Pattern reference: Similar to `collectDebugInfo()` at `convert.go:382-422`

**Shared Schema Criteria**:
A schema is "shared" if it appears in **2+ different endpoints**:
- Example: `Error` schema in `POST /users` 400 response AND `GET /users/{id}` 404 response → SHARED
- Example: `User` schema in `POST /users` request AND `GET /users/{id}` response → SHARED
- Counter-example: `CreateUserRequest` only in `POST /users` request → NOT shared (document inline)
- Counter-example: `UserResponse` in `POST /users` 201 AND `PUT /users/{id}` 200 → SHARED (different endpoints)

**NOT considered shared** (document inline):
- Schema used only within one endpoint (even if in both request and response)
- Schema used only in responses of one endpoint (e.g., 200, 201, 400 of same endpoint)

**Algorithm**:
```go
schemaToEndpoints := make(map[string]map[string]bool) // schema → set of endpoints

for _, endpoint := range endpoints {
    endpointKey := endpoint.method + " " + endpoint.path

    // Extract request body schema if exists
    if requestSchema := extractSchemaRef(endpoint.operation.RequestBody); requestSchema != "" {
        if schemaToEndpoints[requestSchema] == nil {
            schemaToEndpoints[requestSchema] = make(map[string]bool)
        }
        schemaToEndpoints[requestSchema][endpointKey] = true
    }

    // Extract response schemas
    for each response in endpoint.operation.Responses {
        if responseSchema := extractSchemaRef(response); responseSchema != "" {
            if schemaToEndpoints[responseSchema] == nil {
                schemaToEndpoints[responseSchema] = make(map[string]bool)
            }
            schemaToEndpoints[responseSchema][endpointKey] = true
        }
    }
}

// Return only schemas appearing in 2+ endpoints
sharedSchemas := make(map[string]schemaUsage)
for schemaName, endpoints := range schemaToEndpoints {
    if len(endpoints) >= 2 {
        sharedSchemas[schemaName] = schemaUsage{
            schemaName: schemaName,
            endpoints: keysToSlice(endpoints),
        }
    }
}
```

**Context for implementation**:
- Access request schema: `op.RequestBody.Content` → media type → schema
- Access response schemas: `op.Responses.Codes` → response → content → schema
- Extract schema name using existing `extractSchemaName()` at `convert.go:441-454`

#### 2. Shared Definitions Section Rendering
**File**: `convert.go`
**Changes**: Add shared definitions markdown generation

```go
// renderSharedDefinitions renders shared schema definitions section
func renderSharedDefinitions(builder *strings.Builder, sharedSchemas map[string]schemaUsage, examples map[string]json.RawMessage) error
```

**Function Responsibilities**:
- Write "## Shared Schema Definitions" header
- For each shared schema: write "### {SchemaName}" subheader
- Call `renderFieldDefinitions()` to document schema fields
- Include note about where schema is used: "Used in: POST /users (request), GET /users (response)"
- Business logic: Document once, reference multiple times

#### 3. Schema Reference Rendering
**File**: `convert.go`
**Changes**: Update field definitions to reference shared schemas

Update `renderFieldDefinitions()`:
- Check if schema is in shared definitions map
- If shared: write "See [SchemaName](#shared-schema-definitions)" instead of full documentation
- If not shared: render full field definitions inline
- Pattern reference: Markdown anchor link format like TOC at `convert.go:186-194`

**Context for implementation**:
- Before rendering fields, check: `if _, isShared := sharedSchemas[schemaName]; isShared { ... }`
- Write reference: `builder.WriteString("See [" + schemaName + "](#" + anchor + ")")`
- Generate anchor: use `makeAnchor()` pattern from `convert.go:158-166`

#### 4. Integration into Document Generation
**File**: `convert.go`
**Changes**: Update `generateMarkdown()` to include shared definitions

Add shared definitions section:
- Call `identifySharedSchemas()` after `groupByTags()` at line 80
- Render shared definitions section after TOC, before tag sections
- Insert at line 197 (after TOC, before tag loop)

**Context for implementation**:
- Line 80: `sharedSchemas := identifySharedSchemas(endpoints, examples)`
- Line 197: `if err := renderSharedDefinitions(&builder, sharedSchemas, examples); err != nil { ... }`
- Pass `sharedSchemas` to all field definition rendering functions

### Testing Requirements

```go
// Test schema used in multiple endpoints appears in shared section
func TestConvertSharedSchemaAcrossEndpoints(t *testing.T)

// Test schema used in request and response of same endpoint
func TestConvertSharedSchemaWithinEndpoint(t *testing.T)

// Test endpoint field definitions reference shared schema
func TestConvertFieldDefinitionsReferenceShared(t *testing.T)

// Test non-shared schema still documented inline
func TestConvertNonSharedSchemaInline(t *testing.T)

// Test shared definitions section placement
func TestConvertSharedDefinitionsSectionPlacement(t *testing.T)
```

**Test Objectives**:
- Verify "## Shared Schema Definitions" appears after TOC
- Verify schema used in 2+ endpoints documented in shared section
- Verify endpoint field definitions contain reference link
- Verify reference link format: `See [SchemaName](#anchor)`
- Verify non-shared schemas still get full inline documentation
- Verify shared section includes usage note

**Context for implementation**:
- Create OpenAPI spec with `Error` schema used in multiple endpoints
- Create spec with `User` schema used in POST request and GET response
- Assert "## Shared Schema Definitions" appears
- Assert "### Error" and "### User" subsections appear
- Assert endpoint contains "See [Error](#error)" reference

### Validation
- [ ] Run: `go test -v -run TestConvertShared`
- [ ] Verify: Shared schema tests pass
- [ ] Verify: Example markdown shows shared section correctly

## Phase 6: Integration Testing and Golden File Updates

### Overview
Update all existing tests to match new format, regenerate golden files, and ensure complete integration.

### Changes Required

#### 1. Update All Existing Tests
**File**: `convert_test.go`
**Changes**: Update assertions in all existing tests

Tests to update:
- `TestConvertTableOfContents` (line 189): Verify TOC unchanged
- `TestConvertSingleEndpoint` (line 226): Update for new endpoint format
- `TestConvertMultipleEndpoints` (line 258): Verify tag sections with new format
- `TestConvertUntaggedOperations` (line 299): Verify Default APIs section
- `TestConvertMultipleTagsPerOperation` (line 329): Verify duplicate endpoints in tags
- `TestConvertCompleteEndpoint` (line 526): Update all parameter and response assertions
- `TestConvertIntegrationPetStore` (line 712): Full integration test update
- `TestConvertIntegrationComplexAPI` (line 777): Complex scenario update
- `TestConvertCompleteExample` (line 926): Compare against updated golden file

**Context for implementation**:
- Change assertions from table format to field definitions
- Update response heading assertions from "#####" to "####"
- Add assertions for "### Responses" header
- Add assertions for "### Request" section
- Update markdown content checks for new format

#### 2. Regenerate Example Files
**File**: `examples/example.md`
**Changes**: Regenerate using updated conversion

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

**Context for implementation**:
- Run after all phases complete
- Verify output matches expected new format
- Check for field definitions, request sections, shared definitions

#### 3. Update Golden Files
**File**: `testdata/golden/petstore-example.md`
**Changes**: Copy regenerated example as new golden file

```bash
cp examples/example.md testdata/golden/petstore-example.md
```

**Context for implementation**:
- Golden file used by `TestCompleteExampleWithGoldenFile` at line 1151
- Ensures output format stability across changes
- Pattern reference: Test reads golden file and compares byte-for-byte at line 1164

#### 4. Debug Info Updates
**File**: `convert.go`
**Changes**: Update DebugInfo to track new features

Add to `DebugInfo` struct:
```go
type DebugInfo struct {
    ParsedPaths      int
    ExtractedOps     int
    TagsFound        []string
    UntaggedOps      int
    ParameterCounts  map[string]int
    ResponseCounts   map[string]int
    RequestBodyCount int              // NEW: count of endpoints with request bodies
    SharedSchemaCount int             // NEW: count of shared schemas
    NestedSchemaDepth map[string]int  // NEW: max depth for each schema
}
```

Update `collectDebugInfo()` function:
- Count endpoints with request bodies
- Count shared schemas identified
- Track maximum nesting depth encountered

**Context for implementation**:
- Add to `collectDebugInfo()` at line 382-422
- Iterate through endpoints checking for `operation.RequestBody`
- Call `identifySharedSchemas()` and count results
- Useful for testing and observability

### Testing Requirements

```go
// Update existing tests - no new test signatures needed
// All existing tests will be updated to expect new format
```

**Test Objectives**:
- All 19 existing tests pass with new format
- Golden file test passes with updated golden file
- Debug info includes new metrics
- No test regressions

**Context for implementation**:
- Systematically update each test in `convert_test.go`
- Run `go test -v` after each test update
- Fix any assertion failures
- Ensure new format consistently applied

### Validation
- [ ] Run: `go test ./...`
- [ ] Verify: All tests pass (0 failures)
- [ ] Run: `diff examples/example.md testdata/golden/petstore-example.md`
- [ ] Verify: Files are identical
- [ ] Run: `go build ./...`
- [ ] Verify: No compilation errors

## Phase 7: Documentation and Examples

### Overview
Update project documentation and examples to reflect new format and capabilities.

### Changes Required

#### 1. Update CLAUDE.md
**File**: `CLAUDE.md`
**Changes**: Update project overview and architecture description

Update sections:
- "Response Example Generation" → add request body generation
- "Architecture" → add field definitions rendering
- "Internal Flow" → add shared schema identification
- Add new section: "Field Definitions Format"
- Update "Regenerating Example Output" → reflect new format

**Context for implementation**:
- Document the hierarchical field definitions pattern
- Explain shared schema definitions section
- Note parameter format changes
- Reference `examples/updated-example.md` as canonical format

#### 2. Update README.md
**File**: `README.md`
**Changes**: Update usage examples and feature list

Add to features:
- Request body documentation with field definitions
- Nested schema documentation with separate sections
- Shared schema definitions across endpoints
- Parameter documentation with enum support
- Rich response field documentation

Update example output:
- Show snippet of new field definitions format
- Demonstrate nested schema documentation
- Show shared definitions section

#### 3. Create Migration Guide
**File**: `MIGRATION.md` (new)
**Changes**: Document breaking changes and migration path

```markdown
# Migration Guide

## Breaking Changes

### Format Changes
- Parameters: Path and Query now use field definitions format
- Responses: Heading level changed from H5 to H4
- Request bodies: Now documented with JSON examples and field definitions

### What Stays the Same
- Document structure: Title, TOC, tags
- Headers: Still use table format
- Tag grouping behavior
- Response example priority system

## Updating Consumers
...
```

### Testing Requirements

No new test signatures - documentation only changes

**Test Objectives**:
- Documentation accurately reflects implementation
- Examples are up-to-date
- Migration guide is clear

### Validation
- [ ] Verify: `CLAUDE.md` reflects new architecture
- [ ] Verify: `README.md` shows new format examples
- [ ] Verify: `MIGRATION.md` covers all breaking changes
- [ ] Run: `go test ./...` (ensure docs didn't break anything)

---

## Complete Implementation Checklist

### Phase 1: Request Body Rendering
- [ ] Add schema field data structures
- [ ] Implement `extractRequestExample()`
- [ ] Implement basic `renderFieldDefinitions()`
- [ ] Implement `renderRequestBody()`
- [ ] Integrate into `generateMarkdown()`
- [ ] Write and pass 5 request body tests

### Phase 2: Nested Schema Definitions
- [ ] Implement `extractSchemaFields()` with recursion tracking
- [ ] Implement `renderSchemaDefinition()`
- [ ] Update `renderFieldDefinitions()` for nesting
- [ ] Write and pass 5 nested schema tests

### Phase 3: Response Field Definitions
- [ ] Add "Responses" section header
- [ ] Update response heading levels to H4
- [ ] Add field definitions to 2xx responses
- [ ] Implement `identifySharedResponseSchemas()`
- [ ] Consolidate shared response schemas
- [ ] Write and pass 5 response field definition tests

### Phase 4: Parameter Field Definitions
- [ ] Implement `renderPathParametersFieldDef()`
- [ ] Implement `renderQueryParametersFieldDef()`
- [ ] Implement `renderHeaders()`
- [ ] Update `renderParameters()` to use new functions
- [ ] Write and pass 5 parameter format tests
- [ ] Update existing parameter tests

### Phase 5: Shared Schema Definitions
- [ ] Implement `identifySharedSchemas()`
- [ ] Implement `renderSharedDefinitions()`
- [ ] Update `renderFieldDefinitions()` for references
- [ ] Integrate into `generateMarkdown()`
- [ ] Write and pass 5 shared schema tests

### Phase 6: Integration Testing
- [ ] Update all 19 existing tests for new format
- [ ] Regenerate `examples/example.md`
- [ ] Update `testdata/golden/petstore-example.md`
- [ ] Update `DebugInfo` struct
- [ ] All tests pass

### Phase 7: Documentation
- [ ] Update `CLAUDE.md`
- [ ] Update `README.md`
- [ ] Create `MIGRATION.md`
- [ ] Verify documentation accuracy

---

## Success Criteria

### Functional Requirements
✅ Request bodies rendered for POST/PUT/PATCH with JSON and field definitions
✅ Field definitions show nested objects in hierarchical format
✅ Path and Query parameters use field definitions format
✅ Headers remain in table format
✅ Responses include "Responses" header and field definitions for 2xx
✅ Shared schemas documented once in dedicated section
✅ Recursion capped at depth 1, nesting capped at depth 10

### Quality Requirements
✅ All tests pass (existing + new)
✅ Golden file matches generated output
✅ No compilation errors or warnings
✅ Code follows existing patterns and conventions
✅ TDD approach: tests written before implementation

### Documentation Requirements
✅ CLAUDE.md reflects new architecture
✅ README.md shows new format
✅ Migration guide covers breaking changes
✅ Examples regenerated and accurate
