package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/dwilkolek/reference-finder/cmd/runner"
	"github.com/spf13/cobra"
)

func init() {
	flowchart.PersistentFlags().StringP("input", "i", "output.json", "Input file")
	flowchart.PersistentFlags().StringP("output", "o", "flowchart.txt", "Output file")
	rootCmd.AddCommand(flowchart)
}

var flowchart = &cobra.Command{
	Use:   "flowchart",
	Short: "Generates mermaid flowchartfor given json",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		jsonFile, err := os.Open(input)
		if err != nil {
			fmt.Printf("Failed to read file %s: %s\n", input, err)
			os.Exit(1)
		}
		defer jsonFile.Close()
		data, _ := io.ReadAll(jsonFile)
		var resources []runner.Resource

		err = json.Unmarshal([]byte(data), &resources)

		if err != nil {
			fmt.Printf("Failed to parse json from file %s: %s\n", input, err)
			os.Exit(1)
		}

		flowchart := runner.GenerateFlowchart(resources)

		fmt.Printf("Saving to %s\n", output)
		os.Remove(output)
		os.WriteFile(output, []byte(flowchart), 0644)
	},
}
