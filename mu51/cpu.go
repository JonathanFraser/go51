package mu51

type RAM interface {
	ReadAt([]byte, int64) (int, error)
	WriteAt([]byte, int64) (int, error)
	Size() int64
}

type CodeMemory interface {
	ReadAt([]byte, int64) (int, error)
	Size() int64
}

type ALU struct {
	InternalRAM [256]byte //internal IRAM (is also the registers)

	StackPtr     byte   //Stack Pointer (0x81)
	DataPtr      uint16 //16 bit data pointer register (0x82-0x83)
	ProgStatWord byte   //Program status word (0xD0)
	Accum        byte   //accumulator (0xE0)
	BReg         byte   //B register (0xF0)
	ProgCount    uint16 //the program counter
}

func (a *ALU) PushByte(byte uint8) {
	a.StackPtr++
	a.InternalRAM[a.StackPtr] = byte
}

func (a *ALU) PopByte() byte {
	b := a.InternalRAM[a.StackPtr]
	a.StackPtr--
	return b
}

func (a *ALU) PushWord(word uint16) {
	a.InternalRAM[a.StackPtr+1] = uint8(word)
	a.InternalRAM[a.StackPtr+2] = uint8(word >> 8)
	a.StackPtr += 2
}

func (a *ALU) PopWord() uint16 {
	ret := uint16(a.InternalRAM[a.StackPtr]) << 8
	ret = ret | uint16(a.InternalRAM[a.StackPtr-1])
	a.StackPtr -= 2
	return ret
}

type CPU struct {
	*ALU //all the cpu registers

	//callbacks will not be triggered when accessing special CPU registers
	//Read callbacks for when accessing SFR memory space
	ReadCallbacks [256]func() uint8
	//Write callbacks for when accessing SFR memory space
	WriteCallbacks [256]func(uint8)
	ExtRAM         RAM
	ProgMem        CodeMemory
}

func (c *CPU) Step() error {
	return nil
}
