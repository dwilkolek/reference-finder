package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "reference-finder",
	Short: "reference-finder tool to find random references across repositories(directories)",
	Long: `Written in go. It goes through specified repos and looks for references to connect them.
What is resource? RTFM! 🤠
			`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No idea what do you expect 🤨")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		rootCmd.Help()
	}
}
