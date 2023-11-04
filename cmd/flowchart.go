package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dwilkolek/reference-finder/cmd/runner"
	"github.com/spf13/cobra"
)

func init() {
	flowchart.PersistentFlags().StringP("input", "i", "output.json", "Input file")
	flowchart.PersistentFlags().StringP("output", "o", "flowchart.txt", "Output file")
	flowchart.PersistentFlags().StringP("resource", "r", "", "Chart for single resource")
	flowchart.PersistentFlags().StringP("exclude", "e", "", "Exclude from chart")
	flowchart.PersistentFlags().StringP("group-definitions", "g", "", "Group definitions specification")
	flowchart.PersistentFlags().Bool("include-orphans", false, "Include orphan center")
	flowchart.PersistentFlags().StringP("valid-tags", "v", "", "List of valid tags")
	flowchart.PersistentFlags().StringP("translation", "t", "", "Mapping tags to display names. One line - one translation. Separated by ;.")

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

		validTagsFile, _ := cmd.Flags().GetString("valid-tags")
		validTags := []string{}
		if len(validTagsFile) > 0 {
			validTags, _ = readLines(validTagsFile)
		}

		excludeFile, _ := cmd.Flags().GetString("exclude")
		exclude := []string{}
		if len(excludeFile) > 0 {
			exclude, _ = readLines(excludeFile)
		}

		translationMappingFile, _ := cmd.Flags().GetString("translation")
		translationMapping := map[string]string{}
		if len(translationMappingFile) > 0 {
			translations, _ := readLines(translationMappingFile)
			for _, t := range translations {
				pair := strings.Split(t, ";")
				translationMapping[pair[0]] = pair[1]
			}
		}

		flowchart := runner.GenerateFlowchart(resources, tag, exclude, readGrouppingFile(groupDefinitions), orphanCenter, validTags, translationMapping)

		fmt.Printf("Saving to %s\n", output)
		os.Remove(output)
		err = os.WriteFile(output, []byte(flowchart), 0644)
		if err != nil {
			fmt.Println(err)
		}
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

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
