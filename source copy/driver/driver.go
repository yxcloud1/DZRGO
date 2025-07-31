package driver

import (
	"acetek-mes/redishelper"
	"log"
	"sync"
	"time"
)

type IDriver interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	Read() (map[string]interface{}, error)
	Write(tag string, value interface{}) error
	Start() error
	Stop() error
	Reconfig() error
}

type Driver struct {
	ID            string
	Name          string
	Tags          map[string]*Tag
	Mu            sync.Mutex
	Connected     bool
	LastPing      time.Time
	Interval      uint32
	ChCommand     chan string
	ChWrite       chan map[string]interface{}
	ChWriteResult chan error
	FailCount     int
}

func (d *Driver) Connect() error {
	return nil
}
func (d *Driver) Disconnect() error {
	return nil
}
func (d *Driver) IsConnected() bool {
	return false
}
func (d *Driver) Read([]*Tag) (map[string]interface{}, error) {
	return nil, nil
}
func (d *Driver) Write(map[string]interface{}) error {
	return nil
}
func (d *Driver) Start() error {
	return nil
}
func (d *Driver) Stop() error {
	d.ChCommand <- "stop"
	return nil
}
func (d *Driver) Reconfig() error {
	d.ChCommand <- "reconfig"
	return nil
}

func (d *Driver) Update() error {
	for _, v := range d.Tags{
		err := redishelper.Instance().SetRealtime(d.ID, v.Name, v.Value, v.Quality, v.Timestamp)
		if err != nil{
			log.Println("set realtime",err)
		}
	}
	return nil
}
