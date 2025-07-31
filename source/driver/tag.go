package driver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
)

type Tag struct {
	Name      string // 自定义名称
	Address   string // 原始地址字符串，如 "DB1.DBW20" 或 "I0.0"
	Datatype  string // "int", "float", "bool" 等
	Parsed    bool
	Comment   string
	Access    string // r/w/rw
	Writable  bool
	Value     interface{}
	Timestamp time.Time
	Quality   string
	Mate      any
}

const (
	TypeBool    string = "bool"
	TypeInt16   string = "int16"
	TypeUInt16  string = "uint16"
	TypeInt32   string = "int32"
	TypeUInt32  string = "uint32"
	TypeFloat32 string = "float32"
	TypeByte    string = "byte"
	TypeBytes   string = "bytes"
	TypeString string = "string"
)

func (t* Tag) ConvertValue(value interface{}) interface{}{
	return convert(value, t.Datatype)
}

func (t* Tag ) ConvertToBytes(value interface{}) ([]byte, error){
	return toBytes(value, t.Datatype)
}

func convert(value interface{}, t string) interface{} {
	switch t {
	case TypeBool:
		switch v := value.(type) {
		case bool:
			return v
		case string:
			b, _ := strconv.ParseBool(v)
			return b
		case int:
			return v != 0
		}
	case TypeInt16:
		switch v := value.(type) {
		case string:
			i, _ := strconv.ParseInt(v, 10, 16)
			return int16(i)
		case int:
			return int16(v)
		case float64:
			return int16(v)
		}
	case TypeUInt16:
		switch v := value.(type) {
		case string:
			i, _ := strconv.ParseUint(v, 10, 16)
			return uint16(i)
		case int:
			return uint16(v)
		case float64:
			return uint16(v)
		}
	case TypeInt32:
		switch v := value.(type) {
		case string:
			i, _ := strconv.ParseInt(v, 10, 32)
			return int32(i)
		case int:
			return int32(v)
		case float64:
			return int32(v)
		}
	case TypeUInt32:
		switch v := value.(type) {
		case string:
			i, _ := strconv.ParseUint(v, 10, 32)
			return uint32(i)
		case int:
			return uint32(v)
		case float64:
			return uint32(v)
		}
	case TypeFloat32:
		switch v := value.(type) {
		case string:
			f, _ := strconv.ParseFloat(v, 32)
			return float32(f)
		case float64:
			return float32(v)
		case int:
			return float32(v)
		}
	case TypeByte:
		switch v := value.(type) {
		case byte:
			return v
		case int:
			return byte(v)
		case string:
			bs := []byte(v)
			if len(bs) > 0 {
				return bs[0]
			}
		}
	case TypeBytes:
		switch v := value.(type) {
		case []byte:
			return v
		case string:
			return []byte(v)
		}
	case TypeString:
		return fmt.Sprintf("%v", value)
	}
	return nil
}


func toBytes(value interface{}, t string) ([]byte, error) {

	value = convert(value, t)

	buf := new(bytes.Buffer)

	switch t {
	case TypeBool:
		var b byte = 0
		if value.(bool) {
			b = 1
		}
		return []byte{b}, nil

	case TypeInt16:
		binary.Write(buf, binary.BigEndian, value.(int16))
	case TypeUInt16:
		binary.Write(buf, binary.BigEndian, value.(uint16))
	case TypeInt32:
		binary.Write(buf, binary.BigEndian, value.(int32))
	case TypeUInt32:
		binary.Write(buf, binary.BigEndian, value.(uint32))
	case TypeFloat32:
		binary.Write(buf, binary.BigEndian, value.(float32))
	case TypeByte:
		return []byte{value.(byte)}, nil
	case TypeBytes:
		return value.([]byte), nil
	case TypeString:
		s := value.(string)
		b := make([]byte, 256)
		copy(b, []byte(s))
		return b, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", t)
	}
	return buf.Bytes(), nil
}