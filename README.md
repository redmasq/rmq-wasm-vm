# rmq-wasm-vm
**rmq-wasm-vm** is a Go based WASM VM Library to be used by other projects. The general design is based in interpretation given that JIT would require usage of assembly or cgo; if JIT is added, it would likely be done as a fork. By the eventual version 1.0, it is intended to be a modular library with reasonable separation of concerns, with focus on correctness and security, being flexible where appropriate.

## Current badges
[![GitHub Go Coverage](https://github.com/redmasq/rmq-wasm-vm/actions/workflows/go.yml/badge.svg)](https://github.com/redmasq/rmq-wasm-vm/actions/workflows/go.yml)
[![Coveralls Coverage Status](https://coveralls.io/repos/github/redmasq/rmq-wasm-vm/badge.svg?branch=main)](https://coveralls.io/github/redmasq/rmq-wasm-vm?branch=main)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

## Current plan \(abridged\)
Right now focused on the core specification [WebAssembly Specification](https://webassembly.github.io/spec/core/); however, effort is being put into retrofitting to better align with Golang conventions.

Excluding the aforementioned base maintenance items, which will be done periodically as seen fit, I have this as the overall plan at the moment.
1. <del>NOP and END control operations</del>
2. All integer math operations (in progress)
    * math (i32 and i64)
        - <del>const</del>
        - <del>add</del>
        - <del>sub</del>
        - <del>mul</del>
        - div_s (i64 in-progress)
        - <del>div_u</del>
        - rem
3. All bitwise numeric operations
4. Additional control operations along with comparison operations
5. Some either vector or table operations
    * Lack of explicit string coverage implies that strings are handled as either vector or table
6. Implement load and store numeric operations
7. Implement function and external function operations
    * This includes the export tables of the VM
    * This likely will include changing startup to call _start()
8. Work on ./cmd for a bit
    * Move the ./cmd/wasmvm package to ./cmd/smoketest
    * Create new CLI tool to load binary and start VM
9. Stop main development and have empty main() function
    * Check what breaks
    * Stub out, where possible, or implement, where necessary, WASI API as required
    * Once working, go back to ./pkg development
10. Finish the remaining operations that are missing
11. Decide where to proceed (pick one)
    * Implement ring model (basis of shared memory and threading)
    * Modify the the main() to print "hello world" and continue more WASI work, retrofitting stubbed out methods

It is an eventual goal to make DOOM at least start in this VM, but no timeline is set for that at this point.

## Codebase
* A smoketest utility resides under ./cmd/wasmvm
    - Use go run ./cmd/wasmvm to run
* The library itself resides under ./pkg/wasmvm
    - As a side note, try go test "-coverprofile=coverage.out" ./... to run all tests \(includes the ./cmd\)
* Initial written documentation is under ./doc
* Under ./scripts is convert_docbook.sh that converts the docbooks under ./docs/specs from docbook to html under ./docs/output
    - It requires a tool such as pandoc and the docbook.xsl (xslt) files.
    - See comments and code in script for further detail

## License and Contributing
Please see the following files
* [Contributor License Agreement (CLA)](CLA.md)
* [Contributing Guidelines](CONTRIBUTING.md)
* [License](LICENSE)

Overall the project uses Apache 2.0 license, but take special note of the conditions of the CLA.

One way to contribute without code would be to report bugs. If a feature is reported as implemented and has a bug, an issue can be created at the [Github Issues Page for the Project](https://github.com/redmasq/rmq-wasm-vm/issues). If you have code in which to contribute, please consult the [CLA](CLA.md) prior to submitting a [pull request](https://github.com/redmasq/rmq-wasm-vm/pulls). Please note that minimal overall code coverage is 85% as defined by CI/CD configuration; however, any pull requests reviewed should have all major branches covered with meaningful tests. The project maintainers reserve the right to reject pull requests for any reason.

## Related
* [Wikipedia: WebAssembly](https://en.wikipedia.org/wiki/WebAssembly)
* [Wikipedia: Virtual Machine](https://en.wikipedia.org/wiki/Virtual_machine)
* [Wikipedia: Finite-state Machine](https://en.wikipedia.org/wiki/Finite-state_machine)

## Additional Note
This project is internally managed by a private Jira board. If the project grows, opening the board or migrating to another tracking system will be considered at that time.