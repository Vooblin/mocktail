package main

import (
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cmd := newParseCmd()

	if cmd.Use != "parse <schema-file>" {
		t.Errorf("Expected Use 'parse <schema-file>', got '%s'", cmd.Use)
	}

	if !strings.Contains(cmd.Short, "Parse and validate") {
		t.Errorf("Expected Short to contain 'Parse and validate', got '%s'", cmd.Short)
	}

	// Check that output flag exists
	outputFlag := cmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("Expected 'output' flag to exist")
	}

	if outputFlag.Shorthand != "o" {
		t.Errorf("Expected shorthand 'o', got '%s'", outputFlag.Shorthand)
	}
}
