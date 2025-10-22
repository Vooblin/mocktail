package main

import (
	"strings"
	"testing"
)

func TestMockCommand(t *testing.T) {
	cmd := newMockCmd()

	// Check command properties
	if cmd.Use != "mock <schema-file>" {
		t.Errorf("Expected Use 'mock <schema-file>', got '%s'", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "mock API server") {
		t.Errorf("Expected Short to mention 'mock API server', got '%s'", cmd.Short)
	}

	if !strings.Contains(cmd.Long, "OpenAPI") {
		t.Error("Expected Long description to mention 'OpenAPI'")
	}

	// Check flags
	portFlag := cmd.Flags().Lookup("port")
	if portFlag == nil {
		t.Fatal("Expected 'port' flag to exist")
	}
	if portFlag.Shorthand != "p" {
		t.Errorf("Expected port flag shorthand 'p', got '%s'", portFlag.Shorthand)
	}
	if portFlag.DefValue != "8080" {
		t.Errorf("Expected default port '8080', got '%s'", portFlag.DefValue)
	}
}

func TestMockCommandRequiresArg(t *testing.T) {
	cmd := newMockCmd()

	// Set up command with empty args
	cmd.SetArgs([]string{})

	// Execute command - cobra should validate args
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error when no arguments provided, got nil")
	}

	// Check error message mentions missing argument
	if !strings.Contains(err.Error(), "arg") && !strings.Contains(err.Error(), "require") {
		t.Errorf("Expected error about missing argument, got: %v", err)
	}
}
