package runner

import (
	"fmt"
	"slices"
)

func GenerateFlowchart(resources []Resource, tag string, exclude []string, groups map[string][]string, renderOrphans bool,
	validTags []string, tmap map[string]string) string {
	flowchart := "flowchart LR\n"

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
		if len(validTags) > 0 && !slices.Contains(validTags, source) {
			continue
		}
		for groupName, group := range groups {
			if slices.Contains(group, source) {
				groupped[groupName] = append(groupped[groupName], fmt.Sprintf("%s ---> %s\n", team(groupName), tr(source, tmap)))
				break
			}
		}

		for dep := range resource.References {
			if slices.Contains(exclude, dep) {
				continue
			}
			if len(validTags) > 0 && !slices.Contains(validTags, dep) {
				continue
			}
			entry := fmt.Sprintf("%s ---> %s\n", tr(source, tmap), tr(dep, tmap))
			visited[source] = true
			visited[dep] = true

			if len(tag) == 0 || (len(tag) > 0 && (source == tag || dep == tag)) {
				added := false
				for groupName, group := range groups {
					if slices.Contains(group, dep) {
						groupped[groupName] = append(groupped[groupName], fmt.Sprintf("%s ---> %s\n", team(groupName), tr(dep, tmap)))

					}
					if slices.Contains(group, source) && slices.Contains(group, dep) {
						added = true
						groupped[groupName] = append(groupped[groupName], tr(entry, tmap))
						break
					}
				}
				if !added {
					withoutGroup = append(withoutGroup, entry)
				}
			}
		}
	}

	for groupName := range groups {
		entries := unique(groupped[groupName])
		// subgraph subgraph1
		// 	direction TB
		// 	top1[top] --> bottom1[bottom]
		// end
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
				flowchart += fmt.Sprintf("\t\t%s ---> %s\n", tr(child, tmap), tr(orphanCenter, tmap))
			}
		}
		flowchart = flowchart + "\tend\n"
	}

	return flowchart
}

func team(teamName string) string {
	return fmt.Sprintf("team-%s(\"`🧑‍💻 %s 👩‍💻`\")", teamName, teamName)
}

func tr(tag string, translationMapping map[string]string) string {
	v, ok := translationMapping[tag]
	if ok {
		return fmt.Sprintf("%s(\"`%s`\")", tag, v)
	} else {
		return tag
	}
}
