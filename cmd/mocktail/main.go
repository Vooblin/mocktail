package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mocktail",
		Short: "API mocking and contract testing tool",
		Long: `Mocktail is an API mocking and contract testing tool for small teams and indie developers.

Upload an OpenAPI/GraphQL schema, or point it at a staging endpoint, and Mocktail spins up 
a realistic mock server, generates sample and edge-case payloads, and auto-writes contract 
tests for your CI. It then watches traffic to detect breaking changes before they reach production.`,
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add subcommands as they are developed
	rootCmd.AddCommand(newParseCmd())
	rootCmd.AddCommand(newMockCmd())
	// rootCmd.AddCommand(newGenerateCmd())
	// rootCmd.AddCommand(newMonitorCmd())

	return rootCmd
}
