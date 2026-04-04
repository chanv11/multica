package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// Command group annotation key.
const cmdGroupKey = "group"

// Group names used to categorize commands in help output.
const (
	groupCore       = "core"
	groupRuntime    = "runtime"
	groupAdditional = "additional"
)

// setGroup annotates a command with a help group.
func setGroup(cmd *cobra.Command, group string) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[cmdGroupKey] = group
}

// commandsByGroup collects visible subcommands by group name.
// Commands without an annotation go into "additional".
func commandsByGroup(cmd *cobra.Command) map[string][]*cobra.Command {
	groups := map[string][]*cobra.Command{}
	for _, c := range cmd.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		g := groupAdditional
		if v, ok := c.Annotations[cmdGroupKey]; ok {
			g = v
		}
		groups[g] = append(groups[g], c)
	}
	return groups
}

// formatCommandList renders a list of commands in gh style:
//
//	name:          description
func formatCommandList(cmds []*cobra.Command) string {
	if len(cmds) == 0 {
		return ""
	}

	// Find max command name length for alignment.
	maxLen := 0
	for _, c := range cmds {
		if len(c.Name()) > maxLen {
			maxLen = len(c.Name())
		}
	}

	var b strings.Builder
	for _, c := range cmds {
		padding := strings.Repeat(" ", maxLen-len(c.Name()))
		fmt.Fprintf(&b, "  %s:%s   %s\n", c.Name(), padding, c.Short)
	}
	return b.String()
}

// rootHelpFunc returns a custom help function for the root command (gh style).
func rootHelpFunc(cmd *cobra.Command, _ []string) {
	groups := commandsByGroup(cmd)

	fmt.Println(cmd.Long)
	fmt.Println()
	fmt.Println("USAGE")
	fmt.Printf("  %s <command> <subcommand> [flags]\n", cmd.Name())
	fmt.Println()

	// Print command groups in order.
	type section struct {
		title string
		key   string
	}
	sections := []section{
		{"CORE COMMANDS", groupCore},
		{"RUNTIME COMMANDS", groupRuntime},
		{"ADDITIONAL COMMANDS", groupAdditional},
	}

	for _, s := range sections {
		cmds := groups[s.key]
		if len(cmds) == 0 {
			continue
		}
		fmt.Println(s.title)
		fmt.Print(formatCommandList(cmds))
		fmt.Println()
	}

	fmt.Println("FLAGS")
	fmt.Println("  --help      Show help for command")
	fmt.Println("  --version   Show multica version")
	fmt.Println()

	fmt.Println("ENVIRONMENT VARIABLES")
	fmt.Println("  MULTICA_SERVER_URL      Multica server URL")
	fmt.Println("  MULTICA_WORKSPACE_ID    Default workspace ID")
	fmt.Println("  MULTICA_TOKEN           Authentication token")
	fmt.Println()

	fmt.Println("EXAMPLES")
	fmt.Println("  $ multica login")
	fmt.Println("  $ multica issue list --status todo")
	fmt.Println("  $ multica daemon start")
	fmt.Println()

	fmt.Println("LEARN MORE")
	fmt.Printf("  Use `%s <command> <subcommand> --help` for more information about a command.\n", cmd.Name())
	fmt.Println("  Read the documentation at https://multica.ai/docs")
}

// subcommandHelpFunc returns a custom help function for group commands (issue, agent, etc.)
func subcommandHelpFunc(cmd *cobra.Command, _ []string) {
	groups := commandsByGroup(cmd)

	// Print description.
	if cmd.Long != "" {
		fmt.Println(cmd.Long)
	} else {
		fmt.Println(cmd.Short)
	}
	fmt.Println()

	// Usage line.
	fmt.Println("USAGE")
	fmt.Printf("  %s %s <command> [flags]\n", cmd.Root().Name(), cmd.Name())
	fmt.Println()

	// Print grouped subcommands.
	type section struct {
		title string
		key   string
	}
	sections := []section{
		{"GENERAL COMMANDS", groupCore},
		{"TARGETED COMMANDS", groupAdditional},
	}

	// If no groups are annotated, print all as COMMANDS.
	hasGroups := false
	for _, s := range sections {
		if len(groups[s.key]) > 0 && s.key != groupAdditional {
			hasGroups = true
			break
		}
	}

	if !hasGroups {
		// No group annotations — print flat list.
		var all []*cobra.Command
		for _, cmds := range groups {
			all = append(all, cmds...)
		}
		if len(all) > 0 {
			fmt.Println("COMMANDS")
			fmt.Print(formatCommandList(all))
			fmt.Println()
		}
	} else {
		for _, s := range sections {
			cmds := groups[s.key]
			if len(cmds) == 0 {
				continue
			}
			fmt.Println(s.title)
			fmt.Print(formatCommandList(cmds))
			fmt.Println()
		}
	}

	// Flags.
	localFlags := cmd.LocalNonPersistentFlags()
	if localFlags.HasFlags() {
		fmt.Println("FLAGS")
		fmt.Println(localFlags.FlagUsages())
	}

	fmt.Println("INHERITED FLAGS")
	fmt.Println("  --help   Show help for command")
	fmt.Println()

	fmt.Println("LEARN MORE")
	fmt.Printf("  Use `%s %s <command> --help` for more information about a command.\n", cmd.Root().Name(), cmd.Name())
}

// leafHelpFunc provides gh-style help for leaf commands (commands with RunE/Run).
func leafHelpFunc(cmd *cobra.Command, _ []string) {
	// Build full command path.
	path := cmd.CommandPath()

	// Description.
	if cmd.Long != "" {
		fmt.Println(cmd.Long)
	} else {
		fmt.Println(cmd.Short)
	}
	fmt.Println()

	// Usage.
	fmt.Println("USAGE")
	fmt.Printf("  %s [flags]\n", path)
	fmt.Println()

	// Local flags.
	localFlags := cmd.LocalNonPersistentFlags()
	if localFlags.HasFlags() {
		fmt.Println("FLAGS")
		fmt.Print(localFlags.FlagUsages())
		fmt.Println()
	}

	// Inherited flags.
	fmt.Println("INHERITED FLAGS")
	fmt.Println("  --help   Show help for command")
	fmt.Println()

	fmt.Println("LEARN MORE")
	fmt.Printf("  Use `%s <command> --help` for more information about a command.\n", cmd.Root().Name())
}

// applyHelpFuncs recursively sets the appropriate help function on all commands.
func applyHelpFuncs(cmd *cobra.Command) {
	for _, c := range cmd.Commands() {
		if c.HasSubCommands() {
			c.SetHelpFunc(subcommandHelpFunc)
		} else if c.Run != nil || c.RunE != nil {
			c.SetHelpFunc(leafHelpFunc)
		}
		applyHelpFuncs(c)
	}
}
