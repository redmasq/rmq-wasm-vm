package wasmvm

type Instruction func(*VMState) error

func defaultInstructionMap() map[uint8]Instruction {
	return map[uint8]Instruction{
		0x01: NOP,
		0x0B: END, // End of function
		0x41: CONST_I32,
		0x42: CONST_I64,
		0x6A: ADD_I32,
		0x6B: SUB_I32,
		0x6C: MUL_I32,
		0x7C: ADD_I64,
		0x7D: SUB_I64,
		0x7E: MUL_I64,
	}
}
