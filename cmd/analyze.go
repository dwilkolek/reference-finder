package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/dwilkolek/reference-finder/cmd/runner"
	"github.com/spf13/cobra"
)

func init() {
	analyzeCmd.PersistentFlags().StringP("input", "i", "input.json", "Input file")
	analyzeCmd.PersistentFlags().StringSlice("rootlike", make([]string, 0), "Repositories that should be treated as root")
	analyzeCmd.PersistentFlags().StringP("reg", "r", "", "Reference regexp with one capturing group")
	analyzeCmd.PersistentFlags().Int16P("concurrency", "c", 8, "Amount of coroutines to use")
	analyzeCmd.PersistentFlags().String("trim-suffix", "", "Trim matching suffixes from tags")
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Finds all references specified by parameters",
	Long:  `Purpose is to map repositories into single json mapping that could be used to render flowchart.`,
	Run: func(cmd *cobra.Command, args []string) {
		input, err := cmd.Flags().GetString("input")
		if err != nil {
			fmt.Println("input required")
			os.Exit(1)
		}

		output, _ := cmd.Flags().GetString("output")

		referenceRegexpStr, err := cmd.Flags().GetString("reg")
		if err != nil || len(referenceRegexpStr) == 0 {
			fmt.Println("reference regexp required")
			os.Exit(1)
		}

		rootLike, _ := cmd.Flags().GetStringSlice("rootlike")

		referenceRegexp := regexp.MustCompile(referenceRegexpStr)
		concurrency, _ := cmd.Flags().GetInt16("concurrency")

		runner.Execute(runner.Config{
			InputFile:       input,
			OutputFile:      output,
			RootLike:        rootLike,
			ReferenceRegexp: referenceRegexp,
			Concurrency:     concurrency,
		})
	},
}
