package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Vooblin/mocktail/internal/generator"
	"github.com/Vooblin/mocktail/internal/parser"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	var (
		path   string
		method string
		seed   int64
		count  int
	)

	cmd := &cobra.Command{
		Use:   "generate <schema-file>",
		Short: "Generate test payloads from OpenAPI schema",
		Long: `Generate realistic test payloads from OpenAPI schema definitions.

This command creates sample request and response payloads based on your OpenAPI schema,
useful for contract testing, API documentation, and integration tests.

Examples:
  # Generate a response for GET /pets
  mocktail generate examples/petstore.yaml --path /pets --method GET

  # Generate a request body for POST /pets
  mocktail generate examples/petstore.yaml --path /pets --method POST

  # Generate multiple samples with custom seed
  mocktail generate examples/petstore.yaml --path /pets --method GET --count 3 --seed 42`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaFile := args[0]

			// Parse the schema
			p := parser.NewOpenAPIParser()
			schema, err := p.Parse(schemaFile)
			if err != nil {
				return fmt.Errorf("failed to parse schema: %w", err)
			}

			// Validate path and method
			if path == "" {
				return fmt.Errorf("--path flag is required")
			}
			if method == "" {
				return fmt.Errorf("--method flag is required")
			}

			// Find the endpoint
			endpoints, exists := schema.Paths[path]
			if !exists {
				return fmt.Errorf("path %s not found in schema", path)
			}

			var endpoint *parser.Endpoint
			for _, ep := range endpoints {
				if ep.Method == method {
					endpoint = &ep
					break
				}
			}

			if endpoint == nil {
				return fmt.Errorf("method %s not found for path %s", method, path)
			}

			// Use current time as default seed if not specified
			if seed == 0 {
				seed = time.Now().UnixNano()
			}

			// Get the OpenAPI document
			doc, ok := schema.Raw.(*openapi3.T)
			if !ok {
				return fmt.Errorf("invalid schema format")
			}

			pathItem := doc.Paths.Find(path)
			if pathItem == nil {
				return fmt.Errorf("path item not found")
			}

			operation := pathItem.Operations()[method]
			if operation == nil {
				return fmt.Errorf("operation not found")
			}

			// Generate payloads
			fmt.Printf("Generating %d payload(s) for %s %s (seed: %d)\n\n", count, method, path, seed)

			for i := 0; i < count; i++ {
				gen := generator.NewGenerator(seed + int64(i))

				// Generate request body if this is a POST/PUT/PATCH
				if method == "POST" || method == "PUT" || method == "PATCH" {
					if operation.RequestBody != nil && operation.RequestBody.Value != nil {
						jsonContent := operation.RequestBody.Value.Content.Get("application/json")
						if jsonContent != nil && jsonContent.Schema != nil {
							fmt.Printf("=== Request Body #%d ===\n", i+1)
							payload, err := gen.GenerateFromSchema(jsonContent.Schema.Value)
							if err != nil {
								return fmt.Errorf("failed to generate request body: %w", err)
							}

							jsonData, err := json.MarshalIndent(payload, "", "  ")
							if err != nil {
								return fmt.Errorf("failed to marshal JSON: %w", err)
							}
							fmt.Println(string(jsonData))
							fmt.Println()
						}
					}
				}

				// Generate response for 200/201 status
				var responseSchema *openapi3.Schema
				if operation.Responses != nil {
					if resp := operation.Responses.Status(200); resp != nil && resp.Value != nil {
						if jsonContent := resp.Value.Content.Get("application/json"); jsonContent != nil {
							responseSchema = jsonContent.Schema.Value
						}
					} else if resp := operation.Responses.Status(201); resp != nil && resp.Value != nil {
						if jsonContent := resp.Value.Content.Get("application/json"); jsonContent != nil {
							responseSchema = jsonContent.Schema.Value
						}
					}
				}

				if responseSchema != nil {
					fmt.Printf("=== Response Body #%d ===\n", i+1)
					payload, err := gen.GenerateFromSchema(responseSchema)
					if err != nil {
						return fmt.Errorf("failed to generate response body: %w", err)
					}

					jsonData, err := json.MarshalIndent(payload, "", "  ")
					if err != nil {
						return fmt.Errorf("failed to marshal JSON: %w", err)
					}
					fmt.Println(string(jsonData))
					fmt.Println()
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "path", "p", "", "API path (e.g., /pets)")
	cmd.Flags().StringVarP(&method, "method", "m", "", "HTTP method (e.g., GET, POST)")
	cmd.Flags().Int64VarP(&seed, "seed", "s", 0, "Random seed for reproducible output (default: current time)")
	cmd.Flags().IntVarP(&count, "count", "c", 1, "Number of payloads to generate")

	return cmd
}
