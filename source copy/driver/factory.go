package driver

import (
	"fmt"
)

type DriverConstructor func(id string, name string, rawURL string, tags []*Tag) (IDriver, error)

var driverRegistry = make(map[string]DriverConstructor)

func RegisterDriver(protocol string, constructor DriverConstructor) {
	driverRegistry[protocol] = constructor
}

func NewDriver(id string, name string, protocol, rawURL string, tags []*Tag) (IDriver, error) {
	if constructor, ok := driverRegistry[protocol]; ok {
		return constructor(id, name, rawURL, tags)
	}
	return nil, fmt.Errorf("unknown driver: %s", protocol)
}