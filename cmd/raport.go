package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/dwilkolek/reference-finder/cmd/runner"
	"github.com/spf13/cobra"
)

func init() {
	reportCmd.PersistentFlags().StringP("input", "i", "output.json", "Input file")
	reportCmd.PersistentFlags().StringP("output", "o", "RAPORT.md", "Output file")
	reportCmd.PersistentFlags().StringP("exclude", "e", "", "Exclude from chart")
	reportCmd.PersistentFlags().StringP("group-definitions", "g", "", "Group definitions specification")
	reportCmd.PersistentFlags().StringP("valid-tags", "v", "", "List of valid tags")
	reportCmd.PersistentFlags().StringP("translation", "t", "", "Mapping tags to display names. One line - one translation. Separated by ;.")

	rootCmd.AddCommand(reportCmd)
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generates markdown report",
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

		groupDefinitions, _ := cmd.Flags().GetString("group-definitions")
		groups := readGrouppingFile(groupDefinitions)

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

		reportMd := ""
		reportEntires := map[int][]string{}
		for _, resource := range resources {
			priority := 0
			reportEntry := ""
			if slices.Contains(exclude, resource.Tag) {
				continue
			}
			if len(validTags) > 0 && !slices.Contains(validTags, resource.Tag) {
				continue
			}
			fmt.Printf("Adding to report %s\n", resource.Tag)
			appName := resource.Tag
			translated, ok := translationMapping[appName]
			if ok {
				fmt.Printf("Transalted %s to %s", appName, translated)
				appName = translated
			}
			reportEntry += fmt.Sprintf("## %s\n\n", appName)
			for gName, gApps := range groups {
				if slices.Contains(gApps, resource.Tag) {
					priority++
					reportEntry += fmt.Sprintf("### Team: %s\n\n", gName)
				}
			}

			softPart := "### Software:\n\n"
			hasSoftware := false
			for _, soft := range resource.Software {
				hasSoftware = true
				priority++
				softPart += fmt.Sprintf("- %s\n", soft)
			}
			softPart += "\n\n"
			if hasSoftware {
				reportEntry += softPart
			}

			depsPart := "### Dependencies:\n\n"
			hasDeps := false
			for ref := range resource.References {
				if slices.Contains(exclude, ref) {
					continue
				}
				if len(validTags) > 0 && !slices.Contains(validTags, ref) {
					continue
				}
				hasDeps = true
				priority++
				depsPart += fmt.Sprintf("- %s\n", ref)
			}
			depsPart += "\n\n"
			if hasDeps {

				reportEntry += depsPart
			}
			reportEntires[priority] = append(reportEntires[priority], reportEntry)
		}

		maxPrio := 1000
		for prio := maxPrio; prio > -1; prio-- {
			for _, entry := range reportEntires[prio] {
				reportMd += entry
			}
		}

		fmt.Printf("Saving to %s\n", output)
		os.Remove(output)
		err = os.WriteFile(output, []byte(reportMd), 0644)
		if err != nil {
			fmt.Println(err)
		}
	},
}
