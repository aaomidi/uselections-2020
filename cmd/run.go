package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the bot",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO Run the bot
		return nil
	},
}
