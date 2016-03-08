//Package intelhex provides tools for parsing the intel hex file format
//as well as have it appear as a contiguous file, despite its sparse nature
package ihex

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sort"
)

//record did not contain the minimum of 11 bytes on the line
//there is not point in parsing further
var ErrInsufficentRecordLength = errors.New("record of insufficient length")

//did not find a colon as the first character on the line
var ErrNoStartCode = errors.New("line not prefixed with start code ':'")

//decoding the hex string resulted in an abnormal length byte slice
var ErrUnexpectedDecodeLength = errors.New("decoded hex string at unexpected length")

//line checksum did not match that present in record
var ErrChecksum = errors.New("checksum invalid")

//file did not include an EOF record
var ErrNoEOF = errors.New("failed to locate EOF record")

//EOF record found on line other than the last line
var ErrUnexpectedEOF = errors.New("encountered EOF on line other than the last")

//There are more bytes in the data record than the length field specified
var ErrExtraBytes = errors.New("record contained extra data bytes")

//An internal precondition was violated, this is likely the result
//of a code bug. Please report it to the issue tracker
var ErrInternalError = errors.New("an unexpected condition occurred")

//record type specifies a fixed length data block, the actual length disagrees
var ErrIncorrectDataLength = errors.New("incorrect data field length for type")

//data type is not one of the 6 recognized types
var ErrUnknownDataType = errors.New("unrecognized data type")

//Two records specify the data on the same address
var ErrSegmentOverlap = errors.New("segment overlap detected")

//Call to Seek resulted in offset before beginning
var ErrNegativeOffset = errors.New("negative offset")

//Whence value passed which was not the supported, other than (0,1,2)
var ErrUnsupportedWhence = errors.New("whence value unsupported")

//ParseError is returned from parse for all errors
//Err contains the underlying reason for the error
//Line contains the line number where the error occurred
type ParseError struct {
	Line int
	Err  error
}

func (p ParseError) Error() string {
	return fmt.Sprintf("parse error encountered on line %d: %s", p.Line, p.Err.Error())
}

//Type is an Enum of the supported kinds of intel hex records
type Type uint8

const (
	//Data indicates Record is of general data type, this is the data which will be present in memory
	Data Type = iota

	//EoF indicates Record is End of File, there should only be one of these
	//and it should be the last line
	EoF

	//ESA indicates Record is of Extended Segment Address, data portion*16 specifies the offset
	//to add to all future data records
	ESA

	//SSA indicates Record is of Start Segment Address, species the values of the CS and IP register
	//not needed for anything other than x86 architectures. These values are stored but
	//only one of these records per file is expected.
	SSA

	//ELA indicates Record is of Extended Linear Address type. The data field contains a 16-bit number
	//which is the upper portion of all following addresses in a 32-bit address space.
	ELA

	//SLA  indicates Record is of Start Linear Address type. The data field contains a 32-bit number
	//to be loaded into the EIP register. This is only relavant for the 80386 architecture.
	//This value is stored but only one of these records is expected.
	SLA
)

type header struct {
	Count   uint8
	Address uint16
	Type    Type
}

type rawRecord struct {
	Header header
	Data   []byte
}

func parseRecord(r io.Reader) (rawRecord, error) {
	var ret rawRecord
	err := binary.Read(r, binary.BigEndian, &ret.Header)
	if err != nil {
		return ret, err
	}
	ret.Data = make([]byte, ret.Header.Count)
	err = binary.Read(r, binary.BigEndian, &ret.Data)
	return ret, err
}

func parseRecordLine(bs []byte) (rawRecord, error) {
	if len(bs) < 11 { //minimum line size including start code
		return rawRecord{}, ErrInsufficentRecordLength
	}

	if bs[0] != ':' { //records all start with a colon
		return rawRecord{}, ErrNoStartCode
	}

	var length = hex.DecodedLen(len(bs[1:]))
	var decoded = make([]byte, length)
	n, err := hex.Decode(decoded, bs[1:]) //parse the hex values into a parsed byte slice
	if err != nil {
		return rawRecord{}, err
	}

	if n != length {
		return rawRecord{}, ErrUnexpectedDecodeLength
	}
	var sum int8 //confirm the checksum
	for _, v := range decoded {
		sum += int8(v)
	}
	if sum != 0 {
		return rawRecord{}, ErrChecksum
	}

	rdr := bytes.NewReader(decoded[:len(decoded)-1]) //convert new byte slice to reader
	r, err := parseRecord(rdr)
	if err != nil {
		return rawRecord{}, err
	}
	if rdr.Len() != 0 { //check that all the data was consumed
		return rawRecord{}, ErrExtraBytes
	}
	return r, nil
}

