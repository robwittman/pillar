package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	taskAgentID string
	taskPrompt  string
	taskStatus  string
)

func init() {
	rootCmd.AddCommand(taskCmd)
	taskCmd.AddCommand(taskCreateCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskGetCmd)

	taskCreateCmd.Flags().StringVar(&taskAgentID, "agent-id", "", "Agent ID")
	taskCreateCmd.Flags().StringVar(&taskPrompt, "prompt", "", "Task prompt")
	taskCreateCmd.MarkFlagRequired("agent-id")
	taskCreateCmd.MarkFlagRequired("prompt")

	taskListCmd.Flags().StringVar(&taskAgentID, "agent-id", "", "Filter by agent ID")
	taskListCmd.Flags().StringVar(&taskStatus, "status", "", "Filter by status (pending/assigned/running/completed/failed)")
}

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
}

var taskCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a task for an agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		body := map[string]any{
			"agent_id": taskAgentID,
			"prompt":   taskPrompt,
		}
		data, _, err := getClient().Post("/api/v1/tasks", body)
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "/api/v1/tasks"
		params := ""
		if taskAgentID != "" {
			params += "agent_id=" + taskAgentID
		}
		if taskStatus != "" {
			if params != "" {
				params += "&"
			}
			params += "status=" + taskStatus
		}
		if params != "" {
			path += "?" + params
		}

		data, _, err := getClient().Get(path)
		if err != nil {
			return err
		}
		var tasks []struct {
			ID      string `json:"id"`
			AgentID string `json:"agent_id"`
			Status  string `json:"status"`
			Prompt  string `json:"prompt"`
			Created string `json:"created_at"`
		}
		if err := json.Unmarshal(data, &tasks); err != nil {
			return printJSON(data)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tAGENT\tSTATUS\tPROMPT\tCREATED")
		for _, t := range tasks {
			prompt := t.Prompt
			if len(prompt) > 60 {
				prompt = prompt[:60] + "..."
			}
			agentID := t.AgentID
			if len(agentID) > 8 {
				agentID = agentID[:8]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.ID[:8], agentID, t.Status, prompt, t.Created)
		}
		return w.Flush()
	},
}

var taskGetCmd = &cobra.Command{
	Use:   "get <task-id>",
	Short: "Get a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, _, err := getClient().Get("/api/v1/tasks/" + args[0])
		if err != nil {
			return err
		}
		return printJSON(data)
	},
}
