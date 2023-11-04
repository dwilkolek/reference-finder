package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

func fetchRepo(repo Repository, executionConfig ExecutionConfig) string {
	path := fmt.Sprintf("%s/%s", executionConfig.WorkDir, repo.Name)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Printf("Fetching repo %s\n", repo.Name)
		cmd := exec.Command("gh", "repo", "clone", repo.Url, path)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed fetching repo, %s\n", repo.Name)
			os.Exit(1)
		}
	}
	if executionConfig.Sync {
		fmt.Printf("Syncing repo %s\n", repo.Name)
		cmd := exec.Command("git", "pull")
		cmd.Dir = path
		if err := cmd.Run(); err != nil {
			fmt.Printf("!!!! Failed syncing repo, %s\n", repo.Name)
		}
	}
	return path
}

type Findings struct {
	References map[string][]string
	Software   []string
}

func findReferences(forTag string, startingPath string, executionConfig ExecutionConfig) Findings {
	var refMap = make(map[string][]string)
	software := []string{}
	filepath.Walk(startingPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				var fxs = make([]runOnLine, 2)

				refs := make(map[string][]string)
				fxs[0] = func(line int, file string, content string) {
					matches := executionConfig.ReferenceRegexp.FindAllStringSubmatch(content, -1)
					if len(matches) > 0 {
						for _, match := range matches {
							foundTag := strings.TrimSuffix(match[1], executionConfig.TrimSuffix)
							if forTag != foundTag {
								references, ok := refs[foundTag]
								ref := strings.TrimPrefix(fmt.Sprintf("%s:%d", file, line), executionConfig.WorkDir)
								if ok {
									refs[foundTag] = append(references, ref)
								} else {
									refs[foundTag] = []string{ref}
								}
							}

						}

					}

				}
				fxs[1] = func(line int, file string, content string) {
					software = append(software, findSoftware(file, content)...)
				}
				referencesInFile(forTag, path, executionConfig, fxs)
				refMap = mergeRefs(refMap, refs, executionConfig.ValidNames)
				software = unique(software)
			}

			return nil
		})

	return Findings{
		References: refMap,
		Software:   software,
	}
}

var dockerReg = regexp.MustCompile("FROM ([A-Za-z0-9-/]+:[A-Za-z0-9-]+)")
var springReg = regexp.MustCompile("(?:springBootVersion = '|org.springframework.boot\"\\) version \")([0-9A-Z.]+)")
var frontendReg = regexp.MustCompile("\"(typescript|node|react|aws-cdk)\": \"([0-9A-Za-z.]+)\"")
var kotlinReg = regexp.MustCompile("(?:kotlin\\(\"jvm\"\\) version \"|kotlin_version = ')(([0-9\\.]+))")
var mapping = map[string]string{
	"adoptopenjdk/openjdk11":      "Java 11",
	"eclipse-temurin:17-jre":      "Java 17",
	"eclipse-temurin:17":          "Java 17",
	"python:3.10-slim":            "Python 3.10",
	"adoptopenjdk:11-jre-hotspot": "Java 11",
	"eclipse-temurin:19-jre":      "Java 19",
	"node:16-alpine":              "Node 16",
	"nginx:1":                     "",
	"alpine:latest":               "",
	"node:16":                     "Node 16",
	"node:14":                     "Node 14",
	"node:14-alpine":              "Node 14",
	"python:3":                    "Python 3",
	"sonarqube:9":                 "",
	"cypress/included:12":         "",
	"cypress/included:10":         "",
	"cypress/included:13":         "",
	"continuumio/miniconda3:4":    "",
	"ubuntu:18":                   "",
	"golang:alpine":               "Golang",
	"phlptp/units:webserver":      "C++",
	"node:16-buster-slim":         "Node 16",
	"gradle:7":                    "",
	"eclipse-temurin:17-jdk":      "Java 17",
	"debian:buster-slim":          "",
	"node:18":                     "Node 18",
	"postgres:13":                 "",
}

func findSoftware(file string, content string) []string {
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	value := ""
	if filename == "Dockerfile" {
		matches := dockerReg.FindAllStringSubmatch(content, -1)
		if len(matches) > 0 {
			value = matches[0][1]
			if len(value) > 0 {
				mappedValue, ok := mapping[value]
				if ok {
					if len(mappedValue) > 0 {
						return []string{mappedValue}
					}
				} else {
					fmt.Printf("UNKNOWN %s", value)
					return []string{value}
				}
			}
		}
	}
	if filename == "build.gradle" || filename == "build.gradle.kts" {
		// ext.kotlin_version = '1.4.31'
		matches := springReg.FindAllStringSubmatch(content, -1)
		if len(matches) > 0 {
			value = matches[0][1]
			return []string{"Spring " + value}
		}

		matchesKotlin := kotlinReg.FindAllStringSubmatch(content, -1)
		if len(matchesKotlin) > 0 {
			value = matchesKotlin[0][1]
			return []string{"Kotlin " + value}
		}
		// ext.springBootVersion = '2.2.13.RELEASE'
		// springBootVersion = '2.7.2'
	}
	if filename == "package.json" {
		matches := frontendReg.FindAllStringSubmatch(content, -1)
		if len(matches) > 0 {
			return []string{strings.ToUpper(string(matches[0][1][0])) + matches[0][1][1:] + " " + matches[0][2]}
		}
		// "aws-cdk": "2.50.0",
		// "react": "18.2.0",
		// "typescript": "~3.9.7"
		// "node": ">=18.7.0",
	}

	return []string{}
}

type runOnLine func(int, string, string)

func referencesInFile(exludeTag string, file string, executionConfig ExecutionConfig, fxs []runOnLine) map[string][]string {
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
		for _, f := range fxs {
			f(line, file, content)
		}
	}
	return refs
}

func mergeRefs(m1 map[string][]string, m2 map[string][]string, validNames []string) map[string][]string {
	merged := make(map[string][]string)
	for k, v := range m1 {
		if len(validNames) == 0 || slices.Contains(validNames, k) {
			merged[k] = v
		}
	}
	for key, value := range m2 {
		if len(validNames) == 0 || slices.Contains(validNames, key) {
			merged[key] = append(merged[key], value...)
		}

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
