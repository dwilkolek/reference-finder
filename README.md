# Reference Finder

Written in go. It goes through specified repos in input.json and generates output.json.
Output.json contains found resources and references to other resources. What is resource? 

- Repository name (expect rootlike repos) 
- Directory in main rootlike repo 
- First capture grup in regex

## Reguirements

- Configured github cli

## Prepare input file

 `gh repo list <organisation|user> -L 1000 --no-archived --json name,url > input.json`

## Exmaples

- Generate output file

```
go run main.go analize --reg "(?:http|https)://([a-zA-Z0-9-]+)(?:.dev|.demo){0,1}.service" --rootlike "environment-variables" --exclude "gandalf,dsi-service-config,environment-variables,dxp,login-service,access-management,api-manager,api-login" --input gh.json --remove-entries-without-dependencies-from-output true -c 500
```

- Generate output file using config json as input

```
go run main.go analize-json -i config.json
```

- Generate flowchart.txt

```
go run main.go flowchart -i output.json -o flowchart.txt
```
