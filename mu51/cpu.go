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

const (
	ParityBit uint8 = (0x01 << iota)
	UserBit
	OverFlowBit
	RS0Bit
	RS1Bit
	FlagBit
	AuxCarryBit
	CarryBit
)

type ALU struct {
	InternalRAM [256]byte //internal IRAM (is also the registers)

	StackPtr   byte   //Stack Pointer (0x81)
	DataPtr    uint16 //16 bit data pointer register (0x82-0x83)
	ProgStatus byte   //Program status word (0xD0)
	Accum      byte   //accumulator (0xE0)
	BReg       byte   //B register (0xF0)
	ProgCount  uint16 //the program counter
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

func (a *ALU) InstrACALL(rel uint16) {
	a.PushWord(a.ProgCount + 2)
	a.ProgCount = (a.ProgCount+2)&0xF800 + rel&0x7FF
}

func (a *ALU) InstrAJMP(rel uint16) {
	a.ProgCount = (a.ProgCount+2)&0xF800 + rel&0x07FF
}

func (a *ALU) InstrRET() {
	a.ProgCount = a.PopWord()
}

func (a *ALU) InstrRETI() {
	//no interrupts right now
	a.ProgCount = a.PopWord()
}

func (a *ALU) InstrPUSH(loc uint8) {
	a.PushByte(a.InternalRAM[loc])
}

func (a *ALU) InstrPOP(loc uint8) {
	a.InternalRAM[loc] = a.PopByte()
}

func (a *ALU) InstrSWAP() {
	a.Accum = (a.Accum >> 4) & (a.Accum << 4)
}

func (a *ALU) InstrMUL() {
	var res = uint16(a.Accum) * uint16(a.BReg)
	a.BReg = uint8(res >> 16)
	a.Accum = uint8(res)
	a.ProgStatus = a.ProgStatus &^ CarryBit
	if res > 255 {
		a.ProgStatus = a.ProgStatus | OverFlowBit
	} else {
		a.ProgStatus = a.ProgStatus &^ OverFlowBit
	}
}

//retrieves the current indirect address
//of a register by index, takes into account
//program status word
func (a *ALU) regAddr(r uint8) uint8 {
	return ((a.ProgStatus >> 3) & 0x03) * 8
}

func (a *ALU) InstrXCHInd(R1 bool) {
	var regLoc uint8
	var loc uint8
	if R1 {
		regLoc = a.regAddr(1)
	} else {
		regLoc = a.regAddr(0)
	}
	loc = a.InternalRAM[regLoc]
	a.Accum, a.InternalRAM[loc] = a.InternalRAM[loc], a.Accum
}

func (a *ALU) InstrXCHDir(reg uint8) {
	var loc = a.regAddr(reg)
	a.Accum, a.InternalRAM[loc] = a.InternalRAM[loc], a.Accum
}

func (a *ALU) InstrXCHAddr(loc uint8) {
	a.Accum, a.InternalRAM[loc] = a.InternalRAM[loc], a.Accum
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
