# Testing
## Repo Checkout
First, you'll need the repo if you do not already have it. Using the git command line utility

```
    git clone https://github.com/redmasq/rmq-wasm-vm.git
    cd rmq-wasm-vm
```


Or check out using GitHub Desktop or your favorite tool.

## Smoke Test
Currently, there is a smoke test for the CLI, to execute, run the following

```
    go run ./cmd/wasmvm
```

This likely will be moved when a real CLI tool is introduced.

## Go Unit Test Suite
Also, the normal test suite can be executed the usual fashion

```
go test "-coverprofile=coverage.out" ./...
```
