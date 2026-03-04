package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	integType   string
	integName   string
	integConfig string
)

func init() {
	rootCmd.AddCommand(integrationCmd)
	integrationCmd.AddCommand(integrationCreateCmd)
	integrationCmd.AddCommand(integrationListCmd)
	integrationCmd.AddCommand(integrationGetCmd)
	integrationCmd.AddCommand(integrationUpdateCmd)
	integrationCmd.AddCommand(integrationDeleteCmd)

	// create flags
	integrationCreateCmd.Flags().StringVar(&integType, "type", "", "Integration type (e.g. vault, keystone)")
	integrationCreateCmd.Flags().StringVar(&integName, "name", "", "Integration name")
	integrationCreateCmd.Flags().StringVar(&integConfig, "config", "{}", "Integration config as JSON string")
	integrationCreateCmd.MarkFlagRequired("type")
	integrationCreateCmd.MarkFlagRequired("name")

	// update flags
	integrationUpdateCmd.Flags().StringVar(&integName, "name", "", "Integration name")
	integrationUpdateCmd.Flags().StringVar(&integConfig, "config", "", "Integration config as JSON string")
}

var integrationCmd = &cobra.Command{
	Use:   "integration",
	Short: "Manage agent integrations",
}

var integrationCreateCmd = &cobra.Command{
	Use:   "create <agent-id>",
	Short: "Create an integration for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{
			"type": integType,
			"name": integName,
		}
		if integConfig != "" && integConfig != "{}" {
			var cfg map[string]any
			if err := json.Unmarshal([]byte(integConfig), &cfg); err != nil {
				return fmt.Errorf("invalid --config JSON: %w", err)
			}
			body["config"] = cfg
		}
		data, _, err := getClient().Post("/api/v1/agents/"+args[0]+"/integrations", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var integrationListCmd = &cobra.Command{
	Use:   "list <agent-id>",
	Short: "List integrations for an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents/" + args[0] + "/integrations")
		if err != nil {
			return err
		}
		var integrations []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &integrations); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTYPE\tNAME")
		for _, i := range integrations {
			fmt.Fprintf(w, "%s\t%s\t%s\n", i.ID, i.Type, i.Name)
		}
		return w.Flush()
	},
}

var integrationGetCmd = &cobra.Command{
	Use:   "get <agent-id> <integration-id>",
	Short: "Get an integration",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents/" + args[0] + "/integrations/" + args[1])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var integrationUpdateCmd = &cobra.Command{
	Use:   "update <agent-id> <integration-id>",
	Short: "Update an integration",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if integName != "" {
			body["name"] = integName
		}
		if integConfig != "" {
			var cfg map[string]any
			if err := json.Unmarshal([]byte(integConfig), &cfg); err != nil {
				return fmt.Errorf("invalid --config JSON: %w", err)
			}
			body["config"] = cfg
		}
		data, _, err := getClient().Put("/api/v1/agents/"+args[0]+"/integrations/"+args[1], body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var integrationDeleteCmd = &cobra.Command{
	Use:   "delete <agent-id> <integration-id>",
	Short: "Delete an integration",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/agents/" + args[0] + "/integrations/" + args[1])
		if err != nil {
			return err
		}
		fmt.Printf("Integration %s deleted\n", args[1])
		return nil
	},
}
