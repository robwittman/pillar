package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configCreateCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configUpdateCmd)
	configCmd.AddCommand(configDeleteCmd)

	// create flags
	addConfigFlags(configCreateCmd)
	configCreateCmd.MarkFlagRequired("provider")
	configCreateCmd.MarkFlagRequired("model")

	// update flags
	addConfigFlags(configUpdateCmd)
}

var (
	cfgProvider      string
	cfgModel         string
	cfgSystemPrompt  string
	cfgAPICredential string
	cfgMaxIterations int
	cfgTokenBudget   int
	cfgTaskTimeout   int
	cfgAllowedTools  []string
	cfgDeniedTools   []string
)

func addConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&cfgProvider, "provider", "", "Model provider (e.g. claude, openai)")
	cmd.Flags().StringVar(&cfgModel, "model", "", "Model ID")
	cmd.Flags().StringVar(&cfgSystemPrompt, "system-prompt", "", "System prompt")
	cmd.Flags().StringVar(&cfgAPICredential, "api-credential", "", "API credential reference")
	cmd.Flags().IntVar(&cfgMaxIterations, "max-iterations", 0, "Max iterations")
	cmd.Flags().IntVar(&cfgTokenBudget, "token-budget", 0, "Token budget")
	cmd.Flags().IntVar(&cfgTaskTimeout, "task-timeout", 0, "Task timeout in seconds")
	cmd.Flags().StringSliceVar(&cfgAllowedTools, "allowed-tools", nil, "Allowlist of builtin tools (comma-separated)")
	cmd.Flags().StringSliceVar(&cfgDeniedTools, "denied-tools", nil, "Denylist of builtin tools (comma-separated)")
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage agent configurations",
}

var configCreateCmd = &cobra.Command{
	Use:   "create <agent-id>",
	Short: "Create agent config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := buildConfigBody()
		data, _, err := getClient().Post("/api/v1/agents/"+args[0]+"/config", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <agent-id>",
	Short: "Get agent config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/agents/" + args[0] + "/config")
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var configUpdateCmd = &cobra.Command{
	Use:   "update <agent-id>",
	Short: "Update agent config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := buildConfigBody()
		data, _, err := getClient().Put("/api/v1/agents/"+args[0]+"/config", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete <agent-id>",
	Short: "Delete agent config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := getClient().Delete("/api/v1/agents/" + args[0] + "/config")
		if err != nil {
			return err
		}
		fmt.Printf("Config for agent %s deleted\n", args[0])
		return nil
	},
}

func buildConfigBody() map[string]any {
	body := map[string]any{}
	if cfgProvider != "" {
		body["model_provider"] = cfgProvider
	}
	if cfgModel != "" {
		body["model_id"] = cfgModel
	}
	if cfgSystemPrompt != "" {
		body["system_prompt"] = cfgSystemPrompt
	}
	if cfgAPICredential != "" {
		body["api_credential"] = cfgAPICredential
	}
	if cfgMaxIterations != 0 {
		body["max_iterations"] = cfgMaxIterations
	}
	if cfgTokenBudget != 0 {
		body["token_budget"] = cfgTokenBudget
	}
	if cfgTaskTimeout != 0 {
		body["task_timeout_seconds"] = cfgTaskTimeout
	}
	if len(cfgAllowedTools) > 0 || len(cfgDeniedTools) > 0 {
		tp := map[string]any{}
		if len(cfgAllowedTools) > 0 {
			tp["allowed_tools"] = cfgAllowedTools
		}
		if len(cfgDeniedTools) > 0 {
			tp["denied_tools"] = cfgDeniedTools
		}
		body["tool_permissions"] = tp
	}
	return body
}
