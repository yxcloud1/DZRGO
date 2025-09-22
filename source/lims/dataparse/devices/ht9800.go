package devices

import (
	"acetek-mes/lims/dataparse"
	"fmt"
	"strconv"
	"strings"
)



type HT9800Decoder struct{
	buffer []byte
}

func (d *HT9800Decoder) Decode(data []byte) (map[string]interface{}, error) {
    line := strings.TrimSpace(string(data))

    parts := strings.Split(line, ",")
    if len(parts) < 3 {
        return nil, fmt.Errorf("invalid data format: %s", line)
    }

    weightStr := parts[2]
    weightStr = strings.TrimSuffix(weightStr, "kg")

    weight, err := strconv.ParseFloat(weightStr, 64)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "status": parts[0],
        "mode":   parts[1],
        "weight": weight,
    }, nil
}

// init 自动注册
func init() {
    dataparse.Register("HT9800", &HT9800Decoder{})
	dataparse.Register("XT3618", &HT9800Decoder{})
}