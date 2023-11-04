# Reference Finder

Written in go. It goes through specified repos in input.json and generates output.json.
Output.json contains found resources and references to other resources. What is resource? 

- Repository name (expect rootlike repos) 
- Directory in main rootlike repo 
- First capture grup in regex

## Analyzer
Purpose is to map repositories into single json mapping that could be used to render flowchart.

<pre>
Usage:
  reference-finder analyze [flags]

Flags:
  -i, --config string   Config file (default "config.json")
  -h, --help            help for analyze
 </pre>

```
type Config struct {
	ReferenceRegexp *regexp.Regexp `json:"reg"`
	RootLike        []string       `json:"rootlike"`
	Concurrency     int16          `json:"concurrency"`
	InputFile       string         `json:"input"`
	OutputFile      string         `json:"output"`
	TrimSuffix      string         `json:"trimSuffix"`
	Sync            bool           `json:"sync"`
	ExtendedSearch  bool           `json:"extendedSearch"`
}
```

## Flowchart generator 
Generates file to render [Mermaid](https://mermaid.live/) chart.
<pre>
Usage:
  reference-finder flowchart [flags]

Flags:
  -e, --exclude string             Exclude from chart
  -g, --group-definitions string   Group definitions specification
  -h, --help                       help for flowchart
      --include-orphans            Include orphan center
  -i, --input string               Input file (default "output.json")
  -o, --output string              Output file (default "flowchart.txt")
  -r, --resource string            Chart for single resource
  -t, --translation string         Mapping tags to display names. One line - one translation. Separated by ;.
  -v, --valid-tags string          List of valid tags
</pre>
## Reguirements

- Configured github cli

## Prepare input file

 `gh repo list <organisation|user> -L 1000 --no-archived --json name,url > input.json`
