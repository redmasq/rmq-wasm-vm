# rmq-wasm-vm
Go based WASM VM Library to be used by other projects

# Current badges
[![GitHub Go Coverage](https://github.com/redmasq/rmq-wasm-vm/actions/workflows/go.yml/badge.svg)](https://github.com/redmasq/rmq-wasm-vm/actions/workflows/go.yml)
[![Coveralls Coverage Status](https://coveralls.io/repos/github/redmasq/rmq-wasm-vm/badge.svg?branch=main)](https://coveralls.io/github/redmasq/rmq-wasm-vm?branch=main)


# Current plan \(abridged\)
Right now focused on the core specification [WebAssembly Specification](https://webassembly.github.io/spec/core/); however, effort is being put into retrofitting to better align with Golang conventions.
* <del>Initial Control Opcodes such as 0x01 NOP and 0x0B END</del>
* Numeric Instructions by type i32, i64, f32, f64
    - math (i32 and i64)
        - <del>const</del>
        + <del>add</del>
        + <del>sub</del>
        + <del>mul</del>
        + div_s (i64 in-progress)
        + <del>div_u</del>
        + rem
    - math (f32 and f64)
        - const
        + add
        + sub
        + mul
        + div
        + rem
        + additional float math such copysign
    - bit
        + and
        + or
        + xor
        + shl
        + shr
        + rotl
        + rotr
    - comparison
    - other
* Additional Control Instruction Opcodes
* Functions
* To be determined

JavaScript, Web, and WASI APIs will be added later once core operations are at a better point.

# Further thoughts on plan
This will exclude the code base maintenance items such as improving testing and code quality which will be done peroidically as seen fit.
1. <del>NOP and END control operations</del>
2. All integer math operations (in progress)
3. All bitwise numeric operations
4. Additional control operations along with comparison operations
5. Some either vector or table operations
    A. Lack of explicit string coverage implies that strings are handled as either vector or table
6. Implement load and store numeric operations
7. Implement function and external function operations
    A. This includes the export tables of the VM
    B. This likely will include changing startup to call _start()
8. Work on ./cmd for a bit
    A. Modify the move the ./cmd/wasmvm package to ./cmd/smoketest
    B. Create new CLI tool to load binary and start VM
9. Stop main development and have empty main() function
    A. Check what breaks
    B. Stub out, where possible, or implement, where necessary, WASI API as required
    C. Once working, go back to ./pkg development
10. Finish the remaining operations that are missing
11. Decide where to proceed (pick one)
    A. Implement ring model (basis of shared memory and threading)
    B. Modify the the main() to print "hello world" and continue more WASI work, retrofitting stubbed out methods

It is an eventual goal to make DOOM at least start in this VM, but no timeline is set for that at this point.