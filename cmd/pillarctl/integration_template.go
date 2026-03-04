package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	tmplType     string
	tmplName     string
	tmplConfig   string
	tmplSelector []string
)

func init() {
	rootCmd.AddCommand(templateCmd)
	templateCmd.AddCommand(templateCreateCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateGetCmd)
	templateCmd.AddCommand(templateUpdateCmd)
	templateCmd.AddCommand(templateDeleteCmd)
	templateCmd.AddCommand(templatePreviewCmd)

	// create flags
	templateCreateCmd.Flags().StringVar(&tmplType, "type", "", "Integration type (e.g. vault, keycloak)")
	templateCreateCmd.Flags().StringVar(&tmplName, "name", "", "Integration name")
	templateCreateCmd.Flags().StringVar(&tmplConfig, "config", "{}", "Integration config as JSON string")
	templateCreateCmd.Flags().StringSliceVar(&tmplSelector, "selector", nil, "Label selector key=value pairs")
	templateCreateCmd.MarkFlagRequired("type")
	templateCreateCmd.MarkFlagRequired("name")

	// update flags
	templateUpdateCmd.Flags().StringVar(&tmplName, "name", "", "Integration name")
	templateUpdateCmd.Flags().StringVar(&tmplConfig, "config", "", "Integration config as JSON string")
	templateUpdateCmd.Flags().StringSliceVar(&tmplSelector, "selector", nil, "Label selector key=value pairs")
}

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage integration templates",
}

var templateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an integration template",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{
			"type": tmplType,
			"name": tmplName,
		}
		if tmplConfig != "" && tmplConfig != "{}" {
			var cfg map[string]any
			if err := json.Unmarshal([]byte(tmplConfig), &cfg); err != nil {
				return fmt.Errorf("invalid --config JSON: %w", err)
			}
			body["config"] = cfg
		}
		if sel := parseKV(tmplSelector); len(sel) > 0 {
			body["selector"] = sel
		}
		data, _, err := getClient().Post("/api/v1/integration-templates", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List integration templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/integration-templates")
		if err != nil {
			return err
		}
		var templates []struct {
			ID   string `json:"id"`
			Type string `json:"type"`
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &templates); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTYPE\tNAME")
		for _, t := range templates {
			fmt.Fprintf(w, "%s\t%s\t%s\n", t.ID, t.Type, t.Name)
		}
		return w.Flush()
	},
}

var templateGetCmd = &cobra.Command{
	Use:   "get <template-id>",
	Short: "Get an integration template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/integration-templates/" + args[0])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var templateUpdateCmd = &cobra.Command{
	Use:   "update <template-id>",
	Short: "Update an integration template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if tmplName != "" {
			body["name"] = tmplName
		}
		if tmplConfig != "" {
			var cfg map[string]any
			if err := json.Unmarshal([]byte(tmplConfig), &cfg); err != nil {
				return fmt.Errorf("invalid --config JSON: %w", err)
			}
			body["config"] = cfg
		}
		if sel := parseKV(tmplSelector); len(sel) > 0 {
			body["selector"] = sel
		}
		data, _, err := getClient().Put("/api/v1/integration-templates/"+args[0], body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <template-id>",
	Short: "Delete an integration template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/integration-templates/" + args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Integration template %s deleted\n", args[0])
		return nil
	},
}

var templatePreviewCmd = &cobra.Command{
	Use:   "preview <template-id>",
	Short: "Preview agents matching a template's selector",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/integration-templates/" + args[0] + "/preview")
		if err != nil {
			return err
		}
		var agents []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Status string `json:"status"`
		}
		if err := json.Unmarshal(data, &agents); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS")
		for _, a := range agents {
			fmt.Fprintf(w, "%s\t%s\t%s\n", a.ID, a.Name, a.Status)
		}
		return w.Flush()
	},
}
