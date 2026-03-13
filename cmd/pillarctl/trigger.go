package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	triggerSourceID     string
	triggerAgentID      string
	triggerName         string
	triggerTaskTemplate string
	triggerFilterJSON   string
	triggerEnabled      string
)

func init() {
	rootCmd.AddCommand(triggerCmd)
	triggerCmd.AddCommand(triggerCreateCmd)
	triggerCmd.AddCommand(triggerListCmd)
	triggerCmd.AddCommand(triggerGetCmd)
	triggerCmd.AddCommand(triggerUpdateCmd)
	triggerCmd.AddCommand(triggerDeleteCmd)

	triggerCreateCmd.Flags().StringVar(&triggerSourceID, "source-id", "", "Source ID")
	triggerCreateCmd.Flags().StringVar(&triggerAgentID, "agent-id", "", "Agent ID")
	triggerCreateCmd.Flags().StringVar(&triggerName, "name", "", "Trigger name")
	triggerCreateCmd.Flags().StringVar(&triggerTaskTemplate, "task-template", "", "Go template for task prompt")
	triggerCreateCmd.Flags().StringVar(&triggerFilterJSON, "filter", "", `Filter conditions as JSON (e.g. '{"conditions":[{"path":"action","op":"eq","value":"opened"}]}')`)
	triggerCreateCmd.MarkFlagRequired("source-id")
	triggerCreateCmd.MarkFlagRequired("agent-id")
	triggerCreateCmd.MarkFlagRequired("name")

	triggerUpdateCmd.Flags().StringVar(&triggerName, "name", "", "Trigger name")
	triggerUpdateCmd.Flags().StringVar(&triggerTaskTemplate, "task-template", "", "Go template for task prompt")
	triggerUpdateCmd.Flags().StringVar(&triggerFilterJSON, "filter", "", "Filter conditions as JSON")
	triggerUpdateCmd.Flags().StringVar(&triggerEnabled, "enabled", "", "Enable/disable trigger (true/false)")

	triggerListCmd.Flags().StringVar(&triggerSourceID, "source-id", "", "Filter by source ID")
}

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Manage triggers",
}

var triggerCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a trigger",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{
			"source_id": triggerSourceID,
			"agent_id":  triggerAgentID,
			"name":      triggerName,
		}
		if triggerTaskTemplate != "" {
			body["task_template"] = triggerTaskTemplate
		}
		if triggerFilterJSON != "" {
			var filter json.RawMessage
			if err := json.Unmarshal([]byte(triggerFilterJSON), &filter); err != nil {
				return fmt.Errorf("invalid --filter JSON: %w", err)
			}
			body["filter"] = filter
		}
		data, _, err := getClient().Post("/api/v1/triggers", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var triggerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List triggers",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/api/v1/triggers"
		if triggerSourceID != "" {
			path += "?source_id=" + triggerSourceID
		}
		data, _, err := getClient().Get(path)
		if err != nil {
			return err
		}
		var triggers []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			SourceID string `json:"source_id"`
			AgentID  string `json:"agent_id"`
			Enabled  bool   `json:"enabled"`
		}
		if err := json.Unmarshal(data, &triggers); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSOURCE\tAGENT\tENABLED")
		for _, t := range triggers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%v\n", t.ID, t.Name, t.SourceID[:8], t.AgentID[:8], t.Enabled)
		}
		return w.Flush()
	},
}

var triggerGetCmd = &cobra.Command{
	Use:   "get <trigger-id>",
	Short: "Get a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/triggers/" + args[0])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var triggerUpdateCmd = &cobra.Command{
	Use:   "update <trigger-id>",
	Short: "Update a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if triggerName != "" {
			body["name"] = triggerName
		}
		if triggerTaskTemplate != "" {
			body["task_template"] = triggerTaskTemplate
		}
		if triggerFilterJSON != "" {
			var filter json.RawMessage
			if err := json.Unmarshal([]byte(triggerFilterJSON), &filter); err != nil {
				return fmt.Errorf("invalid --filter JSON: %w", err)
			}
			body["filter"] = filter
		}
		if triggerEnabled != "" {
			body["enabled"] = triggerEnabled == "true"
		}
		data, _, err := getClient().Put("/api/v1/triggers/"+args[0], body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var triggerDeleteCmd = &cobra.Command{
	Use:   "delete <trigger-id>",
	Short: "Delete a trigger",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/triggers/" + args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Trigger %s deleted\n", args[0])
		return nil
	},
}