func parseHexFileRecords(r io.Reader) ([]rawRecord, error) {
	var ret []rawRecord
	scn := bufio.NewScanner(r)

	//parse in all the record lines
	i := 0
	for scn.Scan() {
		i++
		r, err := parseRecordLine(scn.Bytes())
		if err != nil {
			return nil, ParseError{Line: i, Err: err}
		}
		ret = append(ret, r)
	}
	if scn.Err() != nil {
		return nil, scn.Err()
	}

	//empty file is invalid file
	if len(ret) == 0 {
		return nil, ErrNoEOF
	}

	//make sure there is not EOF on a line other than the last
	for i, v := range ret[:len(ret)-1] {
		if v.Header.Type == EoF {
			return nil, ParseError{Line: i, Err: ErrUnexpectedEOF}
		}
	}

	//check that the last line is EOF
	if ret[len(ret)-1].Header.Type != EoF {
		return nil, ErrNoEOF
	}

	return ret[:len(ret)-1], nil //strip the EOF record because not needed anymore
}

//Record is a single contiguous block of specified memory
type Record struct {
	Offset uint32 //Offset in memory where the data block starts
	Data   []byte //the data which sits in the memory segment
}

//RecordList is a simple wrapper on a slice of
//Records so that they can be sorted in offset order
type RecordList []Record

func (r RecordList) Len() int {
	return len(r)
}

func (r RecordList) Less(i, j int) bool {
	return r[i].Offset < r[j].Offset
}

func (r RecordList) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

//File encapsulates all data for a parsed intel hex file
//CS,IP and EIP are x86 specific registers
type File struct {
	CS, IP uint16
	EIP    uint32
	Memory RecordList //contains the list of all the memory records in sorted order
}

//Parse an intel hex stream into memory
func Parse(r io.Reader) (File, error) {
	var ret File
	rs, err := parseHexFileRecords(r)
	if err != nil {
		return ret, err
	}
	var offset uint32
	for i, v := range rs {
		switch v.Header.Type {
		case Data:
			ret.Memory = append(ret.Memory, Record{
				Offset: offset + uint32(v.Header.Address),
				Data:   v.Data,
			})
		case EoF:
			//OMGWTFBBQ - there should only be one EOF at the end
			//this should have been verified and stripped previously
			//this should never happen
			return ret, ErrInternalError
		case ESA:
			var temp uint16
			if v.Header.Count != 2 {
				return ret, ParseError{Line: i, Err: ErrIncorrectDataLength}
			}
			err := binary.Read(bytes.NewReader(v.Data), binary.BigEndian, &temp)
			if err != nil {
				return ret, ParseError{Line: i, Err: err}
			}
			offset = uint32(temp) * 16
		case SSA:
			r := bytes.NewReader(v.Data)
			if v.Header.Count != 4 {
				return ret, ParseError{Line: i, Err: ErrIncorrectDataLength}
			}
			err := binary.Read(r, binary.BigEndian, &ret.CS)
			if err != nil {
				return ret, ParseError{Line: i, Err: err}
			}
			err = binary.Read(r, binary.BigEndian, &ret.IP)
			if err != nil {
				return ret, ParseError{Line: i, Err: err}
			}
		case ELA:
			if v.Header.Count != 2 {
				return ret, ParseError{Line: i, Err: ErrIncorrectDataLength}
			}
			r := bytes.NewReader(v.Data)
			var temp uint16
			err := binary.Read(r, binary.BigEndian, &temp)
			if err != nil {
				return ret, ParseError{Line: i, Err: err}
			}
			offset = (1 << 16) * uint32(temp)
		case SLA:
			if v.Header.Count != 4 {
				return ret, ParseError{Line: i, Err: ErrIncorrectDataLength}
			}
			r := bytes.NewReader(v.Data)
			err := binary.Read(r, binary.BigEndian, &ret.EIP)
			if err != nil {
				return ret, ParseError{Line: i, Err: err}
			}
		default:
			return ret, ErrUnknownDataType
		}
	}
	if len(ret.Memory) == 0 {
		return ret, nil //a file with no data is ok
	}
	sort.Sort(ret.Memory)
	var prevoffset = int(ret.Memory[0].Offset) + len(ret.Memory[0].Data)
	for _, v := range ret.Memory[1:] {
		if int(v.Offset) < prevoffset {
			return ret, ErrSegmentOverlap
		}
		prevoffset = int(v.Offset) + len(v.Data)
	}
	return ret, nil
}

