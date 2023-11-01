package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

func fetchRepo(repo Repository, config Config) string {
	path := fmt.Sprintf("%s/%s", config.WorkDir, repo.Name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Fetching repo %s\n", repo.Name)
		cmd := exec.Command("gh", "repo", "clone", repo.Url, path)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed fetching repo, %s\n", repo.Name)
			os.Exit(1)
		}
	}
	return path
}

func findReferences(forTag string, startingPath string, cfg Config) map[string][]string {
	var refMap = make(map[string][]string)

	filepath.Walk(startingPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				refMap = mergeRefs(refMap, referencesInFile(forTag, path, cfg))
			}
			return nil
		})
	return refMap
}

func referencesInFile(exludeTag string, file string, cfg Config) map[string][]string {
	refs := make(map[string][]string)

	readFile, err := os.Open(file)

	if err != nil {
		fmt.Printf("Failed to read file %s: %s\n", file, err)
		os.Exit(1)
	}
	defer readFile.Close()
	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)
	line := 0
	for fileScanner.Scan() {
		line++
		content := fileScanner.Text()
		matches := cfg.ReferenceRegexp.FindAllStringSubmatch(content, -1)
		if len(matches) > 0 {
			for _, match := range matches {
				foundTag := strings.TrimSuffix(match[1], cfg.TrimSuffix)
				if exludeTag != foundTag && !slices.Contains(cfg.Exclude, foundTag) {
					references, ok := refs[foundTag]
					ref := strings.TrimPrefix(fmt.Sprintf("%s:%d", file, line), cfg.WorkDir)
					if ok {
						refs[foundTag] = append(references, ref)
					} else {
						refs[foundTag] = []string{ref}
					}
				}

			}

		}
	}
	return refs
}

func GenerateFlowchart(resources []Resource) string {
	flowchart := "flowchart TD\n"
	for _, resource := range resources {
		source := resource.Tag
		for dep, _ := range resource.References {
			flowchart = flowchart + fmt.Sprintf("\t%s --->|depends on| %s\n", source, dep)
		}
	}
	return flowchart
}

func mergeRefs(m1 map[string][]string, m2 map[string][]string) map[string][]string {
	merged := make(map[string][]string)
	for k, v := range m1 {
		merged[k] = v
	}
	for key, value := range m2 {
		merged[key] = append(merged[key], value...)
	}

	for key, value := range merged {
		merged[key] = unique(value)
	}
	return merged
}

func unique(in []string) []string {
	var unique []string
	m := map[string]bool{}

	for _, v := range in {
		if !m[v] {
			m[v] = true
			unique = append(unique, v)
		}
	}
	return unique
}

func readInputFile(inputFile string) []Repository {
	jsonFile, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Failed to read file %s: %s\n", inputFile, err)
		os.Exit(1)
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)
	var repos []Repository

	_ = json.Unmarshal([]byte(data), &repos)
	return repos
}
