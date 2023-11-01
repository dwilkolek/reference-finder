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
}

var wg sync.WaitGroup

type Config struct {
	ReferenceRegexp        *regexp.Regexp `json:"reg"`
	Exclude                []string       `json:"exclude"`
	RootLike               []string       `json:"rootlike"`
	Repositories           []Repository   `json:"respositories"`
	KeepWithNoDependencies bool           `json:"keepWithNoDependencies"`
	Concurrency            int16          `json:"concurrency"`
	InputFile              string         `json:"input"`
	OutputFile             string         `json:"output"`
	WorkDir                string         `json:"workdir"`
	TrimSuffix             string         `json:"trimSuffix"`
}

type collector struct {
	config    Config
	resources map[string]Resource
	lock      sync.Mutex
}

func (collector *collector) outputResourcesList() []Resource {
	v := make([]Resource, 0)

	for _, res := range collector.resources {
		if !collector.config.KeepWithNoDependencies && len(res.References) == 0 {
			continue
		}
		v = append(v, res)
	}
	return v
}

func (collector *collector) merge(newResources []Resource) {
	collector.lock.Lock()
	defer collector.lock.Unlock()

	for _, newResource := range newResources {
		resource := collector.resources[newResource.Tag]
		merged := mergeRefs(resource.References, newResource.References)
		for _, exclude := range collector.config.Exclude {
			delete(merged, exclude)
		}

		collector.resources[newResource.Tag] = Resource{
			Tag:        newResource.Tag,
			References: merged,
		}
		for _, exclude := range collector.config.Exclude {
			delete(collector.resources, exclude)
		}
	}

}

func Execute(config Config) {
	fmt.Printf("Executing with %+v\n", config)
	config.Repositories = readInputFile(config.InputFile)
	fmt.Printf("Entries to process: %d\n", len(config.Repositories))
	config.WorkDir = "workdir"
	collector := collector{
		config:    config,
		resources: map[string]Resource{},
		lock:      sync.Mutex{},
	}

	_ = os.Mkdir(config.WorkDir, os.ModePerm)

	guard := make(chan struct{}, config.Concurrency)
	done := 0

	for _, repo := range config.Repositories {
		wg.Add(1)
		guard <- struct{}{}
		go func(r Repository) {
			start := time.Now()
			foundResource := process(r, config)
			elapsed := time.Since(start)
			done = done + 1
			fmt.Printf("Processed %d of %d \t %s took %s\n", done, len(config.Repositories), r.Name, elapsed)
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

func process(repo Repository, cfg Config) []Resource {

	location := fetchRepo(repo, cfg)

	if slices.Contains(cfg.RootLike, repo.Name) {
		entries, err := os.ReadDir(location)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		nestedResources := []Resource{}
		for _, e := range entries {
			if e.IsDir() {
				nestedAppName := e.Name()
				nestedLocation := fmt.Sprintf("%s/%s", location, nestedAppName)
				dependencies := findReferences(nestedAppName, nestedLocation, cfg)
				nestedResources = append(nestedResources, Resource{Tag: nestedAppName, References: dependencies})
			}

		}
		return nestedResources
	}

	dependencies := findReferences(repo.Name, location, cfg)

	return []Resource{{
		Tag:        repo.Name,
		References: dependencies,
	}}
}
