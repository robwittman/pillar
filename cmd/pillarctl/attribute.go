package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var attrValue string

func init() {
	rootCmd.AddCommand(attributeCmd)
	attributeCmd.AddCommand(attrSetCmd)
	attributeCmd.AddCommand(attrListCmd)
	attributeCmd.AddCommand(attrGetCmd)
	attributeCmd.AddCommand(attrDeleteCmd)

	attrSetCmd.Flags().StringVar(&attrValue, "value", "{}", "Attribute value as JSON string")
}

var attributeCmd = &cobra.Command{
	Use:   "attribute",
	Short: "Manage agent attributes",
}

var attrSetCmd = &cobra.Command{
	Use:   "set <agent-id> <namespace>",
	Short: "Set an agent attribute",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var value json.RawMessage
		if err := json.Unmarshal([]byte(attrValue), &value); err != nil {
			return fmt.Errorf("invalid --value JSON: %w", err)
		}
		data, _, err := getClient().Put("/api/v1/agents/"+args[0]+"/attributes/"+args[1], value)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var attrListCmd = &cobra.Command{
	Use:   "list <agent-id>",
	Short: "List agent attributes",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents/" + args[0] + "/attributes")
		if err != nil {
			return err
		}
		var attrs []struct {
			Namespace string `json:"namespace"`
			AgentID   string `json:"agent_id"`
		}
		if err := json.Unmarshal(data, &attrs); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAMESPACE\tAGENT_ID")
		for _, a := range attrs {
			fmt.Fprintf(w, "%s\t%s\n", a.Namespace, a.AgentID)
		}
		return w.Flush()
	},
}

var attrGetCmd = &cobra.Command{
	Use:   "get <agent-id> <namespace>",
	Short: "Get an agent attribute",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents/" + args[0] + "/attributes/" + args[1])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var attrDeleteCmd = &cobra.Command{
	Use:   "delete <agent-id> <namespace>",
	Short: "Delete an agent attribute",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/agents/" + args[0] + "/attributes/" + args[1])
		if err != nil {
			return err
		}
		fmt.Printf("Attribute %s deleted for agent %s\n", args[1], args[0])
		return nil
	},
}
