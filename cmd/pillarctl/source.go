package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var sourceName string

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.AddCommand(sourceCreateCmd)
	sourceCmd.AddCommand(sourceListCmd)
	sourceCmd.AddCommand(sourceGetCmd)
	sourceCmd.AddCommand(sourceUpdateCmd)
	sourceCmd.AddCommand(sourceDeleteCmd)
	sourceCmd.AddCommand(sourceRotateSecretCmd)

	sourceCreateCmd.Flags().StringVar(&sourceName, "name", "", "Source name")
	sourceCreateCmd.MarkFlagRequired("name")

	sourceUpdateCmd.Flags().StringVar(&sourceName, "name", "", "Source name")
}

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage event sources",
}

var sourceCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an event source",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Post("/api/v1/sources", map[string]any{
			"name": sourceName,
		})
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var sourceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List event sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/sources")
		if err != nil {
			return err
		}
		var sources []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			CreatedAt string `json:"created_at"`
		}
		if err := json.Unmarshal(data, &sources); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tCREATED")
		for _, s := range sources {
			fmt.Fprintf(w, "%s\t%s\t%s\n", s.ID, s.Name, s.CreatedAt)
		}
		return w.Flush()
	},
}

var sourceGetCmd = &cobra.Command{
	Use:   "get <source-id>",
	Short: "Get a source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/sources/" + args[0])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var sourceUpdateCmd = &cobra.Command{
	Use:   "update <source-id>",
	Short: "Update a source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{}
		if sourceName != "" {
			body["name"] = sourceName
		}
		data, _, err := getClient().Put("/api/v1/sources/"+args[0], body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var sourceDeleteCmd = &cobra.Command{
	Use:   "delete <source-id>",
	Short: "Delete a source",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/sources/" + args[0])
		if err != nil {
			return err
		}
		fmt.Printf("Source %s deleted\n", args[0])
		return nil
	},
}

var sourceRotateSecretCmd = &cobra.Command{
	Use:   "rotate-secret <source-id>",
	Short: "Rotate a source's signing secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Post("/api/v1/sources/"+args[0]+"/rotate-secret", nil)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}
