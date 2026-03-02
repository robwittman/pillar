package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var serverURL string

var rootCmd = &cobra.Command{
	Use:   "pillarctl",
	Short: "CLI for the Pillar agent management system",
}

func init() {
	defaultServer := os.Getenv("PILLAR_SERVER")
	if defaultServer == "" {
		defaultServer = "http://localhost:8080"
	}
	rootCmd.PersistentFlags().StringVarP(&serverURL, "server", "s", defaultServer, "Pillar server URL (env: PILLAR_SERVER)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getClient() *APIClient {
	return NewAPIClient(serverURL)
}
