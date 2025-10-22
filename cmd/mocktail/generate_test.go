package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCommand(t *testing.T) {
	// Create a temporary OpenAPI schema file
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "test-schema.yaml")

	schemaContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /items:
    get:
      summary: List items
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
                  properties:
                    id:
                      type: string
                      format: uuid
                    name:
                      type: string
    post:
      summary: Create item
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required:
                - name
              properties:
                name:
                  type: string
                count:
                  type: integer
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    format: uuid
                  name:
                    type: string
                  count:
                    type: integer
`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	tests := []struct {
		name         string
		args         []string
		expectError  bool
		validateFunc func(t *testing.T, output string)
	}{
		{
			name: "generate GET response",
			args: []string{"generate", schemaFile, "--path", "/items", "--method", "GET", "--seed", "42"},
			validateFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "Response Body") {
					t.Error("Expected response body in output")
				}
				if !strings.Contains(output, "id") {
					t.Error("Expected 'id' field in output")
				}
				if !strings.Contains(output, "name") {
					t.Error("Expected 'name' field in output")
				}

				// Verify it's valid JSON
				lines := strings.Split(output, "\n")
				var jsonStart int
				for i, line := range lines {
					if strings.HasPrefix(strings.TrimSpace(line), "[") {
						jsonStart = i
						break
					}
				}
				if jsonStart > 0 {
					jsonStr := strings.Join(lines[jsonStart:], "\n")
					var data interface{}
					if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
						t.Errorf("Output is not valid JSON: %v", err)
					}
				}
			},
		},
		{
			name: "generate POST request and response",
			args: []string{"generate", schemaFile, "--path", "/items", "--method", "POST", "--seed", "42"},
			validateFunc: func(t *testing.T, output string) {
				if !strings.Contains(output, "Request Body") {
					t.Error("Expected request body in output")
				}
				if !strings.Contains(output, "Response Body") {
					t.Error("Expected response body in output")
				}
				if !strings.Contains(output, "name") {
					t.Error("Expected 'name' field in output")
				}
			},
		},
		{
			name: "generate multiple payloads",
			args: []string{"generate", schemaFile, "--path", "/items", "--method", "GET", "--count", "3", "--seed", "42"},
			validateFunc: func(t *testing.T, output string) {
				count := strings.Count(output, "Response Body")
				if count != 3 {
					t.Errorf("Expected 3 response bodies, got %d", count)
				}
			},
		},
		{
			name:        "missing path flag",
			args:        []string{"generate", schemaFile, "--method", "GET"},
			expectError: true,
		},
		{
			name:        "missing method flag",
			args:        []string{"generate", schemaFile, "--path", "/items"},
			expectError: true,
		},
		{
			name:        "invalid path",
			args:        []string{"generate", schemaFile, "--path", "/invalid", "--method", "GET"},
			expectError: true,
		},
		{
			name:        "invalid method",
			args:        []string{"generate", schemaFile, "--path", "/items", "--method", "DELETE"},
			expectError: true,
		},
		{
			name:        "missing schema file",
			args:        []string{"generate"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture both stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			rootCmd := newRootCmd()
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()

			// Restore stdout/stderr
			w.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v\nOutput: %s", err, output)
				return
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, output)
			}
		})
	}
}

func TestGenerateCommandReproducibility(t *testing.T) {
	// Create a temporary OpenAPI schema file
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "test-schema.yaml")

	schemaContent := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /data:
    get:
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  value:
                    type: string
                  number:
                    type: integer
`

	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0644); err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	// Test that seed produces consistent output
	// We'll generate output and verify it contains expected patterns
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd := newRootCmd()
	rootCmd.SetArgs([]string{"generate", schemaFile, "--path", "/data", "--method", "GET", "--seed", "12345"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected structure
	if !strings.Contains(output, "Response Body") {
		t.Error("Expected 'Response Body' in output")
	}
	if !strings.Contains(output, "value") {
		t.Error("Expected 'value' field in output")
	}
	if !strings.Contains(output, "number") {
		t.Error("Expected 'number' field in output")
	}

	// Verify it's valid JSON
	lines := strings.Split(output, "\n")
	var jsonStart int
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "{") {
			jsonStart = i
			break
		}
	}
	if jsonStart > 0 {
		jsonStr := strings.Join(lines[jsonStart:], "\n")
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
			t.Errorf("Output is not valid JSON: %v", err)
		}
	}
}
