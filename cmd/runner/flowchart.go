package runner

import (
	"fmt"
	"slices"
)

func GenerateFlowchart(resources []Resource, tag string, exclude []string, groups map[string][]string, renderOrphans bool) string {
	flowchart := "flowchart TD\n"

	visited := map[string]bool{}

	withoutGroup := []string{}
	groupped := map[string][]string{}
	for _, resource := range resources {
		source := resource.Tag
		visited[source] = false
	}
	for _, resource := range resources {
		source := resource.Tag
		if slices.Contains(exclude, source) {
			continue
		}
		for groupName, group := range groups {
			if slices.Contains(group, source) {
				groupped[groupName] = append(groupped[groupName], fmt.Sprintf("TEAM-%s ---> %s\n", groupName, source))
				break
			}
		}

		for dep := range resource.References {
			if slices.Contains(exclude, dep) {
				continue
			}
			entry := fmt.Sprintf("%s ---> %s\n", source, dep)
			visited[source] = true
			visited[dep] = true

			if len(tag) == 0 || (len(tag) > 0 && (source == tag || dep == tag)) {
				added := false
				for groupName, group := range groups {
					if slices.Contains(group, dep) {
						groupped[groupName] = append(groupped[groupName], fmt.Sprintf("TEAM-%s ---> %s\n", groupName, dep))

					}
					if slices.Contains(group, source) && slices.Contains(group, dep) {
						added = true
						groupped[groupName] = append(groupped[groupName], entry)
						break
					}
				}
				if !added {
					withoutGroup = append(withoutGroup, entry)
				}
			}
		}
	}

	for groupName, entries := range groupped {
		// subgraph subgraph1
		// 	direction TB
		// 	top1[top] --> bottom1[bottom]
		// end
		entries := unique(entries)
		flowchart = flowchart + fmt.Sprintf("\tsubgraph %s\n", groupName)
		for _, entry := range entries {
			flowchart = flowchart + "\t\t" + entry
		}
		flowchart = flowchart + "\tend\n"
	}

	for _, entry := range withoutGroup {
		flowchart = flowchart + "\t" + entry
	}

	for child, isVisited := range visited {
		if !isVisited && !slices.Contains(exclude, child) {
			fmt.Printf("Orphan found: %s\n", child)
		}
	}

	if renderOrphans {
		orhpanGroupName := "Orphans"
		orphanCenter := "c(Orphan Center)"
		flowchart = flowchart + fmt.Sprintf("\tsubgraph %s\n", orhpanGroupName)
		for child, isVisited := range visited {
			if !isVisited && !slices.Contains(exclude, child) {
				flowchart += fmt.Sprintf("\t\t%s ---> %s\n", child, orphanCenter)
			}
		}
		flowchart = flowchart + "\tend\n"
	}

	return flowchart
}
