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
	analyzeCmd.PersistentFlags().StringP("config", "i", "config.json", "Config file")
	rootCmd.AddCommand(analyzeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Finds all references specified by json file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Println("Config required")
			os.Exit(1)
		}

		jsonFile, err := os.Open(configFile)
		if err != nil {
			fmt.Printf("Failed to read file %s: %s\n", configFile, err)
			os.Exit(1)
		}
		defer jsonFile.Close()
		data, _ := io.ReadAll(jsonFile)
		var config runner.Config

		_ = json.Unmarshal([]byte(data), &config)
		runner.Execute(config)
	},
}
