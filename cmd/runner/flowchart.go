package runner

import (
	"fmt"
)

func GenerateFlowchart(resources []Resource, tag string) string {
	flowchart := "flowchart TD\n"
	for _, resource := range resources {
		source := resource.Tag

		for dep := range resource.References {
			if len(tag) == 0 || (len(tag) > 0 && (source == tag || dep == tag)) {
				flowchart = flowchart + fmt.Sprintf("\t%s ---> %s\n", source, dep)
			}
		}
	}

	return flowchart
}
