package mu51

import (
	"errors"
)

var ErrUnknownOpCode = errors.New("usage of unknown opcode 0xA5")

type OpCode uint8

const (
	ACALL OpCode = iota
	ADD
	ADDC
	AJMP
	ANL
	CJNE
	CLR
	CPL
	DA
	DEC
	DIV
	DJNZ
	INC
	JB
	JBC
	JC
	JMP
	JNB
	JNC
	JNZ
	JZ
	LCALL
	LJMP
	MOV
	MOVC
	MOVX
	MUL
	NOP
	ORL
	POP
	PUSH
	RET
	RETI
	RL
	RLC
	RR
	RRC
	SJMP
	SETB
	SUBB
	SWAP
	XCH
	XCHD
	XRL
	ERR
)

type decoded struct {
	len uint8
	op  OpCode
}

//a jump table to quickly decode which opcode we are dealing with
//and how much more data we should read
var decodeTable = [256]decoded{
	{len: 1, op: NOP},   //0x00
	{len: 2, op: AJMP},  //0x01
	{len: 3, op: LJMP},  //0x02
	{len: 1, op: RR},    //0x03
	{len: 1, op: INC},   //0x04
	{len: 2, op: INC},   //0x05
	{len: 1, op: INC},   //0x06
	{len: 1, op: INC},   //0x07
	{len: 1, op: INC},   //0x08
	{len: 1, op: INC},   //0x09
	{len: 1, op: INC},   //0x0A
	{len: 1, op: INC},   //0x0B
	{len: 1, op: INC},   //0x0C
	{len: 1, op: INC},   //0x0D
	{len: 1, op: INC},   //0x0E
	{len: 1, op: INC},   //0x0F
	{len: 3, op: JBC},   //0x10
	{len: 2, op: ACALL}, //0x11
	{len: 3, op: LCALL}, //0x12
	{len: 1, op: RRC},   //0x13
	{len: 1, op: DEC},   //0x14
	{len: 2, op: DEC},   //0x15
	{len: 1, op: DEC},   //0x16
	{len: 1, op: DEC},   //0x17
	{len: 1, op: DEC},   //0x18
	{len: 1, op: DEC},   //0x19
	{len: 1, op: DEC},   //0x1A
	{len: 1, op: DEC},   //0x1B
	{len: 1, op: DEC},   //0x1C
	{len: 1, op: DEC},   //0x1D
	{len: 1, op: DEC},   //0x1E
	{len: 1, op: DEC},   //0x1F
	{len: 3, op: JB},    //0x20
	{len: 2, op: AJMP},  //0x21
	{len: 1, op: RET},   //0x22
	{len: 1, op: RL},    //0x23
	{len: 2, op: ADD},   //0x24
	{len: 2, op: ADD},   //0x25
	{len: 1, op: ADD},   //0x26
	{len: 1, op: ADD},   //0x27
	{len: 1, op: ADD},   //0x28
	{len: 1, op: ADD},   //0x29
	{len: 1, op: ADD},   //0x2A
	{len: 1, op: ADD},   //0x2B
	{len: 1, op: ADD},   //0x2C
	{len: 1, op: ADD},   //0x2D
	{len: 1, op: ADD},   //0x2E
	{len: 1, op: ADD},   //0x2F
	{len: 3, op: JNB},   //0x30
	{len: 2, op: ACALL}, //0x31
	{len: 1, op: RETI},  //0x32
	{len: 1, op: RLC},   //0x33
	{len: 2, op: ADDC},  //0x34
	{len: 2, op: ADDC},  //0x35
	{len: 1, op: ADDC},  //0x36
	{len: 1, op: ADDC},  //0x37
	{len: 1, op: ADDC},  //0x38
	{len: 1, op: ADDC},  //0x39
	{len: 1, op: ADDC},  //0x3A
	{len: 1, op: ADDC},  //0x3B
	{len: 1, op: ADDC},  //0x3C
	{len: 1, op: ADDC},  //0x3D
	{len: 1, op: ADDC},  //0x3E
	{len: 1, op: ADDC},  //0x3F
	{len: 2, op: JC},    //0x40
	{len: 2, op: AJMP},  //0x41
	{len: 2, op: ORL},   //0x42
	{len: 3, op: ORL},   //0x43
	{len: 2, op: ORL},   //0x44
	{len: 2, op: ORL},   //0x45
	{len: 1, op: ORL},   //0x46
	{len: 1, op: ORL},   //0x47
	{len: 1, op: ORL},   //0x48
	{len: 1, op: ORL},   //0x49
	{len: 1, op: ORL},   //0x4A
	{len: 1, op: ORL},   //0x4B
	{len: 1, op: ORL},   //0x4C
	{len: 1, op: ORL},   //0x4D
	{len: 1, op: ORL},   //0x4E
	{len: 1, op: ORL},   //0x4F
	{len: 2, op: JNC},   //0x50
	{len: 2, op: ACALL}, //0x51
	{len: 2, op: ANL},   //0x52
	{len: 3, op: ANL},   //0x53
	{len: 2, op: ANL},   //0x54
	{len: 2, op: ANL},   //0x55
	{len: 1, op: ANL},   //0x56
	{len: 1, op: ANL},   //0x57
	{len: 1, op: ANL},   //0x58
	{len: 1, op: ANL},   //0x59
	{len: 1, op: ANL},   //0x5A
	{len: 1, op: ANL},   //0x5B
	{len: 1, op: ANL},   //0x5C
	{len: 1, op: ANL},   //0x5D
	{len: 1, op: ANL},   //0x5E
	{len: 1, op: ANL},   //0x5F
	{len: 2, op: JZ},    //0x60
	{len: 2, op: AJMP},  //0x61
	{len: 2, op: XRL},   //0x62
	{len: 3, op: XRL},   //0x63
	{len: 2, op: XRL},   //0x64
	{len: 2, op: XRL},   //0x65
	{len: 1, op: XRL},   //0x66
	{len: 1, op: XRL},   //0x67
	{len: 1, op: XRL},   //0x68
	{len: 1, op: XRL},   //0x69
	{len: 1, op: XRL},   //0x6A
	{len: 1, op: XRL},   //0x6B
	{len: 1, op: XRL},   //0x6C
	{len: 1, op: XRL},   //0x6D
	{len: 1, op: XRL},   //0x6E
	{len: 1, op: XRL},   //0x6F
	{len: 2, op: JNZ},   //0x70
	{len: 2, op: ACALL}, //0x71
	{len: 2, op: ORL},   //0x72
	{len: 1, op: JMP},   //0x73
	{len: 2, op: MOV},   //0x74
	{len: 3, op: MOV},   //0x75
	{len: 2, op: MOV},   //0x76
	{len: 2, op: MOV},   //0x77
	{len: 2, op: MOV},   //0x78
	{len: 2, op: MOV},   //0x79
	{len: 2, op: MOV},   //0x7A
	{len: 2, op: MOV},   //0x7B
	{len: 2, op: MOV},   //0x7C
	{len: 2, op: MOV},   //0x7D
	{len: 2, op: MOV},   //0x7E
	{len: 2, op: MOV},   //0x7F
	{len: 2, op: SJMP},  //0x80
	{len: 2, op: AJMP},  //0x81
	{len: 2, op: ANL},   //0x82
	{len: 1, op: MOVC},  //0x83
	{len: 1, op: DIV},   //0x84
	{len: 3, op: MOV},   //0x85
	{len: 2, op: MOV},   //0x86
	{len: 2, op: MOV},   //0x87
	{len: 2, op: MOV},   //0x88
	{len: 2, op: MOV},   //0x89
	{len: 2, op: MOV},   //0x8A
	{len: 2, op: MOV},   //0x8B
	{len: 2, op: MOV},   //0x8C
	{len: 2, op: MOV},   //0x8D
	{len: 2, op: MOV},   //0x8E
	{len: 2, op: MOV},   //0x8F
	{len: 3, op: MOV},   //0x90
	{len: 2, op: ACALL}, //0x91
	{len: 2, op: MOV},   //0x92
	{len: 1, op: MOVC},  //0x93
	{len: 2, op: SUBB},  //0x94
	{len: 2, op: SUBB},  //0x95
	{len: 1, op: SUBB},  //0x96
	{len: 1, op: SUBB},  //0x97
	{len: 1, op: SUBB},  //0x98
	{len: 1, op: SUBB},  //0x99
	{len: 1, op: SUBB},  //0x9A
	{len: 1, op: SUBB},  //0x9B
	{len: 1, op: SUBB},  //0x9C
	{len: 1, op: SUBB},  //0x9D
	{len: 1, op: SUBB},  //0x9E
	{len: 1, op: SUBB},  //0x9F
	{len: 2, op: ORL},   //0xA0
	{len: 2, op: AJMP},  //0xA1
	{len: 2, op: MOV},   //0xA2
	{len: 1, op: INC},   //0xA3
	{len: 1, op: MUL},   //0xA4
	{len: 1, op: ERR},   //0xA5
	{len: 2, op: MOV},   //0xA6
	{len: 2, op: MOV},   //0xA7
	{len: 2, op: MOV},   //0xA8
	{len: 2, op: MOV},   //0xA9
	{len: 2, op: MOV},   //0xAA
	{len: 2, op: MOV},   //0xAB
	{len: 2, op: MOV},   //0xAC
	{len: 2, op: MOV},   //0xAD
	{len: 2, op: MOV},   //0xAE
	{len: 2, op: MOV},   //0xAF
	{len: 2, op: ANL},   //0xB0
	{len: 2, op: ACALL}, //0xB1
	{len: 2, op: CPL},   //0xB2
	{len: 1, op: CPL},   //0xB3
	{len: 3, op: CJNE},  //0xB4
	{len: 3, op: CJNE},  //0xB5
	{len: 3, op: CJNE},  //0xB6
	{len: 3, op: CJNE},  //0xB7
	{len: 3, op: CJNE},  //0xB8
	{len: 3, op: CJNE},  //0xB9
	{len: 3, op: CJNE},  //0xBA
	{len: 3, op: CJNE},  //0xBB
	{len: 3, op: CJNE},  //0xBC
	{len: 3, op: CJNE},  //0xBD
	{len: 3, op: CJNE},  //0xBE
	{len: 3, op: CJNE},  //0xBF
	{len: 2, op: PUSH},  //0xC0
	{len: 2, op: AJMP},  //0xC1
	{len: 2, op: CLR},   //0xC2
	{len: 1, op: CLR},   //0xC3
	{len: 1, op: SWAP},  //0xC4
	{len: 2, op: XCH},   //0xC5
	{len: 1, op: XCH},   //0xC6
	{len: 1, op: XCH},   //0xC7
	{len: 1, op: XCH},   //0xC8
	{len: 1, op: XCH},   //0xC9
	{len: 1, op: XCH},   //0xCA
	{len: 1, op: XCH},   //0xCB
	{len: 1, op: XCH},   //0xCC
	{len: 1, op: XCH},   //0xCD
	{len: 1, op: XCH},   //0xCE
	{len: 1, op: XCH},   //0xCF
	{len: 2, op: POP},   //0xD0
	{len: 2, op: ACALL}, //0xD1
	{len: 2, op: SETB},  //0xD2
	{len: 1, op: SETB},  //0xD3
	{len: 1, op: DA},    //0xD4
	{len: 3, op: DJNZ},  //0xD5
	{len: 1, op: XCHD},  //0xD6
	{len: 1, op: XCHD},  //0xD7
	{len: 2, op: DJNZ},  //0xD8
	{len: 2, op: DJNZ},  //0xD9
	{len: 2, op: DJNZ},  //0xDA
	{len: 2, op: DJNZ},  //0xDB
	{len: 2, op: DJNZ},  //0xDC
	{len: 2, op: DJNZ},  //0xDD
	{len: 2, op: DJNZ},  //0xDE
	{len: 2, op: DJNZ},  //0xDF
	{len: 1, op: MOVX},  //0xE0
	{len: 2, op: AJMP},  //0xE1
	{len: 1, op: MOVX},  //0xE2
	{len: 1, op: MOVX},  //0xE3
	{len: 1, op: CLR},   //0xE4
	{len: 2, op: MOV},   //0xE5
	{len: 1, op: MOV},   //0xE6
	{len: 1, op: MOV},   //0xE7
	{len: 1, op: MOV},   //0xE8
	{len: 1, op: MOV},   //0xE9
	{len: 1, op: MOV},   //0xEA
	{len: 1, op: MOV},   //0xEB
	{len: 1, op: MOV},   //0xEC
	{len: 1, op: MOV},   //0xED
	{len: 1, op: MOV},   //0xEE
	{len: 1, op: MOV},   //0xEF
	{len: 1, op: MOVX},  //0xF0
	{len: 2, op: ACALL}, //0xF1
	{len: 1, op: MOVX},  //0xF2
	{len: 1, op: MOVX},  //0xF3
	{len: 1, op: CPL},   //0xF4
	{len: 2, op: MOV},   //0xF5
	{len: 1, op: MOV},   //0xF6
	{len: 1, op: MOV},   //0xF7
	{len: 1, op: MOV},   //0xF8
	{len: 1, op: MOV},   //0xF9
	{len: 1, op: MOV},   //0xFA
	{len: 1, op: MOV},   //0xFB
	{len: 1, op: MOV},   //0xFC
	{len: 1, op: MOV},   //0xFD
	{len: 1, op: MOV},   //0xFE
	{len: 1, op: MOV},   //0xFF
}

func ReadInstruction(c CodeMemory, offset int64) (OpCode, []byte, error) {

	//max instruction size is 3
	b := make([]byte, 3)

	//suck in the first byte which contains the opcode
	c.ReadAt(b[:1], offset) //TODO: handle read errors

	//decode the opcode and instruction length
	d := decodeTable[b[0]]

	//read in remainder of data if multibyte instruction
	if d.len > 1 {
		c.ReadAt(b[1:d.len-1], offset) //TODO: handle read errors
	}

	return d.op, b[:d.len], nil
}

type Operation func(*CPU) error

//convert opcode and extra data pair into a permutation function
func DecodeInstruction(op OpCode, data []byte) Operation {
	switch op {
	case ANL:
	case CJNE:
	case CLR:
	case CPL:
	case DA:
	case DEC:
	case DIV:
	case DJNZ:
	case INC:
	case JB:
	case JBC:
	case JC:
	case JMP:
	case JNB:
	case JNC:
	case JNZ:
	case JZ:
	case MOV:
	case MOVC:
	case MOVX:
	case NOP:
	case ORL:
	case RL:
	case RLC:
	case RR:
	case RRC:
	case SJMP:
	case SETB:
	case SUBB:
	case XCHD:
	case XRL:
	}
	return nil
}
