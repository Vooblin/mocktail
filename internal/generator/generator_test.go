package generator

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(42)
	if gen == nil {
		t.Fatal("Expected generator to be created")
	}
	if gen.rng == nil {
		t.Error("Expected rng to be initialized")
	}
}

func TestGenerateString(t *testing.T) {
	gen := NewGenerator(42)

	tests := []struct {
		name   string
		schema *openapi3.Schema
		check  func(t *testing.T, result string)
	}{
		{
			name: "basic string",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
			check: func(t *testing.T, result string) {
				if result == "" {
					t.Error("Expected non-empty string")
				}
			},
		},
		{
			name: "email format",
			schema: &openapi3.Schema{
				Type:   &openapi3.Types{"string"},
				Format: "email",
			},
			check: func(t *testing.T, result string) {
				if !contains(result, "@example.com") {
					t.Errorf("Expected email format, got: %s", result)
				}
			},
		},
		{
			name: "uuid format",
			schema: &openapi3.Schema{
				Type:   &openapi3.Types{"string"},
				Format: "uuid",
			},
			check: func(t *testing.T, result string) {
				if len(result) != 36 {
					t.Errorf("Expected UUID length 36, got %d: %s", len(result), result)
				}
			},
		},
		{
			name: "date-time format",
			schema: &openapi3.Schema{
				Type:   &openapi3.Types{"string"},
				Format: "date-time",
			},
			check: func(t *testing.T, result string) {
				// RFC3339 should contain T separator
				if !contains(result, "T") {
					t.Errorf("Expected RFC3339 date-time format with T separator, got: %s", result)
				}
			},
		},
		{
			name: "enum string",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
				Enum: []interface{}{"dog", "cat", "bird"},
			},
			check: func(t *testing.T, result string) {
				validValues := map[string]bool{"dog": true, "cat": true, "bird": true}
				if !validValues[result] {
					t.Errorf("Expected one of [dog, cat, bird], got: %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.generateString(tt.schema)
			tt.check(t, result)
		})
	}
}

func TestGenerateInteger(t *testing.T) {
	gen := NewGenerator(42)

	tests := []struct {
		name   string
		schema *openapi3.Schema
		check  func(t *testing.T, result int64)
	}{
		{
			name: "basic integer",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"integer"},
			},
			check: func(t *testing.T, result int64) {
				if result < 0 || result > 100 {
					t.Errorf("Expected integer in range [0, 100], got: %d", result)
				}
			},
		},
		{
			name: "integer with min/max",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"integer"},
				Min:  float64Ptr(10),
				Max:  float64Ptr(20),
			},
			check: func(t *testing.T, result int64) {
				if result < 10 || result > 20 {
					t.Errorf("Expected integer in range [10, 20], got: %d", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.generateInteger(tt.schema)
			tt.check(t, result)
		})
	}
}

func TestGenerateNumber(t *testing.T) {
	gen := NewGenerator(42)

	tests := []struct {
		name   string
		schema *openapi3.Schema
		check  func(t *testing.T, result float64)
	}{
		{
			name: "basic number",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"number"},
			},
			check: func(t *testing.T, result float64) {
				if result < 0 || result > 100 {
					t.Errorf("Expected number in range [0, 100], got: %f", result)
				}
			},
		},
		{
			name: "number with min/max",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"number"},
				Min:  float64Ptr(5.5),
				Max:  float64Ptr(10.5),
			},
			check: func(t *testing.T, result float64) {
				if result < 5.5 || result > 10.5 {
					t.Errorf("Expected number in range [5.5, 10.5], got: %f", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.generateNumber(tt.schema)
			tt.check(t, result)
		})
	}
}

func TestGenerateBoolean(t *testing.T) {
	gen := NewGenerator(42)
	result := gen.generateBoolean()
	if result != true && result != false {
		t.Errorf("Expected boolean value, got: %v", result)
	}
}

