package hex
import "bufio"
import "errors"
import "os"
import "strconv"

type Type int

const (
	Data Type = iota
	EoF
	ESA
	SSA
	ELA
	SLA
)

type Cmd struct {
	Address uint16
	Type Type
	Data []byte
}

func getByteArray(chars string) ([]byte,error) {
	if len(chars)%2 != 0 {
		return nil, errors.New("length not kosher")
	}

	if len(chars) == 0 {
		return make([]byte,0),nil
	}
	front := chars[0:2]
	others, err := getByteArray(chars[2:])
	if err != nil {
		return others, err 
	}
	value,err := strconv.ParseUint(front,16,8)
	if err != nil {
		return others, err
	}
	frontVal := make([]byte,1)
	frontVal[0] = byte(value)
	return append(frontVal,others...),nil
}

func parseHexLine(chars string) ([]byte, error) {
	if chars[0] != ':' {
		return nil, errors.New("hex file line does not start with ':'")
	}
	return getByteArray(chars[1:])
}

func compareChecksum(data []byte) bool {
	var sum int8 = 0
	for _,v := range data {
		sum += int8(v)
	}
	return sum == 0
}

func NewCmd(hex string) (Cmd, error) {
	var ret Cmd
	bytes,err := parseHexLine(hex)
	if err != nil {
		return ret, err
	}

	valid := compareChecksum(bytes)
	if !valid {
		return ret, errors.New("Checksum does not match")
	}

	length := bytes[0]

	if len(bytes) != int(length)+4 {
		return ret, errors.New("Record incorrect length")
	}

	ret.Address = uint16(int(bytes[1])*256 + int(bytes[2]))
	ret.Type = Type(bytes[3])
	ret.Data = bytes[4:(len(bytes)-1)]
	return ret,nil
}


type Hex []Cmd

func New(filename string) (Hex, error) {
	var ret Hex = make([]Cmd,0)
	
	hdl, err := os.Open(filename)
	if err != nil {
		return ret,err
	}
	defer hdl.Close()

	scanner := bufio.NewScanner(hdl)
	for scanner.Scan() {
		cmd, err := NewCmd(scanner.Text())
		if err != nil {
			return ret, err
		}
		ret = append(ret,cmd)
	}
	if err := scanner.Err(); err != nil {
		return ret,err
	}
	
	return ret,nil
}
