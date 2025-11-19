package conv

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

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

	endpoints := extractEndpoints(v3Model.Model)
	tagGroups := groupByTags(endpoints)
	markdown := generateMarkdown(opts, endpoints, tagGroups)

	return &ConvertResult{
		Markdown:      []byte(markdown),
		EndpointCount: len(endpoints),
		TagCount:      len(tagGroups),
	}, nil
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

func generateMarkdown(opts ConvertOptions, endpoints []endpoint, tagGroups map[string][]endpoint) string {
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
			}
		}
	}

	return builder.String()
}
