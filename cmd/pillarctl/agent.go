package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(agentCmd)
	agentCmd.AddCommand(agentCreateCmd)
	agentCmd.AddCommand(agentListCmd)
	agentCmd.AddCommand(agentGetCmd)
	agentCmd.AddCommand(agentUpdateCmd)
	agentCmd.AddCommand(agentDeleteCmd)
	agentCmd.AddCommand(agentStartCmd)
	agentCmd.AddCommand(agentStopCmd)
	agentCmd.AddCommand(agentStatusCmd)

	// create flags
	agentCreateCmd.Flags().StringVar(&agentName, "name", "", "Agent name (required)")
	agentCreateCmd.Flags().StringSliceVar(&agentMetadata, "metadata", nil, "Metadata key=value pairs")
	agentCreateCmd.Flags().StringSliceVar(&agentLabels, "labels", nil, "Label key=value pairs")
	agentCreateCmd.MarkFlagRequired("name")

	// update flags
	agentUpdateCmd.Flags().StringVar(&agentName, "name", "", "Agent name")
	agentUpdateCmd.Flags().StringSliceVar(&agentMetadata, "metadata", nil, "Metadata key=value pairs")
	agentUpdateCmd.Flags().StringSliceVar(&agentLabels, "labels", nil, "Label key=value pairs")
}

var (
	agentName     string
	agentMetadata []string
	agentLabels   []string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents",
}

var agentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{
			"name": agentName,
		}
		if md := parseKV(agentMetadata); len(md) > 0 {
			body["metadata"] = md
		}
		if lb := parseKV(agentLabels); len(lb) > 0 {
			body["labels"] = lb
		}

		data, _, err := getClient().Post("/api/v1/agents", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents")
		if err != nil {
			return err
		}

		var agents []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &agents); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS")
		for _, a := range agents {
			fmt.Fprintf(w, "%s\t%s\t%s\n", a.ID, a.Name, a.Status)
		}
		return w.Flush()
	},
}

var agentGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get agent details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents/" + args[0])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var agentUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if agentName != "" {
			body["name"] = agentName
		}
		if md := parseKV(agentMetadata); len(md) > 0 {
			body["metadata"] = md
		}
		if lb := parseKV(agentLabels); len(lb) > 0 {
			body["labels"] = lb
		}

		data, _, err := getClient().Put("/api/v1/agents/"+args[0], body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var agentDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/agents/" + args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Agent %s deleted\n", args[0])
		return nil
	},
}

var agentStartCmd = &cobra.Command{
	Use:   "start <id>",
	Short: "Start an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, err := getClient().Post("/api/v1/agents/"+args[0]+"/start", nil)
		if err != nil {
			return err
		}
		fmt.Printf("Agent %s started\n", args[0])
		return nil
	},
}

var agentStopCmd = &cobra.Command{
	Use:   "stop <id>",
	Short: "Stop an agent",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, _, err := getClient().Post("/api/v1/agents/"+args[0]+"/stop", nil)
		if err != nil {
			return err
		}
		fmt.Printf("Agent %s stopped\n", args[0])
		return nil
	},
}

var agentStatusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Get agent status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents/" + args[0] + "/status")
		if err != nil {
			return err
		}

		var status struct {
			AgentID string `json:"agent_id"`
			Status  string `json:"status"`
			Online  bool   `json:"online"`
		}
		if err := json.Unmarshal(data, &status); err != nil {
			return fmt.Errorf("parse response: %w", err)
		}

		fmt.Printf("%s  status=%s  online=%t\n", status.AgentID, status.Status, status.Online)
		return nil
	},
}

func parseKV(pairs []string) map[string]string {
	if len(pairs) == 0 {
		return nil
	}
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if ok {
			m[k] = v
		}
	}
	return m
}

func printJSON(data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		_, err = os.Stdout.Write(data)
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
