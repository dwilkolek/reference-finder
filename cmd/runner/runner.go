package runner

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"sync"
	"time"
)

type Repository struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type Resource struct {
	Tag        string              `json:"tag"`
	References map[string][]string `json:"references"`
	Software   []string            `json:"software"`
}

var wg sync.WaitGroup

type Config struct {
	ReferenceRegexp *regexp.Regexp    `json:"reg"`
	RootLike        []string          `json:"rootlike"`
	Concurrency     int16             `json:"concurrency"`
	InputFile       string            `json:"input"`
	OutputFile      string            `json:"output"`
	TrimSuffix      string            `json:"trimSuffix"`
	Sync            bool              `json:"sync"`
	ExtendedSearch  bool              `json:"extendedSearch"`
	Aliases         map[string]string `json:"aliases"`
}

type ExecutionConfig struct {
	Config
	Repositories []Repository
	ValidNames   []string
	WorkDir      string
}

type collector struct {
	executionConfig ExecutionConfig
	resources       map[string]Resource
	lock            sync.Mutex
}

func executionConfig(config Config) ExecutionConfig {
	repositories := readInputFile(config.InputFile)
	validNames := []string{}
	if !config.ExtendedSearch {
		for _, r := range repositories {
			validNames = append(validNames, r.Name)
		}
	}
	return ExecutionConfig{
		Config:       config,
		Repositories: repositories,
		ValidNames:   validNames,
	}
}

func (collector *collector) outputResourcesList() []Resource {
	v := make([]Resource, 0)

	for _, res := range collector.resources {
		v = append(v, res)
	}
	return v
}

func (collector *collector) merge(newResources []Resource) {
	collector.lock.Lock()
	defer collector.lock.Unlock()

	for _, newResource := range newResources {
		resource := collector.resources[newResource.Tag]
		merged := mergeRefs(resource.References, newResource.References, collector.executionConfig.ValidNames)
		mergedSoftware := unique(append(resource.Software, newResource.Software...))

		collector.resources[newResource.Tag] = Resource{
			Tag:        newResource.Tag,
			References: merged,
			Software:   mergedSoftware,
		}

		if len(collector.executionConfig.ValidNames) > 0 {
			toRemove := []string{}
			for key := range collector.resources {
				if !slices.Contains(collector.executionConfig.ValidNames, key) {
					toRemove = append(toRemove, key)
				}
			}
			for _, removeKey := range toRemove {
				delete(collector.resources, removeKey)
			}
		}
	}

}

func Execute(config Config) {
	fmt.Printf("Executing with %+v\n", config)
	executionConfig := executionConfig(config)
	fmt.Printf("Entries to process: %d\n", len(executionConfig.Repositories))
	executionConfig.WorkDir = "workdir"
	collector := collector{
		executionConfig: executionConfig,
		resources:       map[string]Resource{},
		lock:            sync.Mutex{},
	}

	_ = os.Mkdir(executionConfig.WorkDir, os.ModePerm)

	guard := make(chan struct{}, executionConfig.Concurrency)
	done := 0

	for _, repo := range executionConfig.Repositories {
		wg.Add(1)
		guard <- struct{}{}
		go func(r Repository) {
			start := time.Now()
			foundResource := process(r, executionConfig)
			elapsed := time.Since(start)
			done = done + 1
			fmt.Printf("Processed %d of %d \t %s took %s\n", done, len(executionConfig.Repositories), r.Name, elapsed)
			collector.merge(foundResource)
			wg.Done()
			<-guard
		}(repo)
	}

	wg.Wait()

	outBytes, _ := json.MarshalIndent(collector.outputResourcesList(), "", "  ")

	fmt.Printf("Saving output to file %s\n", config.OutputFile)
	os.Remove(config.OutputFile)
	os.WriteFile(config.OutputFile, outBytes, 0644)
}

func process(repo Repository, executionConfig ExecutionConfig) []Resource {

	location := fetchRepo(repo, executionConfig)

	if slices.Contains(executionConfig.RootLike, repo.Name) {
		entries, err := os.ReadDir(location)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		nestedResources := []Resource{}
		for _, e := range entries {
			if e.IsDir() {
				nestedAppName := e.Name()
				tag := resolveAlias(nestedAppName, executionConfig.Aliases)
				nestedLocation := fmt.Sprintf("%s/%s", location, nestedAppName)
				findings := findReferences(tag, nestedLocation, executionConfig)
				nestedResources = append(nestedResources, Resource{Tag: tag, References: findings.References, Software: findings.Software})
			}

		}
		return nestedResources
	}

	tag := resolveAlias(repo.Name, executionConfig.Aliases)
	findings := findReferences(tag, location, executionConfig)

	return []Resource{{
		Tag:        tag,
		References: findings.References,
		Software:   findings.Software,
	}}
}
