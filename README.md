This is [vault48.org](https://vault48.org) backend, written in golang.

### Installation
1. Clone this repo `git clone git@github.com:muerwre/vault-golang.git`
2. Copy `config.example.yaml`  
3. Set it up (i know, its hard)

### Running
Simply `go run main.go serve`, restart after comitting changes

### Building
Do the `make build`, then copy `./build/*` somewhere and run 
`./build/vault serve`

### Databases and migration
Gorm will handle initial migration after first launch.

### Architectural notes
```text
Model -> Repository -> DTO <--> Usecase -> Controller -> Router --> API
--------- Storage --------      ----------- Feature -----------
```
