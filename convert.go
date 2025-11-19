package conv

import (
	"fmt"
	"strings"

	"github.com/pb33f/libopenapi"
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

	markdown := generateMarkdown(opts)

	return &ConvertResult{
		Markdown:      []byte(markdown),
		EndpointCount: 0,
		TagCount:      0,
	}, nil
}

func generateMarkdown(opts ConvertOptions) string {
	var builder strings.Builder

	builder.WriteString("# ")
	builder.WriteString(opts.Title)
	builder.WriteString("\n\n")

	if opts.Description != "" {
		builder.WriteString(opts.Description)
		builder.WriteString("\n")
	}

	return builder.String()
}
