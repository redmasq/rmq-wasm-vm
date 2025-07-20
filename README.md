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
