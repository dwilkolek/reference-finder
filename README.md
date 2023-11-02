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
  reference-finder analyze-json [flags]

Flags:
  -i, --config string   Config file (default "config.json")
  -h, --help            help for analyze-json  
 </pre>


## Flowchart generator 
Generates file to render [Mermaid](https://mermaid.live/) chart.
<pre>
Usage:
  reference-finder flowchart [flags]

Flags:
  -e, --exclude strings            Exclude from chart
  -g, --group-definitions string   Group definitions specification
  -h, --help                       help for flowchart
      --include-orphans            Include orphan center
  -i, --input string               Input file (default "output.json")
  -o, --output string              Output file (default "flowchart.txt")
  -r, --resource string            Chart for single resource
</pre>
## Reguirements

- Configured github cli

## Prepare input file

 `gh repo list <organisation|user> -L 1000 --no-archived --json name,url > input.json`

## Exmaples

- Generate output file

```
go run main.go analize --reg "(?:http|https)://([a-zA-Z0-9-]+)(?:.dev|.demo){0,1}.service" --rootlike "environment-variables" --input gh.json --remove-entries-without-dependencies-from-output true -c 500
```

- Generate output file using config json as input

```
go run main.go analize-json -i config.json
```

- Generate [Mermaid](https://mermaid.live/) flowchart.txt

```
go run main.go flowchart -i output.json -o flowchart.txt
```

- Generate [Mermaid](https://mermaid.live/) flowchart.txt for single resource

```
go run main.go flowchart -r resource-1
```

- Generate [Mermaid](https://mermaid.live/) flowchart.txt excluding some resources and groupping others with orgphans

```
go run main.go flowchart -e "resource-2,some-other-4" -g groups.json --include-orphans
```