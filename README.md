# rmq-wasm-vm
Go based WASM VM Library to be used by other projects

# Current plan \(abridged\)
Right now focused on the core specification [WebAssembly Specification](https://webassembly.github.io/spec/core/)
* <del>Initial Control Opcodes such as 0x01 NOP and 0x0B END</del>
* Numeric Instructions by type i32, i64, f32, f64
    - const
    - math (i32 and i64)
        + <del>add</del>
        + <del>sub</del>
        + <del>mul</del>
        + div
        + rem
    - math (f32 and f64)
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
