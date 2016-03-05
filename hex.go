package intelhex

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

var InsufficentRecordLength = errors.New("record of insufficient length")
var NoStartCode = errors.New("line not prefixed with start code ':'")
var UnexpectedDecodeLength = errors.New("decoded hex string at unexpected length")
var ChecksumError = errors.New("checksum invalid")
var NoEOF = errors.New("failed to locate EOF record")
var UnexpectedEOF = errors.New("encountered EOF on line other than the last")
var ExtraBytes = errors.New("record contained extra data bytes")
var InternalError = errors.New("an unexpected condition occurred")
var IncorrectDataLength = errors.New("incorrect data field length for type")
var UnknownDataType = errors.New("unrecognized data type")
var SegmentOverlap = errors.New("segment overlap detected")

type ParseError struct {
	Line int
	Err  error
}

func (p ParseError) Error() string {
	return fmt.Sprintf("parse error encountered on line %d: %s", p.Line, p.Err.Error())
}

type Type uint8

const (
	Data Type = iota
	EoF
	ESA
	SSA
	ELA
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
		return rawRecord{}, InsufficentRecordLength
	}

	if bs[0] != ':' { //records all start with a colon
		return rawRecord{}, NoStartCode
	}

	var length = hex.DecodedLen(len(bs[1:]))
	var decoded = make([]byte, length)
	n, err := hex.Decode(decoded, bs[1:]) //parse the hex values into a parsed byte slice
	if err != nil {
		return rawRecord{}, err
	}

	if n != length {
		return rawRecord{}, UnexpectedDecodeLength
	}
	var sum int8 //confirm the checksum
	for _, v := range decoded {
		sum += int8(v)
	}
	if sum != 0 {
		return rawRecord{}, ChecksumError
	}

	rdr := bytes.NewReader(decoded[:len(decoded)-1]) //convert new byte slice to reader
	r, err := parseRecord(rdr)
	if err != nil {
		return rawRecord{}, err
	}
	if rdr.Len() != 0 { //check that all the data was consumed
		return rawRecord{}, ExtraBytes
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
		return nil, NoEOF
	}

	//make sure there is not EOF on a line other than the last
	for i, v := range ret[:len(ret)-1] {
		if v.Header.Type == EoF {
			return nil, ParseError{Line: i, Err: UnexpectedEOF}
		}
	}

	//check that the last line is EOF
	if ret[len(ret)-1].Header.Type != EoF {
		return nil, NoEOF
	}

	return ret[:len(ret)-1], nil //strip the EOF record because not needed anymore
}

type Record struct {
	Offset uint32
	Data   []byte
}

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

type File struct {
	CS, IP uint16
	EIP    uint32
	Memory RecordList
}

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
			return ret, InternalError
		case ESA:
			var temp uint16
			if v.Header.Count != 2 {
				return ret, ParseError{Line: i, Err: IncorrectDataLength}
			}
			err := binary.Read(bytes.NewReader(v.Data), binary.BigEndian, &temp)
			if err != nil {
				return ret, ParseError{Line: i, Err: err}
			}
			offset = uint32(temp) * 16
		case SSA:
			r := bytes.NewReader(v.Data)
			if v.Header.Count != 4 {
				return ret, ParseError{Line: i, Err: IncorrectDataLength}
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
				return ret, ParseError{Line: i, Err: IncorrectDataLength}
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
				return ret, ParseError{Line: i, Err: IncorrectDataLength}
			}
			r := bytes.NewReader(v.Data)
			err := binary.Read(r, binary.BigEndian, &ret.EIP)
			if err != nil {
				return ret, ParseError{Line: i, Err: err}
			}
		default:
			return ret, UnknownDataType
		}
	}
	if len(ret.Memory) == 0 {
		return ret, nil //a file with no data is ok
	}
	sort.Sort(ret.Memory)
	var prevoffset = int(ret.Memory[0].Offset) + len(ret.Memory[0].Data)
	for _, v := range ret.Memory[1:] {
		if int(v.Offset) < prevoffset {
			return ret, SegmentOverlap
		}
		prevoffset = int(v.Offset) + len(v.Data)
	}
	return ret, nil
}

//retrieve all memory records which overlap the requested
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

//retrieve a single byte from memory
//if no overlap then returns the pad byte
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

//retrieve a memory segment locations which do not align with
//a memory record are filled with pad byte
func (f File) Retrieve(offset uint32, dst []byte, pad byte) {
	s := f.GetSegments(offset, uint32(len(dst)))
	segment_index := 0
	for i := range dst {
		//check if we are out of segments
		//if so finish the loop with pads
		if segment_index == len(s) {
			dst[i] = pad
			continue
		}
		//if we are not within or before the selected segment
		//advance the segment and retry
		//if we run out of segments take the pad and then bail
		for !fillByte(offset+uint32(i), &dst[i], s[segment_index], pad) {
			segment_index++
			if segment_index == len(s) {
				dst[i] = pad
				break
			}
		}
	}
}

func (f File) Size() int64 {
	if len(f.Memory) == 0 {
		return 0
	}
	last := f.Memory[len(f.Memory)-1]
	return int64(last.Offset + uint32(len(last.Data)))
}

type FileReader struct {
	File        //embed the hex file
	Offset int  //current offset within memory
	Pad    byte //byte to fill in unspecified locations
}

func (fr *FileReader) Read(r []byte) (int, error) {
	max := int(fr.File.Size()) - fr.Offset
	if max <= 0 { //no more to read so bail
		return 0, io.EOF
	}

	fr.Retrieve(uint32(fr.Offset), r, fr.Pad)
	if len(r) > max {
		fr.Offset += max
		return max, io.EOF
	}

	fr.Offset += len(r)
	return len(r), nil
}

func (fr *FileReader) ReadAt(r []byte, off int64) (int, error) {
	s := fr.Size()
	fr.Retrieve(uint32(off), r, fr.Pad)
	if off+int64(len(r)) > s {
		return int(s - off), io.EOF
	}
	return len(r), nil
}
