package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
)

type Repository struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type Resource struct {
	Source       string   `json:"source"`
	Dependencies []string `json:"dependencies"`
}

var lock sync.Mutex
var wg sync.WaitGroup

var exludeList = []string{
	"gandalf", "dsi-service-config", "environment-variables", "dxp", "login-service", "access-management", "api-manager", "api-login",
}
var skipRepos = []string{
	"gandalf", "dsi-service-config", "dxp", "login-service", "access-management", "api-manager", "api-login",
}

var reg = regexp.MustCompile("(?m)(?:http|https)://([a-zA-Z0-9-]+)(?:.dev|.demo){0,1}.service")

func main() {
	_ = os.Mkdir("tmp", os.ModePerm)

	var resources []Resource
	jsonFile, err := os.Open("gh.json")
	if err != nil {
		fmt.Println("Create gh.json")
		os.Exit(1)
	}
	defer jsonFile.Close()
	data, _ := io.ReadAll(jsonFile)
	var repos []Repository

	var reposGh []Repository
	_ = json.Unmarshal([]byte(data), &reposGh)
	repos = append(repos, reposGh...)

	maxGoroutines := 10
	guard := make(chan struct{}, maxGoroutines)
	done := 0
	for _, repo := range repos {
		wg.Add(1)
		guard <- struct{}{} // would block if guard channel is already filled
		go func(r Repository) {
			foundResource := process(r, repos)
			lock.Lock()
			resources = append(resources, foundResource...)
			lock.Unlock()
			done = done + 1
			fmt.Printf("DONE %d/%d \n", done, len(repos))
			wg.Done()
			<-guard
		}(repo)
	}

	wg.Wait()

	finalMap := make(map[string][]string)
	for _, resource := range resources {
		deps, ok := finalMap[resource.Source]
		if ok {
			for _, newDep := range resource.Dependencies {
				if !slices.Contains(deps, newDep) {
					deps = append(deps, newDep)
				}
			}
		} else {
			finalMap[resource.Source] = resource.Dependencies
		}
	}

	outBytes, _ := json.Marshal(resources)
	os.Remove("output.json")
	os.WriteFile("output.json", outBytes, 0644)

	printFlowchart(finalMap)
}

func process(repo Repository, allRepos []Repository) []Resource {
	if slices.Contains(skipRepos, repo.Name) {
		fmt.Println(repo.Name + " Skipping")
		return []Resource{}
	}
	fmt.Println(repo.Name + " Processing ")
	location := fetchRepo(repo)

	if repo.Name == "environment-variables" {
		entries, err := os.ReadDir(location)
		if err != nil {
			log.Fatal(err)
		}
		nestedResources := []Resource{}
		for _, e := range entries {
			if e.IsDir() {
				nestedAppName := e.Name()
				nestedLocation := fmt.Sprintf("%s/%s", location, nestedAppName)
				dependencies := findDependencies(Repository{
					Name: nestedAppName,
				}, nestedLocation, allRepos)
				nestedResources = append(nestedResources, Resource{Source: "app:" + nestedAppName, Dependencies: dependencies})
			}

		}
		return nestedResources
	}

	dependencies := findDependencies(repo, location, allRepos)

	fmt.Printf("%s Found %d\n", repo.Name, len(dependencies))
	return []Resource{{
		Source:       "app:" + repo.Name,
		Dependencies: dependencies,
	}}
}

func findDependencies(repo Repository, startingPath string, allRepos []Repository) []string {
	var depsMap = make(map[string]bool)
	var deps = make([]string, 0)

	filepath.Walk(startingPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				matches := reg.FindAllStringSubmatch(string(content), -1)
				if len(matches) > 0 {
					for _, match := range matches {
						appName := strings.TrimSuffix(match[1], "-dev")
						if !depsMap[appName] && appName != repo.Name && !slices.Contains(exludeList, appName) {
							depsMap[appName] = true
							for _, validName := range allRepos {
								if validName.Name == appName {
									deps = append(deps, "app:"+appName)
									break
								}
							}

						}

					}

				}
			}
			return nil
		})

	return deps
}

func fetchRepo(repo Repository) string {
	path := fmt.Sprintf("tmp/%s", repo.Name)

	if _, err := os.Stat(path); os.IsNotExist(err) {

		fmt.Println(repo.Name + " Fetching repo")
		cmd := exec.Command("gh", "repo", "clone", repo.Url, path)
		if err := cmd.Run(); err != nil {
			fmt.Println(repo.Name + " Failed fetching repo")
			os.Exit(1)
		}
	}
	return path
}

func printFlowchart(resources map[string][]string) {
	flowchart := "flowchart TD\n"
	for source, deps := range resources {
		for _, dep := range deps {
			flowchart = flowchart + fmt.Sprintf("\t%s --->|depends on| %s\n", source, dep)
		}
	}
	fmt.Println(flowchart)
	os.Remove("flowchart.txt")
	os.WriteFile("flowchart.txt", []byte(flowchart), 0644)
}
