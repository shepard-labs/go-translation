package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newProvidersCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "providers", Short: "List supported providers"}
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List supported providers",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "anthropic")
			fmt.Fprintln(cmd.OutOrStdout(), "gemini")
			fmt.Fprintln(cmd.OutOrStdout(), "openrouter")
		},
	})
	return cmd
}

func newModelsCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "models", Short: "Show model examples"}
	cmd.AddCommand(&cobra.Command{
		Use:   "examples",
		Short: "Show example model names",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "anthropic: claude-3-5-sonnet-latest")
			fmt.Fprintln(cmd.OutOrStdout(), "gemini: gemini-1.5-pro")
			fmt.Fprintln(cmd.OutOrStdout(), "openrouter: anthropic/claude-3.5-sonnet")
		},
	})
	return cmd
}