//GetSegments retrieves all memory records which overlap the requested
//memory range in offset sorted order
func (f File) GetSegments(offset, length uint32) []Record {
	var ret []Record
	for _, v := range f.Memory {
		if offset < v.Offset+uint32(len(v.Data)) && offset+length > v.Offset {
			ret = append(ret, v)
		}
	}
	return ret
}

//GetByte retrieves a single byte from memory
//if the specified address does not intersect
//memory record then the pad byte is returned
func (f File) GetByte(addr uint32, pad byte) byte {
	s := f.GetSegments(addr, 1)
	if len(s) == 0 {
		return pad
	}
	//there should never be more than one segment
	return s[0].Data[int(addr-s[0].Offset)]
}

//returns true if the specified address is within or preceeds the record
//if preceeds fills with pad, else with the specified record value
func fillByte(addr uint32, dst *byte, r Record, pad byte) bool {
	if addr < r.Offset {
		*dst = pad
		return true
	}
	if addr < r.Offset+uint32(len(r.Data)) {
		*dst = r.Data[int(addr-r.Offset)]
		return true
	}
	return false
}

//Retrieve a block of bytes beginning at offset and of length
//len(dst). Any data in Records corrisponding to the address
//are used and if a Record does not align, dst is filled with
//pad.
func (f File) Retrieve(offset uint32, dst []byte, pad byte) {
	s := f.GetSegments(offset, uint32(len(dst)))
	segIndex := 0
	for i := range dst {
		//check if we are out of segments
		//if so finish the loop with pads
		if segIndex == len(s) {
			dst[i] = pad
			continue
		}
		//if we are not within or before the selected segment
		//advance the segment and retry
		//if we run out of segments take the pad and then bail
		for !fillByte(offset+uint32(i), &dst[i], s[segIndex], pad) {
			segIndex++
			if segIndex == len(s) {
				dst[i] = pad
				break
			}
		}
	}
}

//Size retrieves the value of offset + length from
//the last record. This is the highest specified data
func (f File) Size() int64 {
	if len(f.Memory) == 0 {
		return 0
	}
	last := f.Memory[len(f.Memory)-1]
	return int64(last.Offset + uint32(len(last.Data)))
}

//Retriever specifies required interface to access
//the file as a flat memory space
type Retriever interface {
	Retrieve(offset uint32, dst []byte, pad byte)
}

//Sizer is an object which can specify it's overall length
type Sizer interface {
	Size() int64
}

//RetrieveSizer is a composite interface so that FileReader
//can used additional types, aside from File
type RetrieveSizer interface {
	Retriever
	Sizer
}

//FileReader is a wrapper type to allow File (or composed types)
//to support the idiomatic Read functionality and make the memory
//seem more like a flat file. File does not specify contiguous blocks
//of memory and so Pad is mapped to virtually fill that space
type FileReader struct {
	RetrieveSizer
	Offset int64 //current offset within memory
	Pad    byte  //byte to fill in unspecified locations
}

//Read provides support for the io.Reader interface
func (fr *FileReader) Read(r []byte) (int, error) {
	max := fr.Size() - fr.Offset
	if max <= 0 { //no more to read so bail
		return 0, io.EOF
	}

	fr.Retrieve(uint32(fr.Offset), r, fr.Pad)
	if int64(len(r)) > max {
		fr.Offset += max
		return int(max), io.EOF
	}

	fr.Offset += int64(len(r))
	return len(r), nil
}

//ReadAt pulls data from a specific location in virtual memory space
func (fr *FileReader) ReadAt(r []byte, off int64) (int, error) {
	s := fr.Size()
	fr.Retrieve(uint32(off), r, fr.Pad)
	if off+int64(len(r)) > s {
		return int(s - off), io.EOF
	}
	return len(r), nil
}

//Seek allows to movement of the internal offset, to control subsequent
//Read calls.
func (fr *FileReader) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		fr.Offset = offset
	case 1:
		fr.Offset += offset
	case 2:
		fr.Offset = fr.Size() + offset
	default:
		return fr.Offset, ErrUnsupportedWhence
	}
	if fr.Offset < 0 {
		return fr.Offset, ErrNegativeOffset
	}
	return fr.Offset, nil
}