func TestGenerateArray(t *testing.T) {
	gen := NewGenerator(42)

	tests := []struct {
		name   string
		schema *openapi3.Schema
		check  func(t *testing.T, result []interface{}, err error)
	}{
		{
			name: "array of strings",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"array"},
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
					},
				},
			},
			check: func(t *testing.T, result []interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if len(result) < 2 || len(result) > 5 {
					t.Errorf("Expected array length 2-5, got: %d", len(result))
				}
				for i, item := range result {
					if _, ok := item.(string); !ok {
						t.Errorf("Expected string at index %d, got: %T", i, item)
					}
				}
			},
		},
		{
			name: "array with min/max items",
			schema: &openapi3.Schema{
				Type:     &openapi3.Types{"array"},
				MinItems: 3,
				MaxItems: uint64Ptr(4),
				Items: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"integer"},
					},
				},
			},
			check: func(t *testing.T, result []interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if len(result) < 3 || len(result) > 4 {
					t.Errorf("Expected array length 3-4, got: %d", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.generateArray(tt.schema)
			tt.check(t, result, err)
		})
	}
}

func TestGenerateObject(t *testing.T) {
	gen := NewGenerator(42)

	tests := []struct {
		name   string
		schema *openapi3.Schema
		check  func(t *testing.T, result map[string]interface{}, err error)
	}{
		{
			name: "empty object",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
			},
			check: func(t *testing.T, result map[string]interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected non-nil object")
				}
			},
		},
		{
			name: "object with properties",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"name": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
					"age": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"integer"},
						},
					},
				},
			},
			check: func(t *testing.T, result map[string]interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if _, ok := result["name"]; !ok {
					t.Error("Expected 'name' property in object")
				}
				if _, ok := result["age"]; !ok {
					t.Error("Expected 'age' property in object")
				}
				if _, ok := result["name"].(string); !ok {
					t.Errorf("Expected 'name' to be string, got: %T", result["name"])
				}
				if _, ok := result["age"].(int64); !ok {
					t.Errorf("Expected 'age' to be int64, got: %T", result["age"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.generateObject(tt.schema)
			tt.check(t, result, err)
		})
	}
}

func TestGenerateFromSchema(t *testing.T) {
	gen := NewGenerator(42)

	tests := []struct {
		name   string
		schema *openapi3.Schema
		check  func(t *testing.T, result interface{}, err error)
	}{
		{
			name:   "nil schema",
			schema: nil,
			check: func(t *testing.T, result interface{}, err error) {
				if err == nil {
					t.Error("Expected error for nil schema")
				}
			},
		},
		{
			name: "string type",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"string"},
			},
			check: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if _, ok := result.(string); !ok {
					t.Errorf("Expected string, got: %T", result)
				}
			},
		},
		{
			name: "integer type",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"integer"},
			},
			check: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if _, ok := result.(int64); !ok {
					t.Errorf("Expected int64, got: %T", result)
				}
			},
		},
		{
			name: "boolean type",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"boolean"},
			},
			check: func(t *testing.T, result interface{}, err error) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if _, ok := result.(bool); !ok {
					t.Errorf("Expected bool, got: %T", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.GenerateFromSchema(tt.schema)
			tt.check(t, result, err)
		})
	}
}

func TestDeterministicGeneration(t *testing.T) {
	schema := &openapi3.Schema{
		Type: &openapi3.Types{"object"},
		Properties: openapi3.Schemas{
			"id": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type:   &openapi3.Types{"string"},
					Format: "uuid",
				},
			},
			"name": &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
		},
	}

	// Generate with same seed twice
	gen1 := NewGenerator(100)
	result1, err1 := gen1.GenerateFromSchema(schema)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	gen2 := NewGenerator(100)
	result2, err2 := gen2.GenerateFromSchema(schema)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	// Check that results are identical
	obj1, ok1 := result1.(map[string]interface{})
	obj2, ok2 := result2.(map[string]interface{})

	if !ok1 || !ok2 {
		t.Fatal("Expected both results to be objects")
	}

	if obj1["id"] != obj2["id"] {
		t.Errorf("Expected deterministic UUID generation, got %v and %v", obj1["id"], obj2["id"])
	}

	if obj1["name"] != obj2["name"] {
		t.Errorf("Expected deterministic name generation, got %v and %v", obj1["name"], obj2["name"])
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func float64Ptr(f float64) *float64 {
	return &f
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}
