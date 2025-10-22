package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenAPIParser_Parse(t *testing.T) {
	// Create a temporary test OpenAPI file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-api.yaml")

	spec := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: List users
      description: Returns a list of users
      parameters:
        - name: limit
          in: query
          required: false
          schema:
            type: integer
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
  /users/{id}:
    get:
      summary: Get user by ID
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
`

	if err := os.WriteFile(testFile, []byte(spec), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewOpenAPIParser()
	schema, err := parser.Parse(testFile)

	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if schema.Type != "openapi" {
		t.Errorf("Expected Type 'openapi', got '%s'", schema.Type)
	}

	if schema.Title != "Test API" {
		t.Errorf("Expected Title 'Test API', got '%s'", schema.Title)
	}

	if schema.Version != "3.0.0" {
		t.Errorf("Expected Version '3.0.0', got '%s'", schema.Version)
	}

	if len(schema.Paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(schema.Paths))
	}

	// Check /users endpoint
	usersEndpoints, ok := schema.Paths["/users"]
	if !ok {
		t.Fatalf("Expected /users path to exist")
	}

	if len(usersEndpoints) != 1 {
		t.Fatalf("Expected 1 endpoint for /users, got %d", len(usersEndpoints))
	}

	if usersEndpoints[0].Method != "GET" {
		t.Errorf("Expected method GET, got %s", usersEndpoints[0].Method)
	}

	if usersEndpoints[0].Summary != "List users" {
		t.Errorf("Expected summary 'List users', got '%s'", usersEndpoints[0].Summary)
	}

	// Check parameters
	if len(usersEndpoints[0].Parameters) != 1 {
		t.Fatalf("Expected 1 parameter, got %d", len(usersEndpoints[0].Parameters))
	}

	limitParam := usersEndpoints[0].Parameters[0]
	if limitParam.Name != "limit" {
		t.Errorf("Expected parameter name 'limit', got '%s'", limitParam.Name)
	}

	if limitParam.In != "query" {
		t.Errorf("Expected parameter in 'query', got '%s'", limitParam.In)
	}

	if limitParam.Required {
		t.Errorf("Expected parameter required to be false")
	}

	if limitParam.Type != "integer" {
		t.Errorf("Expected parameter type 'integer', got '%s'", limitParam.Type)
	}
}

func TestOpenAPIParser_ParseInvalidFile(t *testing.T) {
	parser := NewOpenAPIParser()
	_, err := parser.Parse("/nonexistent/file.yaml")

	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestOpenAPIParser_ParseInvalidSpec(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidSpec := `not: valid
yaml: but
missing: openapi
`

	if err := os.WriteFile(testFile, []byte(invalidSpec), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewOpenAPIParser()
	_, err := parser.Parse(testFile)

	if err == nil {
		t.Error("Expected error for invalid OpenAPI spec, got nil")
	}
}
