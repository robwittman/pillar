package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	webhookURL         string
	webhookDescription string
	webhookEventTypes  []string
	webhookStatus      string
)

func init() {
	rootCmd.AddCommand(webhookCmd)
	webhookCmd.AddCommand(webhookCreateCmd)
	webhookCmd.AddCommand(webhookListCmd)
	webhookCmd.AddCommand(webhookGetCmd)
	webhookCmd.AddCommand(webhookUpdateCmd)
	webhookCmd.AddCommand(webhookDeleteCmd)
	webhookCmd.AddCommand(webhookRotateSecretCmd)
	webhookCmd.AddCommand(webhookDeliveriesCmd)

	webhookCreateCmd.Flags().StringVar(&webhookURL, "url", "", "Webhook endpoint URL")
	webhookCreateCmd.Flags().StringVar(&webhookDescription, "description", "", "Webhook description")
	webhookCreateCmd.Flags().StringSliceVar(&webhookEventTypes, "event-types", nil, "Event types to subscribe to")
	webhookCreateCmd.MarkFlagRequired("url")

	webhookUpdateCmd.Flags().StringVar(&webhookDescription, "description", "", "Webhook description")
	webhookUpdateCmd.Flags().StringSliceVar(&webhookEventTypes, "event-types", nil, "Event types to subscribe to")
	webhookUpdateCmd.Flags().StringVar(&webhookStatus, "status", "", "Webhook status (active/inactive)")
}

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage webhooks",
}

var webhookCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{
			"url": webhookURL,
		}
		if webhookDescription != "" {
			body["description"] = webhookDescription
		}
		if len(webhookEventTypes) > 0 {
			body["event_types"] = webhookEventTypes
		}
		data, _, err := getClient().Post("/api/v1/webhooks", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var webhookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/webhooks")
		if err != nil {
			return err
		}
		var webhooks []struct {
			ID          string `json:"id"`
			URL         string `json:"url"`
			Status      string `json:"status"`
			Description string `json:"description"`
		}
		if err := json.Unmarshal(data, &webhooks); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tURL\tSTATUS\tDESCRIPTION")
		for _, wh := range webhooks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", wh.ID, wh.URL, wh.Status, wh.Description)
		}
		return w.Flush()
	},
}

var webhookGetCmd = &cobra.Command{
	Use:   "get <webhook-id>",
	Short: "Get a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/webhooks/" + args[0])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var webhookUpdateCmd = &cobra.Command{
	Use:   "update <webhook-id>",
	Short: "Update a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if webhookDescription != "" {
			body["description"] = webhookDescription
		}
		if len(webhookEventTypes) > 0 {
			body["event_types"] = webhookEventTypes
		}
		if webhookStatus != "" {
			body["status"] = webhookStatus
		}
		data, _, err := getClient().Put("/api/v1/webhooks/"+args[0], body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var webhookDeleteCmd = &cobra.Command{
	Use:   "delete <webhook-id>",
	Short: "Delete a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/webhooks/" + args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Webhook %s deleted\n", args[0])
		return nil
	},
}

var webhookRotateSecretCmd = &cobra.Command{
	Use:   "rotate-secret <webhook-id>",
	Short: "Rotate a webhook's signing secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Post("/api/v1/webhooks/"+args[0]+"/rotate-secret", nil)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var webhookDeliveriesCmd = &cobra.Command{
	Use:   "deliveries <webhook-id>",
	Short: "List deliveries for a webhook",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/webhooks/" + args[0] + "/deliveries")
		if err != nil {
			return err
		}
		var deliveries []struct {
			ID        string `json:"id"`
			EventType string `json:"event_type"`
			Status    string `json:"status"`
			Attempts  int    `json:"attempts"`
		}
		if err := json.Unmarshal(data, &deliveries); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tEVENT_TYPE\tSTATUS\tATTEMPTS")
		for _, d := range deliveries {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", d.ID, d.EventType, d.Status, d.Attempts)
		}
		return w.Flush()
	},
}
