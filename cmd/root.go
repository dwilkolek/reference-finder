package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "reference-finder",
	Short: "reference-finder tool to find random references across repositories(directories)",
	Long: `Written in go. It goes through specified repos in input.json and generates output.json.
Output.json contains found resources and references to other resources. What is resource? 
	- Repository name (expect rootlike repos)
	- Directory in main rootlike repo
	- First capture grup in regex
			`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("No idea what you expect. RTFM!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
