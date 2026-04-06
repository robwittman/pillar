package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(authCmd)

	// Token commands
	authCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(tokenCreateCmd)
	tokenCmd.AddCommand(tokenListCmd)
	tokenCmd.AddCommand(tokenRevokeCmd)
	tokenCreateCmd.Flags().StringVar(&tokenName, "name", "", "Token name (required)")
	tokenCreateCmd.MarkFlagRequired("name")

	// Service account commands
	authCmd.AddCommand(serviceAccountCmd)
	serviceAccountCmd.AddCommand(saCreateCmd)
	serviceAccountCmd.AddCommand(saListCmd)
	serviceAccountCmd.AddCommand(saDeleteCmd)
	serviceAccountCmd.AddCommand(saRotateCmd)
	saCreateCmd.Flags().StringVar(&saName, "name", "", "Service account name (required)")
	saCreateCmd.Flags().StringVar(&saDescription, "description", "", "Description")
	saCreateCmd.MarkFlagRequired("name")
}

var (
	tokenName     string
	saName        string
	saDescription string
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication (tokens, service accounts)",
}

// --- Token commands ---

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens",
}

var tokenCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API token",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		body, _, err := client.Post("/api/v1/auth/tokens", map[string]string{
			"name": tokenName,
		})
		if err != nil {
			return err
		}

		var resp struct {
			Token string `json:"token"`
			Meta  struct {
				ID        string `json:"id"`
				Name      string `json:"name"`
				CreatedAt string `json:"created_at"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}

		fmt.Printf("Token created: %s\n", resp.Meta.Name)
		fmt.Printf("ID:    %s\n", resp.Meta.ID)
		fmt.Printf("Token: %s\n", resp.Token)
		fmt.Println("\nSave this token — it will not be shown again.")
		return nil
	},
}

var tokenListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API tokens",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		body, _, err := client.Get("/api/v1/auth/tokens")
		if err != nil {
			return err
		}

		var tokens []struct {
			ID         string  `json:"id"`
			Name       string  `json:"name"`
			CreatedAt  string  `json:"created_at"`
			LastUsedAt *string `json:"last_used_at"`
			ExpiresAt  *string `json:"expires_at"`
		}
		if err := json.Unmarshal(body, &tokens); err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tCREATED\tLAST USED\tEXPIRES")
		for _, t := range tokens {
			lastUsed := "-"
			if t.LastUsedAt != nil {
				lastUsed = *t.LastUsedAt
			}
			expires := "never"
			if t.ExpiresAt != nil {
				expires = *t.ExpiresAt
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.ID, t.Name, t.CreatedAt, lastUsed, expires)
		}
		return w.Flush()
	},
}

var tokenRevokeCmd = &cobra.Command{
	Use:   "revoke [token-id]",
	Short: "Revoke an API token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		_, err := client.Delete("/api/v1/auth/tokens/" + args[0])
		if err != nil {
			return err
		}
		fmt.Println("Token revoked.")
		return nil
	},
}

// --- Service account commands ---

var serviceAccountCmd = &cobra.Command{
	Use:     "service-account",
	Aliases: []string{"sa"},
	Short:   "Manage service accounts",
}

var saCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new service account",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		body, _, err := client.Post("/api/v1/auth/service-accounts", map[string]string{
			"name":        saName,
			"description": saDescription,
		})
		if err != nil {
			return err
		}

		var resp struct {
			ServiceAccount struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"service_account"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}

		fmt.Printf("Service account created: %s\n", resp.ServiceAccount.Name)
		fmt.Printf("Client ID:     %s\n", resp.ClientID)
		fmt.Printf("Client Secret: %s\n", resp.ClientSecret)
		fmt.Println("\nSave these credentials — the secret will not be shown again.")
		return nil
	},
}

var saListCmd = &cobra.Command{
	Use:   "list",
	Short: "List service accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		body, _, err := client.Get("/api/v1/auth/service-accounts")
		if err != nil {
			return err
		}

		var accounts []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Disabled    bool   `json:"disabled"`
			CreatedAt   string `json:"created_at"`
		}
		if err := json.Unmarshal(body, &accounts); err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION\tDISABLED\tCREATED")
		for _, a := range accounts {
			fmt.Fprintf(w, "%s\t%s\t%s\t%v\t%s\n", a.ID, a.Name, a.Description, a.Disabled, a.CreatedAt)
		}
		return w.Flush()
	},
}

var saDeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a service account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		_, err := client.Delete("/api/v1/auth/service-accounts/" + args[0])
		if err != nil {
			return err
		}
		fmt.Println("Service account deleted.")
		return nil
	},
}

var saRotateCmd = &cobra.Command{
	Use:   "rotate-secret [id]",
	Short: "Rotate a service account's secret",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client := getClient()
		body, _, err := client.Post("/api/v1/auth/service-accounts/"+args[0]+"/rotate-secret", nil)
		if err != nil {
			return err
		}

		var resp struct {
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return err
		}

		fmt.Printf("Secret rotated for: %s\n", resp.ClientID)
		fmt.Printf("New Client Secret: %s\n", resp.ClientSecret)
		fmt.Println("\nSave this secret — it will not be shown again.")
		return nil
	},
}
