package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "multica",
	Short: "Multica CLI — local agent runtime and management tool",
	Long:  "Work seamlessly with Multica from the command line.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().String("server-url", "", "Multica server URL (env: MULTICA_SERVER_URL)")
	rootCmd.PersistentFlags().String("workspace-id", "", "Workspace ID (env: MULTICA_WORKSPACE_ID)")
	rootCmd.PersistentFlags().String("profile", "", "Configuration profile name (e.g. dev) — isolates config, daemon state, and workspaces")

	// Core commands — primary task management.
	setGroup(issueCmd, groupCore)
	setGroup(agentCmd, groupCore)
	setGroup(workspaceCmd, groupCore)

	// Runtime commands — agent execution.
	setGroup(daemonCmd, groupRuntime)

	// Additional commands (default group for the rest).
	setGroup(loginCmd, groupAdditional)
	setGroup(authCmd, groupAdditional)
	setGroup(configCmd, groupAdditional)
	setGroup(repoCmd, groupAdditional)
	setGroup(updateCmd, groupAdditional)
	setGroup(versionCmd, groupAdditional)

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(agentCmd)
	rootCmd.AddCommand(workspaceCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(issueCmd)
	rootCmd.AddCommand(repoCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)

	// Apply gh-style help templates.
	rootCmd.SetHelpFunc(rootHelpFunc)
	applyHelpFuncs(rootCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
