package main

import (
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	rootCmd := newRootCmd()

	if rootCmd.Use != "mocktail" {
		t.Errorf("Expected Use to be 'mocktail', got '%s'", rootCmd.Use)
	}

	if rootCmd.Version != version {
		t.Errorf("Expected Version to be '%s', got '%s'", version, rootCmd.Version)
	}

	if !strings.Contains(rootCmd.Long, "API mocking") {
		t.Error("Expected Long description to contain 'API mocking'")
	}
}

func TestVersionConstant(t *testing.T) {
	if version == "" {
		t.Error("Version should not be empty")
	}
}
