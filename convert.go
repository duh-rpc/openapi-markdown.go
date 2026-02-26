package conv

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	proto "github.com/duh-rpc/openapi-schema.go"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// ConvertResult contains markdown output and generation metadata
type ConvertResult struct {
	Markdown      []byte
	EndpointCount int
	TagCount      int
	Warnings      []string
	Debug         *DebugInfo
}

// DebugInfo provides visibility into internal conversion process for testing
type DebugInfo struct {
	ParsedPaths       int
	ExtractedOps      int
	TagsFound         []string
	UntaggedOps       int
	ParameterCounts   map[string]int
	ResponseCounts    map[string]int
	RequestBodyCount  int
	SharedSchemaCount int
	NestedSchemaDepth map[string]int
}

// ConvertOptions configures markdown generation
type ConvertOptions struct {
	EnableSharedSchemas bool
	Description         string
	Title               string
	Debug               bool
}

// Convert converts OpenAPI 3.x to markdown API documentation
func Convert(openapi []byte, opts ConvertOptions) (*ConvertResult, error) {
	if len(openapi) == 0 {
		return nil, fmt.Errorf("openapi input cannot be empty")
	}

	if opts.Title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	doc, err := libopenapi.NewDocument(openapi)
	if err != nil {
		return nil, fmt.Errorf("failed to parse openapi document: %w", err)
	}

	if doc.GetVersion() == "" {
		return nil, fmt.Errorf("failed to determine openapi version")
	}

	if !strings.HasPrefix(doc.GetVersion(), "3.") {
		return nil, fmt.Errorf("only openapi 3.x is supported, got version: %s", doc.GetVersion())
	}

	v3Model, err := doc.BuildV3Model()
	if err != nil {
		return nil, fmt.Errorf("failed to build openapi 3.x model: %w", err)
	}

	if v3Model == nil {
		return nil, fmt.Errorf("only openapi 3.x is supported")
	}

	examples, err := generateComponentExamples(openapi)
	if err != nil {
		return nil, fmt.Errorf("failed to generate component examples: %w", err)
	}

	endpoints := extractEndpoints(v3Model.Model)
	tagGroups := groupByTags(endpoints)
	sharedSchemas := identifySharedSchemas(endpoints)

	markdownSharedSchemas := map[string]schemaUsage{}
	if opts.EnableSharedSchemas {
		markdownSharedSchemas = sharedSchemas
	}

	markdown, warnings, err := generateMarkdown(opts, endpoints, tagGroups, examples, markdownSharedSchemas, v3Model.Model)
	if err != nil {
		return nil, err
	}

	result := &ConvertResult{
		Markdown:      []byte(markdown),
		EndpointCount: len(endpoints),
		TagCount:      len(tagGroups),
		Warnings:      warnings,
	}

	if opts.Debug {
		result.Debug = collectDebugInfo(v3Model.Model, endpoints, tagGroups, sharedSchemas)
	}

	return result, nil
}

type endpoint struct {
	method      string
	path        string
	summary     string
	description string
	tags        []string
	operation   *v3.Operation
}

// schemaField represents information about a single field in a schema
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
	name   string
	fields []schemaField
}

// schemaUsage tracks where schemas are used across endpoints
type schemaUsage struct {
	schemaName string
	endpoints  []string // e.g., ["POST /users", "PUT /users/{id}"]
}

func extractEndpoints(model v3.Document) []endpoint {
	var endpoints []endpoint

	if model.Paths == nil || model.Paths.PathItems == nil {
		return endpoints
	}

	for pathPair := model.Paths.PathItems.First(); pathPair != nil; pathPair = pathPair.Next() {
		path := pathPair.Key()
		pathItem := pathPair.Value()

		for opPair := pathItem.GetOperations().First(); opPair != nil; opPair = opPair.Next() {
			method := opPair.Key()
			op := opPair.Value()

			e := endpoint{
				method:      strings.ToUpper(method),
				path:        path,
				summary:     op.Summary,
				description: op.Description,
				operation:   op,
			}

			if len(op.Tags) > 0 {
				e.tags = op.Tags
			}

			endpoints = append(endpoints, e)
		}
	}

	return endpoints
}

func groupByTags(endpoints []endpoint) map[string][]endpoint {
	tagGroups := make(map[string][]endpoint)

	for _, e := range endpoints {
		if len(e.tags) == 0 {
			tagGroups["Default APIs"] = append(tagGroups["Default APIs"], e)
		} else {
			for _, tag := range e.tags {
				tagGroups[tag] = append(tagGroups[tag], e)
			}
		}
	}

	return tagGroups
}

