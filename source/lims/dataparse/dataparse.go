package dataparse

import "fmt"

type Decoder interface {
    Decode(data []byte) (map[string]interface{}, error)
}

var decoders = make(map[string]Decoder)

func Register(device string, decoder Decoder) {
    if _, exists := decoders[device]; exists {
        panic(fmt.Sprintf("decoder for %s already registered", device))
    }
    decoders[device] = decoder
}

func Decode(device string, data []byte) (map[string]interface{}, error) {
    decoder, ok := decoders[device]
    if !ok {
        return nil, fmt.Errorf("no decoder found for device %s", device)
    }
    return decoder.Decode(data)
}