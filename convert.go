package conv

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	proto "github.com/duh-rpc/openapi-proto.go"
	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ConvertResult contains markdown output and generation metadata
type ConvertResult struct {
	Markdown      []byte
	EndpointCount int
	TagCount      int
	Debug         *DebugInfo
}

// DebugInfo provides visibility into internal conversion process for testing
type DebugInfo struct {
	ParsedPaths     int
	ExtractedOps    int
	TagsFound       []string
	UntaggedOps     int
	ParameterCounts map[string]int
	ResponseCounts  map[string]int
}

// ConvertOptions configures markdown generation
type ConvertOptions struct {
	Title       string
	Description string
	Debug       bool
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
	markdown, err := generateMarkdown(opts, endpoints, tagGroups, examples)
	if err != nil {
		return nil, err
	}

	result := &ConvertResult{
		Markdown:      []byte(markdown),
		EndpointCount: len(endpoints),
		TagCount:      len(tagGroups),
	}

	if opts.Debug {
		result.Debug = collectDebugInfo(v3Model.Model, endpoints, tagGroups)
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

func makeAnchor(method, path string) string {
	combined := method + " " + path
	combined = strings.ToLower(combined)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	combined = reg.ReplaceAllString(combined, "")

	return combined
}

func generateMarkdown(opts ConvertOptions, endpoints []endpoint, tagGroups map[string][]endpoint, examples map[string]json.RawMessage) (string, error) {
	var builder strings.Builder

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
				builder.WriteString(e.summary)
				builder.WriteString("\n\n")
				if e.description != "" && e.description != e.summary {
					builder.WriteString(e.description)
					builder.WriteString("\n\n")
				}

				renderParameters(&builder, e.operation)
				if err := renderResponses(&builder, e.operation, examples); err != nil {
					return "", err
				}
			}
		}
	}

	return builder.String(), nil
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

	renderParamTable(builder, "Path", pathParams)
	renderParamTable(builder, "Query", queryParams)
	renderParamTable(builder, "Header", headerParams)
}

func renderParamTable(builder *strings.Builder, paramType string, params []v3.Parameter) {
	if len(params) == 0 {
		return
	}

	builder.WriteString("#### ")
	builder.WriteString(paramType)
	builder.WriteString(" Parameters\n\n")
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

func renderResponses(builder *strings.Builder, op *v3.Operation, examples map[string]json.RawMessage) error {
	if op == nil || op.Responses == nil || op.Responses.Codes == nil {
		return nil
	}

	codes := []string{}
	for pair := op.Responses.Codes.First(); pair != nil; pair = pair.Next() {
		codes = append(codes, pair.Key())
	}
	sort.Strings(codes)

	for _, code := range codes {
		resp := op.Responses.Codes.GetOrZero(code)

		builder.WriteString("##### ")
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
	}

	return nil
}

func collectDebugInfo(model v3.Document, endpoints []endpoint, tagGroups map[string][]endpoint) *DebugInfo {
	debug := &DebugInfo{
		ParameterCounts: make(map[string]int),
		ResponseCounts:  make(map[string]int),
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