// identifySharedSchemas finds schemas used in multiple endpoints
func identifySharedSchemas(endpoints []endpoint) map[string]schemaUsage {
	schemaToEndpoints := make(map[string]map[string]bool)

	for _, e := range endpoints {
		endpointKey := e.method + " " + e.path

		// Extract request body schema if exists
		if e.operation != nil && e.operation.RequestBody != nil && e.operation.RequestBody.Content != nil {
			for pair := e.operation.RequestBody.Content.First(); pair != nil; pair = pair.Next() {
				if pair.Key() != "application/json" {
					continue
				}

				mt := pair.Value()
				if mt == nil || mt.Schema == nil {
					continue
				}

				if mt.Schema.IsReference() {
					ref := mt.Schema.GetReference()
					schemaName, err := extractSchemaName(ref)
					if err == nil {
						if schemaToEndpoints[schemaName] == nil {
							schemaToEndpoints[schemaName] = make(map[string]bool)
						}
						schemaToEndpoints[schemaName][endpointKey] = true
					}
				}
			}
		}

		// Extract response schemas
		if e.operation != nil && e.operation.Responses != nil && e.operation.Responses.Codes != nil {
			for pair := e.operation.Responses.Codes.First(); pair != nil; pair = pair.Next() {
				resp := pair.Value()
				if resp == nil || resp.Content == nil {
					continue
				}

				for contentPair := resp.Content.First(); contentPair != nil; contentPair = contentPair.Next() {
					if contentPair.Key() != "application/json" {
						continue
					}

					mt := contentPair.Value()
					if mt == nil || mt.Schema == nil {
						continue
					}

					if mt.Schema.IsReference() {
						ref := mt.Schema.GetReference()
						schemaName, err := extractSchemaName(ref)
						if err == nil {
							if schemaToEndpoints[schemaName] == nil {
								schemaToEndpoints[schemaName] = make(map[string]bool)
							}
							schemaToEndpoints[schemaName][endpointKey] = true
						}
					}
				}
			}
		}
	}

	// Return only schemas appearing in 2+ endpoints
	sharedSchemas := make(map[string]schemaUsage)
	for schemaName, endpointsSet := range schemaToEndpoints {
		if len(endpointsSet) >= 2 {
			endpointsList := make([]string, 0, len(endpointsSet))
			for endpoint := range endpointsSet {
				endpointsList = append(endpointsList, endpoint)
			}
			sort.Strings(endpointsList)
			sharedSchemas[schemaName] = schemaUsage{
				schemaName: schemaName,
				endpoints:  endpointsList,
			}
		}
	}

	return sharedSchemas
}

func makeAnchor(method, path string) string {
	combined := method + " " + path
	combined = strings.ToLower(combined)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	combined = reg.ReplaceAllString(combined, "")

	return combined
}

// makeSchemaAnchor creates an anchor for a schema name
func makeSchemaAnchor(schemaName string) string {
	anchor := strings.ToLower(schemaName)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	anchor = reg.ReplaceAllString(anchor, "-")
	return strings.Trim(anchor, "-")
}

