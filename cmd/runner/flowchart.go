package runner

import (
	"fmt"
)

func GenerateFlowchart(resources []Resource, tag string) string {
	flowchart := "flowchart TD\n"
	for _, resource := range resources {
		source := resource.Tag
		if len(tag) != 0 && tag != source {
			continue
		}
		for dep := range resource.References {
			flowchart = flowchart + fmt.Sprintf("\t%s ---> %s\n", source, dep)
		}
	}
	return flowchart
}
