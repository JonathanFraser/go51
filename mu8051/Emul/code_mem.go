package Emul

import "errors"
import "github.com/binaryblade/go/hex"

type Memory map[uint16]byte

func New() (*Memory) {
	ret := make(map[uint16]byte)
	return (*Memory)(&ret)
}

func NewHex(memFile hex.File) (*Memory,error) {
	ret := make(map[uint16]byte)

	for k,v := range memFile {	
		switch v.Type {
			case hex.Data:
				for off,data := range v.Data {
					ret[v.Address+uint16(off)] = data
				}
			case hex.EoF:
				if k != len(memFile)-1 {
					return (*Memory)(&ret), errors.New("end of file occured early")
				}
			default:
				//don't do anything for unsupported memory commands
		}
	}
	
	return (*Memory)(&ret), errors.New("No EOF in file")
}


func (m *Memory) Read(addr uint16) byte {
	return (*m)[addr]
}

func (m *Memory) Write(addr uint16, data byte) {
	(*m)[addr] = data
}
