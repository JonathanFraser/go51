package mu51

//RAM is a random access interface nessecary to act as
//backing ram for the CPU
type RAM interface {
	ReadAt([]byte, int64) (int, error)
	WriteAt([]byte, int64) (int, error)
	Size() int64
}

//CodeMemory is a read-only interface nessecary to act as
//a code storage for the CPU
type CodeMemory interface {
	ReadAt([]byte, int64) (int, error)
	Size() int64
}

const (
	//ParityBit is a mask for the Parity bit in the PSW
	ParityBit uint8 = (0x01 << iota)

	//UserBit is a mask fro the User bit in the PSW
	UserBit

	//OverFlowBit is a mask for the Overflow bit in the PSW
	OverFlowBit

	//RS0Bit is a mask for the lower bit of the register bank select in the PSW
	RS0Bit

	//RS1Bit is a mask for the upper bit of the register bank select in the PSW
	RS1Bit

	//FlagBit is a mask for the user available Flag bit in the PSW
	FlagBit

	//AuxCarryBit is a mask for the Auxilery carry bit used in some arithmatic instructions
	AuxCarryBit

	//CarryBit is a mask for the Carry bit used in arithmatic instructions
	CarryBit
)

//ALU is the data set for an 8051 core, the main registers R0-R1
//are contained within InternalRAM while the other special registers
//are tracked seperately
type ALU struct {
	InternalRAM [256]byte //internal IRAM (is also the registers)

	StackPtr   byte   //Stack Pointer (0x81)
	DataPtr    uint16 //16 bit data pointer register (0x82-0x83)
	ProgStatus byte   //Program status word (0xD0)
	Accum      byte   //accumulator (0xE0)
	BReg       byte   //B register (0xF0)
	ProgCount  uint16 //the program counter
}

//PushByte pushes a single byte on the stack
//maintaining increment ordering
func (a *ALU) PushByte(byte uint8) {
	a.StackPtr++
	a.InternalRAM[a.StackPtr] = byte
}

//PopByte removes a single byte from the stack
//maintaining decrement ordering
func (a *ALU) PopByte() byte {
	b := a.InternalRAM[a.StackPtr]
	a.StackPtr--
	return b
}

//PushWord is a helper function to place a word
//on the stack, maintaining both byte and
//increment ordering
func (a *ALU) PushWord(word uint16) {
	a.InternalRAM[a.StackPtr+1] = uint8(word)
	a.InternalRAM[a.StackPtr+2] = uint8(word >> 8)
	a.StackPtr += 2
}

//PopWord is a helper function to remove a 16bit
//word from the stack, maintaining both byte and
//decrement ordering
func (a *ALU) PopWord() uint16 {
	ret := uint16(a.InternalRAM[a.StackPtr]) << 8
	ret = ret | uint16(a.InternalRAM[a.StackPtr-1])
	a.StackPtr -= 2
	return ret
}

//InstrACALL exectues an absolute call within the same
//2KiB page as the following instruction, rel specifies
//the absolute address within that page
func (a *ALU) InstrACALL(rel uint16) {
	a.PushWord(a.ProgCount + 2)
	a.ProgCount = (a.ProgCount+2)&0xF800 + rel&0x7FF
}

//InstrAJMP executes an absolute jump within the same
//2KiB page as the following instruction, rel specifies
//the absolute address within that page
func (a *ALU) InstrAJMP(rel uint16) {
	a.ProgCount = (a.ProgCount+2)&0xF800 + rel&0x07FF
}

//InstrRET exectutes a return by restoring the
//program counter from the value saved on the stack
func (a *ALU) InstrRET() {
	a.ProgCount = a.PopWord()
}

//InstrRETI executes a return from interrupt
//it acts like a RET instruction but also restores
//lower priority interrupts
func (a *ALU) InstrRETI() {
	//no interrupts right now
	a.ProgCount = a.PopWord()
}

//InstrPUSH pushes a byte onto the stack located
//at the specified internal ram location
func (a *ALU) InstrPUSH(loc uint8) {
	a.PushByte(a.InternalRAM[loc])
}

//InstrPOP pops a byte off the stack and places it in
//the specified internal ram location
func (a *ALU) InstrPOP(loc uint8) {
	a.InternalRAM[loc] = a.PopByte()
}

//InstrSWAP executes a nibble swap in the accumulator
func (a *ALU) InstrSWAP() {
	a.Accum = (a.Accum >> 4) & (a.Accum << 4)
}

//InstrMUL executes a multiply instruction
//it multiplies the A and B registers together
//placing the results into B:A. the carry bit
//is cleared and the overflow bit is set if the
//result is larger than 255
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

//InstrXCHInd executes and exchange between
//the accumulator and an address specified with
//indirect addressing. Setting R1 will use the R1
//register for the address, else it will look to R0
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

//InstrXCHDir exectutes an exchange with direct addressing
//swaps the accumulator with the specified register
func (a *ALU) InstrXCHDir(reg uint8) {
	var loc = a.regAddr(reg)
	a.Accum, a.InternalRAM[loc] = a.InternalRAM[loc], a.Accum
}

//InstrXCHImmed performs an exchange between the
//accumulator and the specified address in internal ram
func (a *ALU) InstrXCHImmed(loc uint8) {
	a.Accum, a.InternalRAM[loc] = a.InternalRAM[loc], a.Accum
}

//InstrLCALL executes a long call
//pushes location of next instruction on the stack
//then sets program counter to specified address
func (a *ALU) InstrLCALL(addr uint16) {
	a.PushWord(a.ProgCount + 3)
	a.ProgCount = addr
}

//InstrLJMP executes a Long jump
//unconditionally sets program counter to address
func (a *ALU) InstrLJMP(addr uint16) {
	a.ProgCount = addr
}

func (a *ALU) InstrADDLit(literal uint8) {

}

func (a *ALU) InstrADDImmed(addr uint8) {
	var val = a.InternalRAM[addr]
}

func (a *ALU) InstrADDInd(R1 bool) {
	var addr uint8
	//lookup the current memory location
	//of R0 or R1
	if R1 {
		addr = a.regAddr(1)
	} else {
		addr = a.regAddr(0)
	}
	//use the value at that address as an
	//address
	var val = a.InternalRAM[a.InternalRAM[addr]]
}

func (a *ALU) InstrADDDir(reg uint8) {
	var val = a.InternalRAM[a.regAddr(reg)]
}

func (a *ALU) InstrADDCLit(literal uint8) {

}

func (a *ALU) InstrADDCImmed(addr uint8) {
	var val = a.InternalRAM[addr]
}

func (a *ALU) InstrADDCInd(R1 bool) {
	var addr uint8
	//lookup the current memory location
	//of R0 or R1
	if R1 {
		addr = a.regAddr(1)
	} else {
		addr = a.regAddr(0)
	}
	//use the value at that address as an
	//address
	var val = a.InternalRAM[a.InternalRAM[addr]]
}

func (a *ALU) InstrADDCDir(reg uint8) {
	var val = a.InternalRAM[a.regAddr(reg)]
}

//CPU is a structure representing a complete 8051 CPU, including
//code memory and external ram as well as access to special function
//registers
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
