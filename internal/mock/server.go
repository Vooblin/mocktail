package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Vooblin/mocktail/internal/generator"
	"github.com/Vooblin/mocktail/internal/parser"
	"github.com/getkin/kin-openapi/openapi3"
)

// Server represents a mock API server
type Server struct {
	schema    *parser.Schema
	server    *http.Server
	port      int
	generator *generator.Generator
}

// NewServer creates a new mock server from a parsed schema
func NewServer(schema *parser.Schema, port int) *Server {
	return &Server{
		schema:    schema,
		port:      port,
		generator: generator.NewGenerator(time.Now().UnixNano()),
	}
}

// Start begins serving mock responses
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Register all endpoints from the schema - group by path
	for path, endpoints := range s.schema.Paths {
		// Create a closure to capture the endpoints for this path
		pathEndpoints := endpoints
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			s.handlePath(w, r, pathEndpoints)
		})
	}

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"server": "mocktail",
		})
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.loggingMiddleware(mux),
	}

	log.Printf("🍹 Mocktail server starting on http://localhost:%d", s.port)
	log.Printf("📋 Schema: %s (version %s)", s.schema.Title, s.schema.Version)
	log.Printf("🎯 Registered %d paths", len(s.schema.Paths))

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	log.Println("🛑 Shutting down mock server...")
	return s.server.Shutdown(ctx)
}

// handlePath handles all methods for a given path
func (s *Server) handlePath(w http.ResponseWriter, r *http.Request, endpoints []parser.Endpoint) {
	// Find the endpoint that matches the request method
	var matchedEndpoint *parser.Endpoint
	for i, endpoint := range endpoints {
		if strings.EqualFold(r.Method, endpoint.Method) {
			matchedEndpoint = &endpoints[i]
			break
		}
	}

	// If no matching method found, return 405
	if matchedEndpoint == nil {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate mock response based on the endpoint
	response := s.generateMockResponse(*matchedEndpoint, r)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Mocktail-Server", "true")

	// Set status code based on method
	statusCode := s.getStatusCode(matchedEndpoint.Method)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

// generateMockResponse creates a mock response for an endpoint
func (s *Server) generateMockResponse(endpoint parser.Endpoint, r *http.Request) interface{} {
	// Try to generate from OpenAPI schema first
	if doc, ok := s.schema.Raw.(*openapi3.T); ok {
		if pathItem := doc.Paths.Value(endpoint.Path); pathItem != nil {
			operation := pathItem.Operations()[endpoint.Method]
			if operation != nil {
				// Determine status code
				statusCode := s.getStatusCodeString(endpoint.Method)

				// Try to generate from schema
				if response, err := s.generator.GenerateResponse(operation, statusCode); err == nil {
					// For list endpoints, wrap in array structure
					if !strings.Contains(endpoint.Path, "{") && endpoint.Method == "GET" {
						if items, ok := response.(map[string]interface{}); ok {
							// If the response is a single object, make it an array
							return map[string]interface{}{
								"data":  []interface{}{items, items}, // Generate 2 items for lists
								"total": 2,
							}
						}
					}
					return response
				}
			}
		}
	}

	// Fallback to basic mock response structure
	response := make(map[string]interface{})
	switch endpoint.Method {
	case "GET":
		if strings.Contains(endpoint.Path, "{") {
			response["id"] = "550e8400-e29b-41d4-a716-446655440000"
			response["name"] = "Mock Resource"
			response["createdAt"] = time.Now().Format(time.RFC3339)
		} else {
			response["data"] = []map[string]interface{}{
				{
					"id":        "550e8400-e29b-41d4-a716-446655440000",
					"name":      "Mock Resource 1",
					"createdAt": time.Now().Format(time.RFC3339),
				},
				{
					"id":        "550e8400-e29b-41d4-a716-446655440001",
					"name":      "Mock Resource 2",
					"createdAt": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
				},
			}
			response["total"] = 2
		}
	case "POST":
		response["id"] = "550e8400-e29b-41d4-a716-446655440000"
		response["name"] = "New Mock Resource"
		response["createdAt"] = time.Now().Format(time.RFC3339)
		response["message"] = "Resource created successfully"
	case "PUT", "PATCH":
		response["id"] = "550e8400-e29b-41d4-a716-446655440000"
		response["name"] = "Updated Mock Resource"
		response["updatedAt"] = time.Now().Format(time.RFC3339)
		response["message"] = "Resource updated successfully"
	case "DELETE":
		response["message"] = "Resource deleted successfully"
	}

	return response
}

// getStatusCodeString returns the status code as a string for looking up responses
func (s *Server) getStatusCodeString(method string) string {
	switch method {
	case "POST":
		return "201"
	case "DELETE":
		return "204"
	default:
		return "200"
	}
}

// getStatusCode returns the appropriate status code for a method
func (s *Server) getStatusCode(method string) int {
	switch method {
	case "POST":
		return http.StatusCreated
	case "DELETE":
		return http.StatusOK
	default:
		return http.StatusOK
	}
}

// loggingMiddleware logs all incoming requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
