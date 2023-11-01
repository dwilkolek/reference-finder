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
	analyzeCmd.PersistentFlags().StringSliceP("exclude", "e", make([]string, 0), "Exclude tags from output")
	analyzeCmd.PersistentFlags().Int16P("concurrency", "c", 8, "Amount of coroutines to use")
	analyzeCmd.PersistentFlags().Bool("remove-entries-without-dependencies-from-output", false, "Amount of coroutines to use")
	analyzeCmd.PersistentFlags().String("trim-suffix", "", "Trim matching suffixes from tags")
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analize",
	Short: "Finds all references specified by parameters",
	Long:  `TODO`,
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

		exclude, _ := cmd.Flags().GetStringSlice("exclude")
		rootLike, _ := cmd.Flags().GetStringSlice("rootlike")

		referenceRegexp := regexp.MustCompile(referenceRegexpStr)
		concurrency, _ := cmd.Flags().GetInt16("concurrency")
		removeEmpty, _ := cmd.Flags().GetBool("remove-entries-without-dependencies-from-output")

		runner.Execute(runner.Config{
			InputFile:              input,
			OutputFile:             output,
			Exclude:                exclude,
			RootLike:               rootLike,
			ReferenceRegexp:        referenceRegexp,
			KeepWithNoDependencies: !removeEmpty,
			Concurrency:            concurrency,
		})
	},
}