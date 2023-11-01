package runner

import (
	"fmt"
	"slices"
)

func GenerateFlowchart(resources []Resource, tag string, exclude []string) string {
	flowchart := "flowchart TD\n"
	for _, resource := range resources {
		source := resource.Tag
		if slices.Contains(exclude, source) {
			continue
		}
		for dep := range resource.References {
			if slices.Contains(exclude, dep) {
				continue
			}
			if len(tag) == 0 || (len(tag) > 0 && (source == tag || dep == tag)) {
				flowchart = flowchart + fmt.Sprintf("\t%s ---> %s\n", source, dep)
			}
		}
	}

	return flowchart
}
