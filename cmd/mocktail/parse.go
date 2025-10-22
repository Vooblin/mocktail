package main

import (
	"fmt"

	"github.com/Vooblin/mocktail/internal/parser"
	"github.com/spf13/cobra"
)

func newParseCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "parse <schema-file>",
		Short: "Parse and validate an API schema",
		Long: `Parse an OpenAPI 3.x or GraphQL schema file and validate its structure.

This command reads the schema file, validates it according to the specification,
and displays a summary of the parsed content.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filepath := args[0]

			// Create parser based on file extension or content
			// For now, we only support OpenAPI
			parser := parser.NewOpenAPIParser()

			schema, err := parser.Parse(filepath)
			if err != nil {
				return fmt.Errorf("failed to parse schema: %w", err)
			}

			// Display summary
			fmt.Printf("✓ Successfully parsed %s schema\n\n", schema.Type)
			fmt.Printf("Title:   %s\n", schema.Title)
			fmt.Printf("Version: %s\n", schema.Version)
			fmt.Printf("Paths:   %d\n\n", len(schema.Paths))

			if outputFormat == "verbose" {
				fmt.Println("Endpoints:")
				for path, endpoints := range schema.Paths {
					for _, endpoint := range endpoints {
						fmt.Printf("  %s %s\n", endpoint.Method, path)
						if endpoint.Summary != "" {
							fmt.Printf("    Summary: %s\n", endpoint.Summary)
						}
						if len(endpoint.Parameters) > 0 {
							fmt.Printf("    Parameters: %d\n", len(endpoint.Parameters))
						}
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "summary", "Output format (summary|verbose)")

	return cmd
}
