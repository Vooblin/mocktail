package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Vooblin/mocktail/internal/parser"
)

func TestNewServer(t *testing.T) {
	schema := &parser.Schema{
		Type:    "openapi",
		Version: "3.0.0",
		Title:   "Test API",
		Paths:   make(map[string][]parser.Endpoint),
	}

	server := NewServer(schema, 8080)

	if server == nil {
		t.Fatal("Expected server to be created")
	}
	if server.schema != schema {
		t.Error("Expected schema to be set")
	}
	if server.port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.port)
	}
}

func TestServerStartAndStop(t *testing.T) {
	schema := &parser.Schema{
		Type:    "openapi",
		Version: "3.0.0",
		Title:   "Test API",
		Paths: map[string][]parser.Endpoint{
			"/test": {
				{Method: "GET", Path: "/test", Summary: "Test endpoint"},
			},
		},
	}

	server := NewServer(schema, 8081)

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test health check
	resp, err := http.Get("http://localhost:8081/health")
	if err != nil {
		t.Fatalf("Failed to reach server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", health["status"])
	}
	if health["server"] != "mocktail" {
		t.Errorf("Expected server 'mocktail', got '%s'", health["server"])
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Verify server stopped (give it a moment)
	time.Sleep(100 * time.Millisecond)
	_, err = http.Get("http://localhost:8081/health")
	if err == nil {
		t.Error("Expected server to be stopped, but it's still reachable")
	}
}

func TestServerEndpoints(t *testing.T) {
	tests := []struct {
		name           string
		endpoint       parser.Endpoint
		method         string
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "GET list endpoint",
			endpoint: parser.Endpoint{
				Method:  "GET",
				Path:    "/items",
				Summary: "List items",
			},
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if _, ok := response["data"]; !ok {
					t.Error("Expected 'data' field in response")
				}
				if _, ok := response["total"]; !ok {
					t.Error("Expected 'total' field in response")
				}
			},
		},
		{
			name: "GET single resource",
			endpoint: parser.Endpoint{
				Method:  "GET",
				Path:    "/items/{id}",
				Summary: "Get item by ID",
			},
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if _, ok := response["id"]; !ok {
					t.Error("Expected 'id' field in response")
				}
				if _, ok := response["name"]; !ok {
					t.Error("Expected 'name' field in response")
				}
			},
		},
		{
			name: "POST create resource",
			endpoint: parser.Endpoint{
				Method:  "POST",
				Path:    "/items",
				Summary: "Create item",
			},
			method:         "POST",
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if _, ok := response["id"]; !ok {
					t.Error("Expected 'id' field in response")
				}
				if _, ok := response["message"]; !ok {
					t.Error("Expected 'message' field in response")
				}
			},
		},
		{
			name: "PUT update resource",
			endpoint: parser.Endpoint{
				Method:  "PUT",
				Path:    "/items/{id}",
				Summary: "Update item",
			},
			method:         "PUT",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if _, ok := response["updatedAt"]; !ok {
					t.Error("Expected 'updatedAt' field in response")
				}
			},
		},
		{
			name: "DELETE resource",
			endpoint: parser.Endpoint{
				Method:  "DELETE",
				Path:    "/items/{id}",
				Summary: "Delete item",
			},
			method:         "DELETE",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				if err := json.Unmarshal(body, &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if msg, ok := response["message"]; !ok {
					t.Error("Expected 'message' field in response")
				} else if !contains(msg.(string), "deleted") {
					t.Error("Expected message to contain 'deleted'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create schema with single endpoint
			schema := &parser.Schema{
				Type:    "openapi",
				Version: "3.0.0",
				Title:   "Test API",
				Paths: map[string][]parser.Endpoint{
					tt.endpoint.Path: {tt.endpoint},
				},
			}

			// Use unique port for each test
			port := 8090 + len(tt.name)%10
			server := NewServer(schema, port)

			// Start server
			go server.Start()
			time.Sleep(100 * time.Millisecond)
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				server.Stop(ctx)
			}()

			// Make request
			url := fmt.Sprintf("http://localhost:%d%s", port, tt.endpoint.Path)
			req, err := http.NewRequest(tt.method, url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			client := &http.Client{Timeout: 2 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check headers
			if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}
			if resp.Header.Get("X-Mocktail-Server") != "true" {
				t.Error("Expected X-Mocktail-Server header to be 'true'")
			}

			// Read and check response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, body)
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	schema := &parser.Schema{
		Type:    "openapi",
		Version: "3.0.0",
		Title:   "Test API",
		Paths: map[string][]parser.Endpoint{
			"/test": {
				{Method: "GET", Path: "/test", Summary: "Test endpoint"},
			},
		},
	}

	server := NewServer(schema, 8092)
	go server.Start()
	time.Sleep(100 * time.Millisecond)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Try POST on a GET-only endpoint
	resp, err := http.Post("http://localhost:8092/test", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}()))
}
