package valconv

import (
	"fmt"
	"strconv"
	"strings"
)

func StringToTargetType(value string, typ string) (interface{}, error) {
	switch strings.ToLower(typ) {
	case "int":
		return strconv.Atoi(value)
	case "int8":
		v, err := strconv.ParseInt(value, 10, 8)
		return int8(v), err
	case "int16":
		v, err := strconv.ParseInt(value, 10, 16)
		return int16(v), err
	case "int32":
		v, err := strconv.ParseInt(value, 10, 32)
		return int32(v), err
	case "int64":
		return strconv.ParseInt(value, 10, 64)
	case "uint":
		v, err := strconv.ParseUint(value, 10, 0)
		return uint(v), err
	case "uint8":
		v, err := strconv.ParseUint(value, 10, 8)
		return uint8(v), err
	case "uint16":
		v, err := strconv.ParseUint(value, 10, 16)
		return uint16(v), err
	case "uint32":
		v, err := strconv.ParseUint(value, 10, 32)
		return uint32(v), err
	case "uint64":
		return strconv.ParseUint(value, 10, 64)
	case "float32":
		v, err := strconv.ParseFloat(value, 32)
		return float32(v), err
	case "float64":
		return strconv.ParseFloat(value, 64)
	case "bool":
		return strconv.ParseBool(value)
	case "string":
		return value, nil
	default:
		return nil, fmt.Errorf("unsupported type: %s", typ)
	}
}