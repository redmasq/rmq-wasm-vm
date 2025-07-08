package wasmvm

type Instruction func(*VMState) error

func defaultInstructionMap() map[uint8]Instruction {
	return map[uint8]Instruction{
		0x01: NOP,
		0x0B: END, // End of function
		0x43: CONST_I32,
		0x6A: ADD_I32,
		0x6B: SUB_I32,
	}
}
