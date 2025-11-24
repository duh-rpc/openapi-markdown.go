# Migration Guide

This guide documents breaking changes introduced by the field definitions format update and provides guidance for migrating to the new output format.

## Breaking Changes

### Format Changes

#### 1. Request Bodies Now Documented

**Previous behavior**: Request bodies were completely ignored and not rendered in the markdown output.

**New behavior**: Request bodies for POST/PUT/PATCH/DELETE operations now include:
- JSON example generated from schema
- Field definitions section documenting each field with type, required status, and description
- Nested objects documented in separate sections

**Impact**: Generated markdown will include new `### Request` sections for endpoints with request bodies.

#### 2. Parameters Use Field Definitions Format

**Previous behavior**: Path, query, and header parameters all used table format:

```markdown
#### Path Parameters

Name | Description | Required | Type
-----|-------------|----------|-----
id | User ID | true | string
```

**New behavior**:
- Path and query parameters use field definitions format
- Headers remain in table format

```markdown
#### Path Parameters

**id** (string, required)
- User ID

#### Headers

Name | Description | Required | Type
-----|-------------|----------|-----
X-Token | Auth token | true | string
```

**Impact**: Parameter sections will have different markdown structure. Automated parsers expecting table format will need updates.

#### 3. Response Heading Level Changed

**Previous behavior**: Response headings used H5 (`##### 200 Response`)

**New behavior**: Response headings use H4 (`#### 200 Response`)

**Impact**:
- Markdown heading hierarchy changed
- Automated TOC generators may produce different structure
- CSS selectors targeting response headings need adjustment

#### 4. Responses Section Header Added

**Previous behavior**: Responses rendered directly without section header

**New behavior**: All responses grouped under `### Responses` header

**Impact**:
- Adds consistent section structure
- Improves visual organization
- May affect automated parsing

#### 5. Field Definitions for Success Responses

**Previous behavior**: Responses only showed JSON examples

**New behavior**:
- 2xx success responses include field definitions documenting schema
- 4xx/5xx error responses show JSON only (no field definitions)

**Impact**:
- Success responses significantly more detailed
- Better documentation of response structure
- Increased markdown size for endpoints with complex schemas

#### 6. Enum Support in Parameters and Fields

**Previous behavior**: Enum values not shown in parameters or field definitions

**New behavior**: Enums shown inline with field documentation:

```markdown
**status** (string, required)
- The order status. Enums: `pending`, `confirmed`, `shipped`, `delivered`
```

**Impact**:
- Richer parameter documentation
- Users can see allowed values without checking schema
- Markdown size increases for fields with many enum values

#### 7. Shared Schema Definitions Section

**Previous behavior**: Schema documentation duplicated across endpoints

**New behavior**:
- Schemas used across multiple endpoints documented once in "Shared Schema Definitions" section
- Endpoints reference shared schemas instead of duplicating documentation
- Section appears after TOC, before endpoint documentation

**Impact**:
- Reduced duplication for common schemas (Error, User, etc.)
- Document structure includes new top-level section
- References use markdown anchor links

## What Stays the Same

The following behaviors are unchanged:

- **Document structure**: Title, description, table of contents
- **Tag grouping**: Endpoints organized by tags, "Default APIs" for untagged
- **Tag behavior**: Multiple tags per operation, single-tag flattening
- **Response example priority**: Explicit > named > schema-generated
- **Schema validation**: Response schemas must use $ref (no inline schemas)
- **Example generation**: Uses openapi-schema.go library for schema-based examples
- **OpenAPI support**: OpenAPI 3.x only (2.0 rejected)

## Updating Consumers

### Markdown Parsers

If you parse the generated markdown programmatically:

1. **Update response heading selectors**: Change from H5 (`#####`) to H4 (`####`)
2. **Handle new sections**: Account for `### Request` and `### Responses` headers
3. **Update parameter parsing**: Path/query params use field definitions, not tables
4. **Handle shared schemas**: Parse "Shared Schema Definitions" section if needed

### Documentation Toolchains

If you process the markdown with other tools:

1. **TOC generators**: May produce different structure due to heading changes
2. **CSS styling**: Update selectors for response headings (H5 → H4)
3. **Static site generators**: Verify heading hierarchy renders correctly
4. **Search indexing**: May need reindexing due to structural changes

### Testing and Validation

If you have tests that validate markdown output:

1. **Update golden files**: Regenerate expected output with new format
2. **Update assertions**: Change heading level checks, table format checks
3. **Add new checks**: Validate field definitions sections appear
4. **Update snapshots**: If using snapshot testing, regenerate snapshots

## Migration Steps

### For Library Users

1. **Update library version**: `go get github.com/duh-rpc/openapi-markdown.go@latest`
2. **Regenerate documentation**: Run converter on your OpenAPI specs
3. **Review output**: Check that field definitions appear correctly
4. **Update downstream tools**: If you parse the markdown, update parsers
5. **Test thoroughly**: Verify documentation renders correctly in your environment

### For Contributors

1. **Read CLAUDE.md**: Updated to reflect new architecture
2. **Review examples/updated-example.md**: Shows canonical format
3. **Run tests**: `go test ./...` to verify implementation
4. **Check golden files**: Compare `examples/example.md` with expected output

## Rollback Plan

If you need to revert to the old format:

```bash
# Use the last version before field definitions format
go get github.com/duh-rpc/openapi-markdown.go@v0.6.0
```

Note: Version number is illustrative. Check releases for actual pre-migration version.

## Support

For issues or questions about migration:

1. Check examples/updated-example.md for format reference
2. Review CLAUDE.md for architecture details
3. Open GitHub issue with specific migration question
4. Include your OpenAPI spec snippet if relevant

## Timeline

- **Field definitions format**: Implemented in phases 1-6
- **Documentation update**: Phase 7 (this migration guide)
- **Recommended migration window**: Update at your convenience
- **Old format support**: None (complete format replacement)
