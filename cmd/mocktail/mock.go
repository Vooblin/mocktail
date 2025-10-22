package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vooblin/mocktail/internal/mock"
	"github.com/Vooblin/mocktail/internal/parser"
	"github.com/spf13/cobra"
)

func newMockCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "mock <schema-file>",
		Short: "Start a mock API server from a schema",
		Long: `Start a mock API server that serves responses based on an OpenAPI or GraphQL schema.

The server will parse the schema and automatically create endpoints with realistic mock responses.
Press Ctrl+C to stop the server.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schemaFile := args[0]

			// Validate file exists
			if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
				return fmt.Errorf("schema file not found: %s", schemaFile)
			}

			// Parse the schema
			fmt.Printf("📖 Parsing schema: %s\n", schemaFile)
			p := parser.NewOpenAPIParser()
			schema, err := p.Parse(schemaFile)
			if err != nil {
				return fmt.Errorf("failed to parse schema: %w", err)
			}

			// Create and start the mock server
			server := mock.NewServer(schema, port)

			// Handle graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			errChan := make(chan error, 1)
			go func() {
				errChan <- server.Start()
			}()

			// Wait for interrupt or error
			select {
			case sig := <-sigChan:
				log.Printf("\n📦 Received signal: %v", sig)
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				return server.Stop(ctx)
			case err := <-errChan:
				return err
			}
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the mock server on")

	return cmd
}
