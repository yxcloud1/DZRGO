package s7

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	driver "acetek-mes/driver"
)

const (
	AreaDB = 0x84 // 数据块
	AreaPE = 0x81 // 输入
	AreaPA = 0x82 // 外设输入
	AreaMK = 0x83 // 内存
	AreaTM = 0x1D
	AreaCT = 0x1C
)

type WordLen byte

const (
	Bit   WordLen = 0x01
	Byte  WordLen = 0x02
	Word  WordLen = 0x04
	DWord WordLen = 0x06
	Real  WordLen = 0x08 // float32
)

type S7Tag struct {
	Area      int
	DBNumber  int
	Start     int
	Bit       int
	WordLen   WordLen
	Length    int
	Valid     bool
	Raw       string
	ErrorInfo string
	DataType  string
}

func ParseAddress(address string) (*S7Tag, error) {
	tag := &S7Tag{Raw: address, Length: 1}

	dbRe := regexp.MustCompile(`(?i)^DB(\d+)\.DB([BXWD])(\d+)(?:\.(\d))?$`)
	stringRe := regexp.MustCompile(`(?i)^DB(\d+)\.STRING(\d+)(?:\((\d+)\))?$`)
	simpleRe := regexp.MustCompile(`(?i)^([PIQMVTCS])(\d+)(?:\.(\d))?$`)
	memRe := regexp.MustCompile(`(?i)^([VMIQ])([BWD])(\d+)$`) // 新增支持 VD100 等格式

	// 处理 STRING 类型
	if match := stringRe.FindStringSubmatch(address); match != nil {
		db, _ := strconv.Atoi(match[1])
		tag.DBNumber = db
		tag.Area = AreaDB
		tag.WordLen = Byte
		offset, _ := strconv.Atoi(match[2])
		tag.Length = 32
		l, err := strconv.Atoi(match[3])
		if err == nil && l > 0 && l <= 254 {
			tag.Length = l + 2
		}
		tag.DataType = "STRING"
		tag.Start = offset
		tag.Valid = true
		return tag, nil
	}

	// 处理 DBX/DBB/DBW/DBD 类型
	if match := dbRe.FindStringSubmatch(address); match != nil {
		db, _ := strconv.Atoi(match[1])
		offset, _ := strconv.Atoi(match[3])
		tag.DBNumber = db
		tag.Start = offset
		tag.Area = AreaDB

		switch strings.ToUpper(match[2]) {
		case "B":
			tag.WordLen = Byte
			tag.Length = 1
			tag.DataType = "BYTE"
		case "W":
			tag.WordLen = Word
			tag.Length = 2
			tag.DataType = "WORD"
		case "D":
			tag.WordLen = DWord
			tag.Length = 4
			tag.DataType = "DWORD"
		case "X":
			tag.WordLen = Bit
			tag.Length = 1
			tag.DataType = "BOOL"
			if match[4] != "" {
				tag.Bit, _ = strconv.Atoi(match[4])
			}
		}

		tag.Valid = true
		return tag, nil
	}

	// 新增处理 VD100 / MW200 等格式
	if match := memRe.FindStringSubmatch(address); match != nil {
		area := strings.ToUpper(match[1])
		dt := strings.ToUpper(match[2])
		offset, _ := strconv.Atoi(match[3])
		tag.Start = offset

		switch area {
		case "I":
			tag.Area = AreaPE
		case "Q":
			tag.Area = AreaPA
		case "M":
			tag.Area = AreaMK
		case "V":
			tag.Area = AreaDB
			tag.DBNumber = 1 // 假定 V 区在 DB1，可调整
		default:
			return nil, fmt.Errorf("unsupported area: %s", area)
		}

		switch dt {
		case "B":
			tag.WordLen = Byte
			tag.Length = 1
			tag.DataType = "BYTE"
		case "W":
			tag.WordLen = Word
			tag.Length = 2
			tag.DataType = "WORD"
		case "D":
			tag.WordLen = DWord
			tag.Length = 4
			tag.DataType = "DWORD"
		default:
			return nil, fmt.Errorf("unsupported datatype: %s", dt)
		}

		tag.Valid = true
		return tag, nil
	}

	// 处理单纯的 M10、Q20.2 这类
	if match := simpleRe.FindStringSubmatch(address); match != nil {
		area := strings.ToUpper(match[1])
		offset, _ := strconv.Atoi(match[2])
		tag.Start = offset
		if match[3] != "" {
			tag.Bit, _ = strconv.Atoi(match[3])
			tag.WordLen = Bit
			tag.Length = 1
			tag.DataType = "BOOL"
		} else {
			tag.WordLen = Byte
			tag.Length = 1
			tag.DataType = "BYTE"
		}

		switch area {
		case "I":
			tag.Area = AreaPE
		case "Q":
			tag.Area = AreaPA
		case "M":
			tag.Area = AreaMK
		case "V":
			tag.Area = AreaDB
			tag.DBNumber = 1 // 同样假定为 DB1
		case "T":
			tag.Area = AreaTM
		case "C":
			tag.Area = AreaCT
		default:
			return nil, fmt.Errorf("unsupported area: %s", area)
		}

		tag.Valid = true
		return tag, nil
	}

	return nil, fmt.Errorf("invalid address format: %s", address)
}

func ParseValueFromBuffer(tag S7Tag, buffer []byte) (any, error) {
	offset := 0
	if offset < 0 || offset >= len(buffer) {
		return nil, errors.New("byte offset out of bounds")
	}

	switch tag.DataType {
	case driver.TypeBool:
		if offset >= len(buffer) {
			return nil, errors.New("offset out of range for bool")
		}
		if tag.Bit > 7 {
			return nil, errors.New("bit offset must be 0-7 for bool")
		}
		b := buffer[offset]
		mask := byte(1 << tag.Bit)
		return (b & mask) != 0, nil

	case driver.TypeByte:
		return buffer[offset], nil

	case driver.TypeBytes:
		end := offset + tag.Length
		if end > len(buffer) {
			return nil, errors.New("length out of range for bytes")
		}
		return buffer[offset:end], nil

	case driver.TypeInt16:
		if offset+1 >= len(buffer) {
			return nil, errors.New("offset out of range for int16")
		}
		return int16(binary.BigEndian.Uint16(buffer[offset:])), nil

	case driver.TypeUInt16:
		if offset+1 >= len(buffer) {
			return nil, errors.New("offset out of range for uint16")
		}
		return binary.BigEndian.Uint16(buffer[offset:]), nil

	case driver.TypeInt32:
		if offset+3 >= len(buffer) {
			return nil, errors.New("offset out of range for int32")
		}
		return int32(binary.BigEndian.Uint32(buffer[offset:])), nil

	case driver.TypeUInt32:
		if offset+3 >= len(buffer) {
			return nil, errors.New("offset out of range for uint32")
		}
		return binary.BigEndian.Uint32(buffer[offset:]), nil

	case driver.TypeFloat32:
		if offset+3 >= len(buffer) {
			return nil, errors.New("offset out of range for float32")
		}
		bits := binary.BigEndian.Uint32(buffer[offset:])
		return math.Float32frombits(bits), nil
	case driver.TypeString:
		if len(buffer) < 2 {
			return nil, errors.New("buffer too short for S7 string header")
		}
		actualLen := int(buffer[1])
		if len(buffer) < 2+actualLen {
			return nil, errors.New("buffer too short for string data")
		}
		return string(buffer[2 : 2+actualLen]), nil

	default:
		return nil, errors.New("unsupported data type: " + string(tag.DataType))
	}
}
