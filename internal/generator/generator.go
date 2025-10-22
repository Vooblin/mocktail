package generator

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Generator creates mock data from OpenAPI schemas
type Generator struct {
	rng *rand.Rand
}

// NewGenerator creates a new generator with a seed for reproducibility
func NewGenerator(seed int64) *Generator {
	return &Generator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// GenerateFromSchema generates mock data from an OpenAPI schema
func (g *Generator) GenerateFromSchema(schema *openapi3.Schema) (interface{}, error) {
	if schema == nil {
		return nil, fmt.Errorf("schema is nil")
	}

	// Handle schema references
	if schema.Type == nil || len(schema.Type.Slice()) == 0 {
		// Default to object if no type specified
		return g.generateObject(schema)
	}

	schemaType := schema.Type.Slice()[0]

	switch schemaType {
	case "string":
		return g.generateString(schema), nil
	case "integer":
		return g.generateInteger(schema), nil
	case "number":
		return g.generateNumber(schema), nil
	case "boolean":
		return g.generateBoolean(), nil
	case "array":
		return g.generateArray(schema)
	case "object":
		return g.generateObject(schema)
	default:
		return nil, fmt.Errorf("unsupported schema type: %s", schemaType)
	}
}

// generateString generates a string value based on format and constraints
func (g *Generator) generateString(schema *openapi3.Schema) string {
	// Check for enum values
	if len(schema.Enum) > 0 {
		idx := g.rng.Intn(len(schema.Enum))
		if str, ok := schema.Enum[idx].(string); ok {
			return str
		}
	}

	// Generate based on format
	switch schema.Format {
	case "date-time":
		return time.Now().Add(-time.Duration(g.rng.Intn(365*24)) * time.Hour).Format(time.RFC3339)
	case "date":
		return time.Now().Add(-time.Duration(g.rng.Intn(365)) * 24 * time.Hour).Format("2006-01-02")
	case "email":
		return fmt.Sprintf("user%d@example.com", g.rng.Intn(1000))
	case "uuid":
		return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
			g.rng.Uint32(),
			uint16(g.rng.Uint32()),
			uint16(g.rng.Uint32())|0x4000,
			uint16(g.rng.Uint32())|0x8000,
			uint64(g.rng.Uint32())<<16|uint64(g.rng.Uint32()>>16))
	case "uri":
		return fmt.Sprintf("https://example.com/resource/%d", g.rng.Intn(1000))
	default:
		// Generate a generic string
		words := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "theta"}
		return words[g.rng.Intn(len(words))]
	}
}

// generateInteger generates an integer value respecting min/max constraints
func (g *Generator) generateInteger(schema *openapi3.Schema) int64 {
	min := int64(0)
	max := int64(100)

	if schema.Min != nil {
		min = int64(*schema.Min)
	}
	if schema.Max != nil {
		max = int64(*schema.Max)
	}

	if max <= min {
		return min
	}

	return min + int64(g.rng.Int63n(max-min+1))
}

// generateNumber generates a floating-point number
func (g *Generator) generateNumber(schema *openapi3.Schema) float64 {
	min := 0.0
	max := 100.0

	if schema.Min != nil {
		min = *schema.Min
	}
	if schema.Max != nil {
		max = *schema.Max
	}

	if max <= min {
		return min
	}

	return min + g.rng.Float64()*(max-min)
}

// generateBoolean generates a random boolean value
func (g *Generator) generateBoolean() bool {
	return g.rng.Intn(2) == 1
}

// generateArray generates an array of values
func (g *Generator) generateArray(schema *openapi3.Schema) ([]interface{}, error) {
	if schema.Items == nil || schema.Items.Value == nil {
		return []interface{}{}, nil
	}

	// Determine array length (default 2-5 items)
	minItems := 2
	maxItems := 5

	if schema.MinItems > 0 {
		minItems = int(schema.MinItems)
	}
	if schema.MaxItems != nil && *schema.MaxItems > 0 {
		maxItems = int(*schema.MaxItems)
	}

	length := minItems
	if maxItems > minItems {
		length = minItems + g.rng.Intn(maxItems-minItems+1)
	}

	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		item, err := g.GenerateFromSchema(schema.Items.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to generate array item: %w", err)
		}
		result[i] = item
	}

	return result, nil
}

// generateObject generates an object with properties
func (g *Generator) generateObject(schema *openapi3.Schema) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	if schema.Properties == nil {
		return result, nil
	}

	for propName, propRef := range schema.Properties {
		if propRef.Value == nil {
			continue
		}

		value, err := g.GenerateFromSchema(propRef.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to generate property %s: %w", propName, err)
		}
		result[propName] = value
	}

	return result, nil
}

// GenerateResponse generates a mock response for an OpenAPI operation
func (g *Generator) GenerateResponse(operation *openapi3.Operation, statusCode string) (interface{}, error) {
	if operation == nil || operation.Responses == nil {
		return nil, fmt.Errorf("operation or responses is nil")
	}

	responseRef := operation.Responses.Value(statusCode)
	if responseRef == nil {
		return nil, fmt.Errorf("no response defined for status code %s", statusCode)
	}

	response := responseRef.Value
	if response == nil || response.Content == nil {
		return map[string]interface{}{}, nil
	}

	// Look for application/json content
	jsonContent := response.Content.Get("application/json")
	if jsonContent == nil || jsonContent.Schema == nil || jsonContent.Schema.Value == nil {
		return map[string]interface{}{}, nil
	}

	return g.GenerateFromSchema(jsonContent.Schema.Value)
}
