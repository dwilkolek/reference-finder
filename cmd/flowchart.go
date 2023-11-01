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
	flowchart.PersistentFlags().StringP("resource", "r", "", "Chart for single resource")
	flowchart.PersistentFlags().StringSliceP("exclude", "e", make([]string, 0), "Exclude from chart")
	flowchart.PersistentFlags().StringP("group-definitions", "g", "", "Group definitions specification")
	flowchart.PersistentFlags().Bool("include-orphans", false, "Include orphan center")

	rootCmd.AddCommand(flowchart)
}

var flowchart = &cobra.Command{
	Use:   "flowchart",
	Short: "Generates mermaid flowchartfor given json",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		tag, _ := cmd.Flags().GetString("resource")
		exclude, _ := cmd.Flags().GetStringSlice("exclude")

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

		groupDefinitions, _ := cmd.Flags().GetString("group-definitions")

		orphanCenter, _ := cmd.Flags().GetBool("include-orphans")

		flowchart := runner.GenerateFlowchart(resources, tag, exclude, readGrouppingFile(groupDefinitions), orphanCenter)

		fmt.Printf("Saving to %s\n", output)
		os.Remove(output)
		os.WriteFile(output, []byte(flowchart), 0644)
	},
}

func readGrouppingFile(file string) map[string][]string {
	if len(file) == 0 {
		return map[string][]string{}
	}

	jsonFile, err := os.Open(file)
	if err != nil {
		fmt.Printf("Failed to read file %s: %s\n", file, err)
		os.Exit(1)
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)
	var groups map[string][]string

	err = json.Unmarshal([]byte(data), &groups)

	if err != nil {
		fmt.Printf("Failed to parse json from file %s: %s\n", file, err)
		os.Exit(1)
	}
	return groups
}