// renderSharedDefinitions renders shared schema definitions section
func renderSharedDefinitions(builder *strings.Builder, sharedSchemas map[string]schemaUsage, model v3.Document, examples map[string]json.RawMessage) error {
	if len(sharedSchemas) == 0 {
		return nil
	}

	builder.WriteString("## Shared Schema Definitions\n\n")

	// Sort schema names for consistent ordering
	schemaNames := make([]string, 0, len(sharedSchemas))
	for name := range sharedSchemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	for _, schemaName := range schemaNames {
		usage := sharedSchemas[schemaName]

		builder.WriteString("### ")
		builder.WriteString(schemaName)
		builder.WriteString("\n\n")

		// Add usage note
		if len(usage.endpoints) > 0 {
			builder.WriteString("Used in: ")
			for i, endpoint := range usage.endpoints {
				if i > 0 {
					builder.WriteString(", ")
				}
				builder.WriteString(endpoint)
			}
			builder.WriteString("\n\n")
		}

		// Get schema from model
		if model.Components != nil && model.Components.Schemas != nil {
			schemaPair := model.Components.Schemas.GetOrZero(schemaName)
			if schemaPair != nil {
				schema := schemaPair.Schema()
				if schema != nil {
					if err := renderSharedSchemaFields(builder, schema, schemaName); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// renderSharedFieldsList renders a list of schema fields in the shared definitions format
func renderSharedFieldsList(builder *strings.Builder, fields []schemaField, nestedDefs []schemaDefinition, schemaName string) error {
	for _, field := range fields {
		builder.WriteString("- `")
		builder.WriteString(field.name)
		builder.WriteString("`")

		if field.typeStr != "" {
			builder.WriteString(" *(")

			if field.isArray && !field.isObject {
				builder.WriteString(field.typeStr)
				builder.WriteString(" array")
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
				builder.WriteString(field.typeStr)
			}

			if field.required {
				builder.WriteString(", required")
			}
			builder.WriteString(")*")
		}

		if field.description != "" {
			builder.WriteString(" ")
			builder.WriteString(field.description)
		} else if !field.isObject {
			log.Printf("Warning: Field '%s' in schema '%s' is missing a description", field.name, schemaName)
		}

		if len(field.enum) > 0 {
			builder.WriteString(" Enums: ")
			for i, enumVal := range field.enum {
				if i > 0 {
					builder.WriteString(", ")
				}
				builder.WriteString("`")
				fmt.Fprintf(builder, "%v", enumVal)
				builder.WriteString("`")
			}
		}

		builder.WriteString("\n")
	}

	builder.WriteString("\n")

	for _, nestedDef := range nestedDefs {
		if err := renderSchemaDefinition(builder, nestedDef); err != nil {
			return err
		}
	}

	return nil
}

// renderSharedSchemaFields renders fields for a shared schema, handling oneOf, allOf, and plain properties
func renderSharedSchemaFields(builder *strings.Builder, schema *base.Schema, schemaName string) error {
	const maxDepth = 10

	// Handle oneOf schemas
	if len(schema.OneOf) > 0 {
		if schema.Discriminator != nil && schema.Discriminator.PropertyName != "" {
			builder.WriteString("Request body is one of the following variants, selected by the `")
			builder.WriteString(schema.Discriminator.PropertyName)
			builder.WriteString("` field:\n\n")
		}

		for _, variantProxy := range schema.OneOf {
			if variantProxy == nil {
				continue
			}
			if variantProxy.IsReference() {
				ref := variantProxy.GetReference()
				name, err := extractSchemaName(ref)
				if err == nil {
					builder.WriteString("**")
					builder.WriteString(name)
					builder.WriteString("**\n")
				}
			}

			variantSchema := variantProxy.Schema()
			if variantSchema == nil {
				continue
			}

			visited := make(map[string]int)
			visited[schemaName] = 1
			fields, nestedDefs, err := extractSchemaFieldsFromProperties(variantSchema, visited, maxDepth)
			if err != nil {
				return err
			}

			if err := renderSharedFieldsList(builder, fields, nestedDefs, schemaName); err != nil {
				return err
			}
		}

		return nil
	}

	// Handle allOf or plain properties
	mergedProps, _ := mergeAllOfProperties(schema)
	if mergedProps == nil || mergedProps.Len() == 0 {
		return nil
	}

	visited := make(map[string]int)
	visited[schemaName] = 1

	fields, nestedDefs, err := extractSchemaFieldsFromProperties(schema, visited, maxDepth)
	if err != nil {
		return err
	}

	return renderSharedFieldsList(builder, fields, nestedDefs, schemaName)
}

func generateMarkdown(opts ConvertOptions, endpoints []endpoint, tagGroups map[string][]endpoint, examples map[string]json.RawMessage, sharedSchemas map[string]schemaUsage, model v3.Document) (string, []string, error) {
	var builder strings.Builder
	var warnings []string

	builder.WriteString("# ")
	builder.WriteString(opts.Title)
	builder.WriteString("\n\n")

	if opts.Description != "" {
		builder.WriteString(opts.Description)
		builder.WriteString("\n\n")
	}

	if len(endpoints) > 0 {
		builder.WriteString("## Table of Contents\n\n")
		builder.WriteString("HTTP Request | Description\n")
		builder.WriteString("-------------|------------\n")

		for _, e := range endpoints {
			anchor := makeAnchor(e.method, e.path)
			builder.WriteString(e.method)
			builder.WriteString(" [")
			builder.WriteString(e.path)
			builder.WriteString("](#")
			builder.WriteString(anchor)
			builder.WriteString(") | ")
			builder.WriteString(e.summary)
			builder.WriteString("\n")
		}

		builder.WriteString("\n")

		tags := make([]string, 0, len(tagGroups))
		for tag := range tagGroups {
			tags = append(tags, tag)
		}
		sort.Strings(tags)

		defaultIndex := -1
		for i, tag := range tags {
			if tag == "Default APIs" {
				defaultIndex = i
				break
			}
		}
		if defaultIndex != -1 {
			tags = append(tags[:defaultIndex], tags[defaultIndex+1:]...)
			tags = append(tags, "Default APIs")
		}

		if len(tags) > 1 {
			for _, tag := range tags {
				endpoints := tagGroups[tag]

				builder.WriteString("## ")
				builder.WriteString(tag)
				builder.WriteString("\n\n")

				for _, e := range endpoints {
					builder.WriteString("### ")
					builder.WriteString(e.method)
					builder.WriteString(" ")
					builder.WriteString(e.path)
					builder.WriteString("\n\n")

					if e.description != "" {
						builder.WriteString(e.description)
						builder.WriteString("\n\n")
					} else if e.summary != "" {
						builder.WriteString(e.summary)
						builder.WriteString("\n\n")
					} else {
						log.Printf("Warning: No description or summary for %s %s", e.method, e.path)
					}

					renderParameters(&builder, e.operation)
					if err := renderRequestBody(&builder, e.operation, examples, sharedSchemas); err != nil {
						return "", nil, err
					}
					if err := renderResponses(&builder, e.operation, examples, sharedSchemas); err != nil {
						return "", nil, err
					}
				}
			}
		} else {
			for _, e := range endpoints {
				builder.WriteString("## ")
				builder.WriteString(e.method)
				builder.WriteString(" ")
				builder.WriteString(e.path)
				builder.WriteString("\n\n")

				if e.description != "" {
					builder.WriteString(e.description)
					builder.WriteString("\n\n")
				} else if e.summary != "" {
					builder.WriteString(e.summary)
					builder.WriteString("\n\n")
				} else {
					log.Printf("Warning: No description or summary for %s %s", e.method, e.path)
				}

				renderParameters(&builder, e.operation)
				if err := renderRequestBody(&builder, e.operation, examples, sharedSchemas); err != nil {
					return "", nil, err
				}
				if err := renderResponses(&builder, e.operation, examples, sharedSchemas); err != nil {
					return "", nil, err
				}
			}
		}

		// Render shared schema definitions section at the bottom
		if err := renderSharedDefinitions(&builder, sharedSchemas, model, examples); err != nil {
			return "", nil, err
		}
	}

	return builder.String(), warnings, nil
}

func renderParameters(builder *strings.Builder, op *v3.Operation) {
	if op == nil || op.Parameters == nil {
		return
	}

	pathParams := []v3.Parameter{}
	queryParams := []v3.Parameter{}
	headerParams := []v3.Parameter{}

	for _, param := range op.Parameters {
		if param == nil {
			continue
		}
		switch param.In {
		case "path":
			pathParams = append(pathParams, *param)
		case "query":
			queryParams = append(queryParams, *param)
		case "header":
			headerParams = append(headerParams, *param)
		}
	}

	renderPathParametersFieldDef(builder, pathParams)
	renderQueryParametersFieldDef(builder, queryParams)
	renderHeaders(builder, headerParams)
}

// renderPathParametersFieldDef renders path parameters in field definitions format
func renderPathParametersFieldDef(builder *strings.Builder, params []v3.Parameter) {
	if len(params) == 0 {
		return
	}

	builder.WriteString("#### Path Parameters\n\n")

	for _, param := range params {
		if param.Schema != nil && param.Schema.Schema() != nil {
			schema := param.Schema.Schema()

			// Get type
			typeStr := ""
			if len(schema.Type) > 0 {
				typeStr = schema.Type[0]
			}

			// Get required status
			required := param.Required != nil && *param.Required

			// Format: `paramName` *(type, required)* Description
			builder.WriteString("- `")
			builder.WriteString(param.Name)
			builder.WriteString("` *(")
			builder.WriteString(typeStr)
			if required {
				builder.WriteString(", required")
			}
			builder.WriteString(")*")

			// Description inline
			if param.Description != "" {
				builder.WriteString(" ")
				builder.WriteString(param.Description)
			}

			// Add enums if present
			if len(schema.Enum) > 0 {
				builder.WriteString(" Enums: ")
				for i, enumVal := range schema.Enum {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString("`")
					fmt.Fprintf(builder, "%v", enumVal.Value)
					builder.WriteString("`")
				}
			}

			builder.WriteString("\n")
		}

		builder.WriteString("\n")
	}
}

// renderQueryParametersFieldDef renders query parameters in field definitions format
func renderQueryParametersFieldDef(builder *strings.Builder, params []v3.Parameter) {
	if len(params) == 0 {
		return
	}

	builder.WriteString("#### Query Parameters\n\n")

	for _, param := range params {
		if param.Schema != nil && param.Schema.Schema() != nil {
			schema := param.Schema.Schema()

			// Get type
			typeStr := ""
			if len(schema.Type) > 0 {
				typeStr = schema.Type[0]
			}

			// Get required status
			required := param.Required != nil && *param.Required

			// Format: `paramName` *(type, required)* Description
			builder.WriteString("- `")
			builder.WriteString(param.Name)
			builder.WriteString("` *(")
			builder.WriteString(typeStr)
			if required {
				builder.WriteString(", required")
			}
			builder.WriteString(")*")

			// Description inline
			if param.Description != "" {
				builder.WriteString(" ")
				builder.WriteString(param.Description)
			}

			// Add enums if present
			if len(schema.Enum) > 0 {
				builder.WriteString(" Enums: ")
				for i, enumVal := range schema.Enum {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString("`")
					fmt.Fprintf(builder, "%v", enumVal.Value)
					builder.WriteString("`")
				}
			}

			builder.WriteString("\n")
		}

		builder.WriteString("\n")
	}
}

// renderHeaders renders header parameters in table format
func renderHeaders(builder *strings.Builder, params []v3.Parameter) {
	if len(params) == 0 {
		return
	}

	builder.WriteString("#### Headers\n\n")
	builder.WriteString("Name | Description | Required | Type\n")
	builder.WriteString("-----|-------------|----------|-----\n")

	for _, param := range params {
		builder.WriteString(param.Name)
		builder.WriteString(" | ")

		if param.Description != "" {
			builder.WriteString(param.Description)
		}
		builder.WriteString(" | ")

		if param.Required != nil && *param.Required {
			builder.WriteString("true")
		} else {
			builder.WriteString("false")
		}
		builder.WriteString(" | ")

		if param.Schema != nil && param.Schema.Schema() != nil {
			schema := param.Schema.Schema()
			if len(schema.Type) > 0 {
				builder.WriteString(schema.Type[0])
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
}

// identifySharedResponseSchemas finds schemas used in multiple 2xx responses within the same endpoint
func identifySharedResponseSchemas(op *v3.Operation) map[string][]string {
	if op == nil || op.Responses == nil || op.Responses.Codes == nil {
		return nil
	}

	schemaToResponses := make(map[string][]string)

	// Iterate through all response codes
	for pair := op.Responses.Codes.First(); pair != nil; pair = pair.Next() {
		code := pair.Key()

		// Only consider 2xx responses
		if !strings.HasPrefix(code, "2") {
			continue
		}

		resp := pair.Value()
		if resp == nil || resp.Content == nil {
			continue
		}

		// Find application/json media type
		for contentPair := resp.Content.First(); contentPair != nil; contentPair = contentPair.Next() {
			if contentPair.Key() != "application/json" {
				continue
			}

			mt := contentPair.Value()
			if mt == nil || mt.Schema == nil {
				continue
			}

			if mt.Schema.IsReference() {
				ref := mt.Schema.GetReference()
				schemaName, err := extractSchemaName(ref)
				if err == nil {
					schemaToResponses[schemaName] = append(schemaToResponses[schemaName], code)
				}
			}
		}
	}

	// Return only schemas that appear in 2+ responses
	sharedSchemas := make(map[string][]string)
	for schemaName, responseCodes := range schemaToResponses {
		if len(responseCodes) >= 2 {
			sharedSchemas[schemaName] = responseCodes
		}
	}

	return sharedSchemas
}

func renderResponses(builder *strings.Builder, op *v3.Operation, examples map[string]json.RawMessage, sharedSchemas map[string]schemaUsage) error {
	if op == nil || op.Responses == nil || op.Responses.Codes == nil {
		return nil
	}

	builder.WriteString("### Responses\n\n")

	codes := []string{}
	for pair := op.Responses.Codes.First(); pair != nil; pair = pair.Next() {
		codes = append(codes, pair.Key())
	}
	sort.Strings(codes)

	// Identify shared schemas across 2xx responses within this endpoint
	responseSharedSchemas := identifySharedResponseSchemas(op)

	// Track which schemas we've already rendered field definitions for
	renderedSchemas := make(map[string]bool)

	// Render each response
	for _, code := range codes {
		resp := op.Responses.Codes.GetOrZero(code)

		builder.WriteString("#### ")
		builder.WriteString(code)
		builder.WriteString(" Response\n\n")

		if resp.Description != "" {
			builder.WriteString(resp.Description)
			builder.WriteString("\n\n")
		}

		exampleJSON, err := extractResponseExample(resp, examples)
		if err != nil {
			return err
		}

		if exampleJSON != "" {
			builder.WriteString("```json\n")
			builder.WriteString(exampleJSON)
			builder.WriteString("\n```\n\n")
		}

		// Only render field definitions for 2xx responses
		if strings.HasPrefix(code, "2") {
			// Get schema for this response
			var schemaProxy *base.SchemaProxy
			if resp.Content != nil {
				for pair := resp.Content.First(); pair != nil; pair = pair.Next() {
					if pair.Key() == "application/json" {
						mt := pair.Value()
						if mt != nil && mt.Schema != nil {
							schemaProxy = mt.Schema
							break
						}
					}
				}
			}

			if schemaProxy != nil && schemaProxy.IsReference() {
				schemaName, err := extractSchemaName(schemaProxy.GetReference())
				if err == nil {
					// Check if this schema is shared with other 2xx responses
					if responseCodes, isShared := responseSharedSchemas[schemaName]; isShared && !renderedSchemas[schemaName] {
						// Render field definitions once with note about which responses it applies to
						builder.WriteString("#### Field Definitions (applies to ")
						for i, rc := range responseCodes {
							if i > 0 {
								builder.WriteString(", ")
							}
							builder.WriteString(rc)
						}
						builder.WriteString(" responses)\n\n")

						if err := renderFieldDefinitionsContent(builder, schemaProxy, examples, sharedSchemas); err != nil {
							return err
						}

						renderedSchemas[schemaName] = true
					} else if !isShared {
						// Not shared, render field definitions normally
						builder.WriteString("#### Field Definitions\n\n")
						if err := renderFieldDefinitionsContent(builder, schemaProxy, examples, sharedSchemas); err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}

func collectDebugInfo(model v3.Document, endpoints []endpoint, tagGroups map[string][]endpoint, sharedSchemas map[string]schemaUsage) *DebugInfo {
	debug := &DebugInfo{
		ParameterCounts:   make(map[string]int),
		ResponseCounts:    make(map[string]int),
		NestedSchemaDepth: make(map[string]int),
	}

	if model.Paths != nil && model.Paths.PathItems != nil {
		debug.ParsedPaths = model.Paths.PathItems.Len()
	}

	debug.ExtractedOps = len(endpoints)

	tags := make([]string, 0, len(tagGroups))
	for tag := range tagGroups {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	debug.TagsFound = tags

	for _, e := range endpoints {
		if len(e.tags) == 0 {
			debug.UntaggedOps++
		}

		if e.operation != nil && e.operation.RequestBody != nil {
			debug.RequestBodyCount++
		}

		if e.operation != nil && e.operation.Parameters != nil {
			for _, param := range e.operation.Parameters {
				if param != nil {
					debug.ParameterCounts[param.In]++
				}
			}
		}

		if e.operation != nil && e.operation.Responses != nil && e.operation.Responses.Codes != nil {
			for pair := e.operation.Responses.Codes.First(); pair != nil; pair = pair.Next() {
				debug.ResponseCounts[pair.Key()]++
			}
		}
	}

	debug.SharedSchemaCount = len(sharedSchemas)

	return debug
}

// generateComponentExamples generates JSON examples for all component schemas
func generateComponentExamples(openapi []byte) (map[string]json.RawMessage, error) {
	const maxDepth = 5
	const seed = 42

	result, err := proto.ConvertToExamples(openapi, proto.ExampleOptions{
		IncludeAll: true,
		MaxDepth:   maxDepth,
		Seed:       seed,
	})
	if err != nil {
		return make(map[string]json.RawMessage), nil
	}

	return result.Examples, nil
}

// extractSchemaName extracts schema name from $ref (e.g., "#/components/schemas/Pet" -> "Pet")
func extractSchemaName(ref string) (string, error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return "", fmt.Errorf("invalid schema reference format: %s", ref)
	}

	name := strings.TrimPrefix(ref, prefix)
	if name == "" {
		return "", fmt.Errorf("empty schema name in reference: %s", ref)
	}

	return name, nil
}

// getExampleFromSchema generates example from schema using pre-generated examples
func getExampleFromSchema(schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage) (string, error) {
	if schemaProxy == nil {
		return "", nil
	}

	if !schemaProxy.IsReference() {
		return "", fmt.Errorf("inline schemas not supported in responses, use $ref")
	}

	ref := schemaProxy.GetReference()
	schemaName, err := extractSchemaName(ref)
	if err != nil {
		return "", err
	}

	exampleJSON, found := examples[schemaName]
	if !found {
		return "", nil
	}

	var value interface{}
	if err := json.Unmarshal(exampleJSON, &value); err != nil {
		return "", nil
	}

	formatted, err := json.MarshalIndent(value, "", "   ")
	if err != nil {
		return "", nil
	}

	return string(formatted), nil
}

// getExampleFromMediaType extracts explicit example from MediaType
func getExampleFromMediaType(mt *v3.MediaType) string {
	if mt.Example != nil {
		var value interface{}
		if err := mt.Example.Decode(&value); err == nil {
			if formatted, err := json.MarshalIndent(value, "", "   "); err == nil {
				return string(formatted)
			}
		}
	}

	if mt.Examples != nil && mt.Examples.Len() > 0 {
		pair := mt.Examples.First()
		if pair != nil {
			example := pair.Value()
			if example != nil && example.Value != nil {
				var value interface{}
				if err := example.Value.Decode(&value); err == nil {
					if formatted, err := json.MarshalIndent(value, "", "   "); err == nil {
						return string(formatted)
					}
				}
			}
		}
	}

	return ""
}

// extractResponseExample extracts or generates a JSON example for a response
func extractResponseExample(resp *v3.Response, examples map[string]json.RawMessage) (string, error) {
	if resp.Content == nil || resp.Content.Len() == 0 {
		return "", nil
	}

	for pair := resp.Content.First(); pair != nil; pair = pair.Next() {
		mediaType := pair.Key()
		if mediaType != "application/json" {
			continue
		}

		mt := pair.Value()
		if mt == nil {
			continue
		}

		if explicit := getExampleFromMediaType(mt); explicit != "" {
			return explicit, nil
		}

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
}

// extractRequestExample extracts or generates JSON example for request body
func extractRequestExample(op *v3.Operation, examples map[string]json.RawMessage) (string, error) {
	if op.RequestBody == nil || op.RequestBody.Content == nil {
		return "", nil
	}

	for pair := op.RequestBody.Content.First(); pair != nil; pair = pair.Next() {
		mediaType := pair.Key()
		if mediaType != "application/json" {
			continue
		}

		mt := pair.Value()
		if mt == nil {
			continue
		}

		if explicit := getExampleFromMediaType(mt); explicit != "" {
			return explicit, nil
		}

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
}

// mergeAllOfProperties merges properties from allOf members with the schema's own properties.
// If allOf is empty, returns the schema's own Properties and Required unchanged.
func mergeAllOfProperties(schema *base.Schema) (*orderedmap.Map[string, *base.SchemaProxy], []string) {
	if len(schema.AllOf) == 0 {
		return schema.Properties, schema.Required
	}

	merged := orderedmap.New[string, *base.SchemaProxy]()
	var required []string

	for _, memberProxy := range schema.AllOf {
		if memberProxy == nil {
			continue
		}
		member := memberProxy.Schema()
		if member == nil {
			continue
		}
		if member.Properties != nil {
			for pair := member.Properties.First(); pair != nil; pair = pair.Next() {
				merged.Set(pair.Key(), pair.Value())
			}
		}
		required = append(required, member.Required...)
	}

	// Own properties take precedence over allOf members
	if schema.Properties != nil {
		for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
			merged.Set(pair.Key(), pair.Value())
		}
	}
	required = append(required, schema.Required...)

	return merged, required
}

// renderFieldDefinitionsContent renders the content of field definitions (without the header)
func renderFieldDefinitionsContent(builder *strings.Builder, schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage, sharedSchemas map[string]schemaUsage) error {
	if schemaProxy == nil {
		return nil
	}

	// Check if this schema is shared (only for references)
	if schemaProxy.IsReference() {
		ref := schemaProxy.GetReference()
		schemaName, err := extractSchemaName(ref)
		if err == nil {
			if _, isShared := sharedSchemas[schemaName]; isShared {
				// Render reference to shared schema instead of full documentation
				anchor := makeSchemaAnchor(schemaName)
				builder.WriteString("See [")
				builder.WriteString(schemaName)
				builder.WriteString("](#")
				builder.WriteString(anchor)
				builder.WriteString(")\n\n")
				return nil
			}
		}
	}

	schema := schemaProxy.Schema()
	if schema == nil {
		return nil
	}

	// Handle oneOf schemas (discriminated unions)
	if len(schema.OneOf) > 0 {
		if schema.Discriminator != nil && schema.Discriminator.PropertyName != "" {
			builder.WriteString("Request body is one of the following variants, selected by the `")
			builder.WriteString(schema.Discriminator.PropertyName)
			builder.WriteString("` field:\n\n")
		}

		for _, variantProxy := range schema.OneOf {
			if variantProxy == nil {
				continue
			}
			if variantProxy.IsReference() {
				ref := variantProxy.GetReference()
				name, err := extractSchemaName(ref)
				if err == nil {
					builder.WriteString("**")
					builder.WriteString(name)
					builder.WriteString("**\n")
				}
			}
			if err := renderFieldDefinitionsContent(builder, variantProxy, examples, sharedSchemas); err != nil {
				return err
			}
		}

		return nil
	}

	mergedProps, _ := mergeAllOfProperties(schema)
	if mergedProps == nil || mergedProps.Len() == 0 {
		return nil
	}

	const maxDepth = 10
	visited := make(map[string]int)

	fields, nestedDefs, err := extractSchemaFields(schemaProxy, examples, visited, maxDepth)
	if err != nil {
		return err
	}

	// Render top-level fields
	for _, field := range fields {
		builder.WriteString("- `")
		builder.WriteString(field.name)
		builder.WriteString("`")

		if field.typeStr != "" {
			builder.WriteString(" *(")

			if field.isArray && !field.isObject {
				builder.WriteString(field.typeStr)
				builder.WriteString(" array")
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
				builder.WriteString(field.typeStr)
			}

			if field.required {
				builder.WriteString(", required")
			}
			builder.WriteString(")*")
		}

		if field.description != "" {
			builder.WriteString(" ")
			builder.WriteString(field.description)
		} else if !field.isObject {
			// Warn about missing description for non-object fields
			log.Printf("Warning: Field '%s' is missing a description", field.name)
		}

		if len(field.enum) > 0 {
			builder.WriteString(" Enums: ")
			for i, enumVal := range field.enum {
				if i > 0 {
					builder.WriteString(", ")
				}
				builder.WriteString("`")
				fmt.Fprintf(builder, "%v", enumVal)
				builder.WriteString("`")
			}
		}

		builder.WriteString("\n")
	}

	builder.WriteString("\n")

	// Render nested schema definitions
	for _, nestedDef := range nestedDefs {
		if err := renderSchemaDefinition(builder, nestedDef); err != nil {
			return err
		}
	}

	return nil
}

// renderFieldDefinitions renders field definitions section for a schema
func renderFieldDefinitions(builder *strings.Builder, schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage, sharedSchemas map[string]schemaUsage) error {
	if schemaProxy == nil {
		return nil
	}

	if !schemaProxy.IsReference() {
		return nil
	}

	schema := schemaProxy.Schema()
	if schema == nil {
		return nil
	}

	// Handle oneOf schemas (discriminated unions)
	if len(schema.OneOf) > 0 {
		builder.WriteString("#### Field Definitions\n\n")

		if schema.Discriminator != nil && schema.Discriminator.PropertyName != "" {
			builder.WriteString("Request body is one of the following variants, selected by the `")
			builder.WriteString(schema.Discriminator.PropertyName)
			builder.WriteString("` field:\n\n")
		}

		for _, variantProxy := range schema.OneOf {
			if variantProxy == nil {
				continue
			}
			if variantProxy.IsReference() {
				ref := variantProxy.GetReference()
				name, err := extractSchemaName(ref)
				if err == nil {
					builder.WriteString("**")
					builder.WriteString(name)
					builder.WriteString("**\n")
				}
			}
			if err := renderFieldDefinitionsContent(builder, variantProxy, examples, sharedSchemas); err != nil {
				return err
			}
		}

		return nil
	}

	mergedProps, _ := mergeAllOfProperties(schema)
	if mergedProps == nil || mergedProps.Len() == 0 {
		return nil
	}

	builder.WriteString("#### Field Definitions\n\n")

	return renderFieldDefinitionsContent(builder, schemaProxy, examples, sharedSchemas)
}

// renderRequestBody renders request section with JSON and field definitions
func renderRequestBody(builder *strings.Builder, op *v3.Operation, examples map[string]json.RawMessage, sharedSchemas map[string]schemaUsage) error {
	if op == nil || op.RequestBody == nil {
		return nil
	}

	exampleJSON, err := extractRequestExample(op, examples)
	if err != nil {
		return err
	}

	// Determine if a JSON schema exists for field definitions
	hasSchema := false
	if op.RequestBody.Content != nil {
		for pair := op.RequestBody.Content.First(); pair != nil; pair = pair.Next() {
			if pair.Key() == "application/json" {
				mt := pair.Value()
				if mt != nil && mt.Schema != nil {
					hasSchema = true
					break
				}
			}
		}
	}

	// Nothing to render if there is no example and no schema
	if exampleJSON == "" && !hasSchema {
		return nil
	}

	builder.WriteString("### Request\n\n")

	if exampleJSON != "" {
		builder.WriteString("```json\n")
		builder.WriteString(exampleJSON)
		builder.WriteString("\n```\n\n")
	}

	if op.RequestBody.Content != nil {
		for pair := op.RequestBody.Content.First(); pair != nil; pair = pair.Next() {
			mediaType := pair.Key()
			if mediaType != "application/json" {
				continue
			}

			mt := pair.Value()
			if mt != nil && mt.Schema != nil {
				if err := renderFieldDefinitions(builder, mt.Schema, examples, sharedSchemas); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// extractSchemaFieldsFromProperties extracts field information directly from schema properties
func extractSchemaFieldsFromProperties(schema *base.Schema, visited map[string]int, maxDepth int) ([]schemaField, []schemaDefinition, error) {
	if schema == nil {
		return nil, nil, nil
	}

	mergedProps, mergedRequired := mergeAllOfProperties(schema)
	if mergedProps == nil || mergedProps.Len() == 0 {
		return nil, nil, nil
	}

	// Check max depth
	depth := 0
	for _, v := range visited {
		if v > depth {
			depth = v
		}
	}
	if depth >= maxDepth {
		return nil, nil, nil
	}

	var fields []schemaField
	var nestedDefs []schemaDefinition

	requiredFields := make(map[string]bool)
	for _, req := range mergedRequired {
		requiredFields[req] = true
	}

	for pair := mergedProps.First(); pair != nil; pair = pair.Next() {
		fieldName := pair.Key()
		propSchema := pair.Value()

		if propSchema == nil || propSchema.Schema() == nil {
			continue
		}

		prop := propSchema.Schema()

		field := schemaField{
			name:        fieldName,
			required:    requiredFields[fieldName],
			description: prop.Description,
		}

		if len(prop.Enum) > 0 {
			for _, enumVal := range prop.Enum {
				field.enum = append(field.enum, enumVal.Value)
			}
		}

		if len(prop.Type) > 0 {
			field.typeStr = prop.Type[0]

			if prop.Type[0] == "array" && prop.Items != nil && prop.Items.IsA() {
				field.isArray = true
				itemSchema := prop.Items.A.Schema()
				if itemSchema != nil {
					if len(itemSchema.Type) > 0 {
						field.typeStr = itemSchema.Type[0]
					}

					// If array items are objects with a reference, handle recursively
					if prop.Items.A.IsReference() {
						itemRef := prop.Items.A.GetReference()
						itemSchemaName, err := extractSchemaName(itemRef)
						if err == nil {
							field.nestedSchemaRef = itemSchemaName
							field.isObject = true

							// Check recursion
							if visited[itemSchemaName] <= 1 {
								visited[itemSchemaName]++
								itemSchemaActual := prop.Items.A.Schema()
								if itemSchemaActual != nil {
									nestedFields, nestedNested, err := extractSchemaFieldsFromProperties(itemSchemaActual, visited, maxDepth)
									if err != nil {
										return nil, nil, err
									}

									if len(nestedFields) > 0 {
										nestedDef := schemaDefinition{
											name:   itemSchemaName,
											fields: nestedFields,
										}
										nestedDefs = append(nestedDefs, nestedDef)
										nestedDefs = append(nestedDefs, nestedNested...)
									}
								}
								visited[itemSchemaName]--
							}
						}
					}
				}
			} else if prop.Type[0] == "object" {
				field.isObject = true

				// Check if this is a reference to another schema
				if propSchema.IsReference() {
					propRef := propSchema.GetReference()
					nestedSchemaName, err := extractSchemaName(propRef)
					if err == nil {
						field.nestedSchemaRef = nestedSchemaName

						// Check recursion
						if visited[nestedSchemaName] <= 1 {
							visited[nestedSchemaName]++
							nestedFields, nestedNested, err := extractSchemaFieldsFromProperties(prop, visited, maxDepth)
							if err != nil {
								return nil, nil, err
							}

							if len(nestedFields) > 0 {
								nestedDef := schemaDefinition{
									name:   nestedSchemaName,
									fields: nestedFields,
								}
								nestedDefs = append(nestedDefs, nestedDef)
								nestedDefs = append(nestedDefs, nestedNested...)
							}
							visited[nestedSchemaName]--
						}
					}
				}
			}
		}

		fields = append(fields, field)
	}

	return fields, nestedDefs, nil
}

// extractSchemaFields recursively extracts field information from schema
func extractSchemaFields(schemaProxy *base.SchemaProxy, examples map[string]json.RawMessage, visited map[string]int, maxDepth int) ([]schemaField, []schemaDefinition, error) {
	if schemaProxy == nil {
		return nil, nil, nil
	}

	if !schemaProxy.IsReference() {
		return nil, nil, nil
	}

	ref := schemaProxy.GetReference()
	schemaName, err := extractSchemaName(ref)
	if err != nil {
		return nil, nil, err
	}

	// Check recursion depth
	if visited[schemaName] > 1 {
		return nil, nil, nil
	}

	// Check max depth
	depth := 0
	for _, v := range visited {
		if v > depth {
			depth = v
		}
	}
	if depth >= maxDepth {
		return nil, nil, nil
	}

	schema := schemaProxy.Schema()
	if schema == nil {
		return nil, nil, nil
	}

	mergedProps, mergedRequired := mergeAllOfProperties(schema)
	if mergedProps == nil || mergedProps.Len() == 0 {
		return nil, nil, nil
	}

	visited[schemaName]++
	defer func() { visited[schemaName]-- }()

	var fields []schemaField
	var nestedDefs []schemaDefinition

	requiredFields := make(map[string]bool)
	for _, req := range mergedRequired {
		requiredFields[req] = true
	}

	for pair := mergedProps.First(); pair != nil; pair = pair.Next() {
		fieldName := pair.Key()
		propSchema := pair.Value()

		if propSchema == nil || propSchema.Schema() == nil {
			continue
		}

		prop := propSchema.Schema()

		field := schemaField{
			name:        fieldName,
			required:    requiredFields[fieldName],
			description: prop.Description,
		}

		if len(prop.Enum) > 0 {
			for _, enumVal := range prop.Enum {
				field.enum = append(field.enum, enumVal.Value)
			}
		}

		if len(prop.Type) > 0 {
			field.typeStr = prop.Type[0]

			if prop.Type[0] == "array" && prop.Items != nil && prop.Items.IsA() {
				field.isArray = true
				itemSchema := prop.Items.A.Schema()
				if itemSchema != nil {
					if len(itemSchema.Type) > 0 {
						field.typeStr = itemSchema.Type[0]
					}

					// If array items are objects with a reference, handle recursively
					if prop.Items.A.IsReference() {
						itemRef := prop.Items.A.GetReference()
						itemSchemaName, err := extractSchemaName(itemRef)
						if err == nil {
							field.nestedSchemaRef = itemSchemaName
							field.isObject = true

							// Recursively extract nested schema
							nestedFields, nestedNested, err := extractSchemaFields(prop.Items.A, examples, visited, maxDepth)
							if err != nil {
								return nil, nil, err
							}

							if len(nestedFields) > 0 {
								nestedDef := schemaDefinition{
									name:   itemSchemaName,
									fields: nestedFields,
								}
								nestedDefs = append(nestedDefs, nestedDef)
								nestedDefs = append(nestedDefs, nestedNested...)
							}
						}
					}
				}
			} else if prop.Type[0] == "object" {
				field.isObject = true

				// Check if this is a reference to another schema
				if propSchema.IsReference() {
					propRef := propSchema.GetReference()
					nestedSchemaName, err := extractSchemaName(propRef)
					if err == nil {
						field.nestedSchemaRef = nestedSchemaName

						// Recursively extract nested schema
						nestedFields, nestedNested, err := extractSchemaFields(propSchema, examples, visited, maxDepth)
						if err != nil {
							return nil, nil, err
						}

						if len(nestedFields) > 0 {
							nestedDef := schemaDefinition{
								name:   nestedSchemaName,
								fields: nestedFields,
							}
							nestedDefs = append(nestedDefs, nestedDef)
							nestedDefs = append(nestedDefs, nestedNested...)
						}
					}
				}
			}
		}

		fields = append(fields, field)
	}

	return fields, nestedDefs, nil
}

// renderSchemaDefinition renders a single schema definition section
func renderSchemaDefinition(builder *strings.Builder, def schemaDefinition) error {
	builder.WriteString("**")
	builder.WriteString(def.name)
	builder.WriteString("**\n")

	for _, field := range def.fields {
		builder.WriteString("- `")
		builder.WriteString(field.name)
		builder.WriteString("`")

		if field.typeStr != "" {
			builder.WriteString(" *(")

			if field.isArray && !field.isObject {
				builder.WriteString(field.typeStr)
				builder.WriteString(" array")
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
				builder.WriteString(field.typeStr)
			}

			if field.required {
				builder.WriteString(", required")
			}
			builder.WriteString(")*")
		}

		if field.description != "" {
			builder.WriteString(": ")
			builder.WriteString(field.description)

			if len(field.enum) > 0 {
				builder.WriteString(". Enums: ")
				for i, enumVal := range field.enum {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString("`")
					fmt.Fprintf(builder, "%v", enumVal)
					builder.WriteString("`")
				}
			}
		}

		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	return nil
}
