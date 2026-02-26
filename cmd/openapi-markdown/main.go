package main

import (
	"flag"
	"fmt"
	"os"

	conv "github.com/duh-rpc/openapi-markdown.go"
)

func main() {
	title := flag.String("title", "", "API documentation title")
	description := flag.String("description", "", "API documentation description")
	output := flag.String("o", "", "output file path (defaults to stdout)")
	sharedSchemas := flag.Bool("shared-schemas", false, "enable shared schema definitions")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: openapi-markdown [flags] <openapi-file>\n\nFlags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	openapi, err := os.ReadFile(flag.Arg(0))
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

	if *output != "" {
		if err := os.WriteFile(*output, result.Markdown, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if _, err := os.Stdout.Write(result.Markdown); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
		os.Exit(1)
	}
}
