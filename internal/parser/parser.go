package parser

import (
	"context"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

// Parser defines the interface for schema parsers
type Parser interface {
	Parse(filepath string) (*Schema, error)
}

// Schema represents a parsed API schema
type Schema struct {
	Type    string                // "openapi" or "graphql"
	Version string                // e.g., "3.0.0"
	Title   string                // API title
	Paths   map[string][]Endpoint // Path -> methods
	Raw     interface{}           // Original parsed object
}

// Endpoint represents a single API endpoint
type Endpoint struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Parameters  []Parameter
}

// Parameter represents an API parameter
type Parameter struct {
	Name     string
	In       string // "query", "path", "header", etc.
	Required bool
	Type     string
}

// OpenAPIParser implements Parser for OpenAPI 3.x specifications
type OpenAPIParser struct{}

// NewOpenAPIParser creates a new OpenAPI parser
func NewOpenAPIParser() *OpenAPIParser {
	return &OpenAPIParser{}
}

// Parse reads and parses an OpenAPI 3.x specification file
func (p *OpenAPIParser) Parse(filepath string) (*Schema, error) {
	// Read the file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the OpenAPI document
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	// Validate the document
	ctx := context.Background()
	if err := doc.Validate(ctx); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	// Convert to our Schema format
	schema := &Schema{
		Type:    "openapi",
		Version: doc.OpenAPI,
		Title:   doc.Info.Title,
		Paths:   make(map[string][]Endpoint),
		Raw:     doc,
	}

	// Extract endpoints
	for path, pathItem := range doc.Paths.Map() {
		var endpoints []Endpoint

		for method, operation := range pathItem.Operations() {
			endpoint := Endpoint{
				Method:      method,
				Path:        path,
				Summary:     operation.Summary,
				Description: operation.Description,
				Parameters:  extractParameters(operation),
			}
			endpoints = append(endpoints, endpoint)
		}

		if len(endpoints) > 0 {
			schema.Paths[path] = endpoints
		}
	}

	return schema, nil
}

// extractParameters converts OpenAPI parameters to our simplified format
func extractParameters(operation *openapi3.Operation) []Parameter {
	var params []Parameter

	for _, paramRef := range operation.Parameters {
		if paramRef.Value == nil {
			continue
		}

		param := Parameter{
			Name:     paramRef.Value.Name,
			In:       paramRef.Value.In,
			Required: paramRef.Value.Required,
		}

		// Extract type from schema if available
		if paramRef.Value.Schema != nil && paramRef.Value.Schema.Value != nil {
			param.Type = paramRef.Value.Schema.Value.Type.Slice()[0]
		}

		params = append(params, param)
	}

	return params
}
