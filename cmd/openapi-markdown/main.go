package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	conv "github.com/duh-rpc/openapi-markdown.go"
)

func main() {
	title := flag.String("title", "", "API documentation title (defaults to input filename)")
	description := flag.String("description", "", "API documentation description")
	output := flag.String("o", "", "output file path (defaults to input filename with .md extension)")
	sharedSchemas := flag.Bool("shared-schemas", false, "enable shared schema definitions")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: openapi-markdown [flags] <openapi-file>\n\nConverts an OpenAPI 3.x YAML file to markdown documentation.\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)
	baseName := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	if *title == "" {
		*title = baseName
	}

	if *output == "" {
		*output = baseName + ".md"
	}

	openapi, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	result, err := conv.Convert(openapi, conv.ConvertOptions{
		Title:               *title,
		Description:         *description,
		EnableSharedSchemas: *sharedSchemas,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, result.Markdown, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Wrote %s\n", *output)
}
