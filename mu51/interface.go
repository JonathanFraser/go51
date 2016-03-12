//Package mu51 implements the central processing core of an 8051 processor
package mu51

//CodeMemory is a read-only interface which Processor
//requires to pull its opcodes from
type CodeMemory interface {
	Read(p []byte) (n int, err error)
	Size() int64
}

//RAMemory is a full read-write interface which
//acts as random access memory for the Processor
type RAMemory interface {
	WriteAt(p []byte, off int64) (n int, err error)
	Read(p []byte) (n int, err error)
	Size() int64
}

//Processor is the generic interface supported by any 8051 compatible core
//it is implimented by the various varients
type Processor interface {
	//specify a function to be called when the cpu core
	//writes to a cpu port
	SetOutputCallback(uint8, func(uint8))

	//specify a function to be called when the cpu core
	//attempts to read from a cpu port
	SetInputCallback(uint8, func() uint8)

	//sets the backed to use as a memory for the processing core
	SetMemory(RAMemory)

	//sets the memory to be read for the cpu core
	SetCode(CodeMemory)

	//reset all the cpu registers to their power on values
	Reset()

	//increment by a single clock step (this may become instruction)
	Step()
}
