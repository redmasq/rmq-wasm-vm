package wasmvm

type Instruction func(*VMState) error

const (
	// Control Instructions
	OP_NOP = 0x01
	OP_END = 0x0B

	// Numeric instructions
	OP_CONST_I32 = 0x41
	OP_CONST_I64 = 0x42

	// Numeric i32 arithmatic instructions
	OP_ADD_I32  = 0x6A
	OP_SUB_I32  = 0x6B
	OP_MUL_I32  = 0x6C
	OP_DIVU_I32 = 0x6E

	// Numeric I64 arithmatic instructions
	OP_ADD_I64  = 0x7C
	OP_SUB_I64  = 0x7D
	OP_MUL_I64  = 0x7E
	OP_DIVU_I64 = 0x80
)

func defaultInstructionMap() map[uint8]Instruction {
	return map[uint8]Instruction{
		OP_NOP:       NOP,
		OP_END:       END, // End of function
		OP_CONST_I32: CONST_I32,
		OP_CONST_I64: CONST_I64,
		OP_ADD_I32:   ADD_I32,
		OP_SUB_I32:   SUB_I32,
		OP_MUL_I32:   MUL_I32,
		OP_DIVU_I32:  DIVU_I32,
		OP_ADD_I64:   ADD_I64,
		OP_SUB_I64:   SUB_I64,
		OP_MUL_I64:   MUL_I64,
		OP_DIVU_I64:  DIVU_I64,
	}
}
