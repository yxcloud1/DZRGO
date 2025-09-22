package s7

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"time"

	"acetek-mes/driver"

	"github.com/robinson/gos7"
	"github.com/yxcloud1/go-comm/logger"
)

type S7Client struct {
	driver.Driver
	handler *gos7.TCPClientHandler
	client  gos7.Client
	addr    string
	rack    int
	slot    int
}

func NewS7Client(id string, name string, rawURL string, tags []*driver.Tag) (driver.IDriver, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "102"
	}
	rack := 0
	slot := 1
	var interval uint32 = 1000
	if r := u.Query().Get("rack"); r != "" {
		rack, _ = strconv.Atoi(r)
	}
	if s := u.Query().Get("slot"); s != "" {
		slot, _ = strconv.Atoi(s)
	}
	if i := u.Query().Get("interval"); i != "" {
		if intval, err := strconv.Atoi(i); err == nil {
			interval = uint32(intval)
		}
	}
	handler := gos7.NewTCPClientHandler(fmt.Sprintf("%s:%s", host, port), rack, slot)
	handler.Timeout = 2 * time.Second

	client := gos7.NewClient(handler)
	mtags := make(map[string]*driver.Tag)
	for _, v := range tags {
		if s7t, err := ParseAddress(v.Address); err == nil {
			v.Parsed = true
			s7t.DataType = v.Datatype
			v.Mate = *s7t
			mtags[v.Name] = v
		} else {
			v.Parsed = false
			v.Mate = nil
			log.Println("parse address error:", err)
		}
	}

	return &S7Client{

		handler: handler,
		client:  client,
		addr:    rawURL,
		rack:    rack,
		slot:    slot,
		Driver: driver.Driver{
			ID:            id,
			Name:          name,
			Tags:          mtags,
			Interval:      interval,
			ChCommand:     make(chan string, 100),
			ChWrite:       make(chan map[string]interface{}, 100),
			ChWriteResult: make(chan error),
		},
	}, nil
}

// 注册 S7 驱动
func init() {
	logger.TxtLog("register driver s7")
	driver.RegisterDriver("s7", NewS7Client)
}

func (c *S7Client) Connect() error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if c.Connected {
		return nil
	}
	if err := c.handler.Connect(); err != nil {
		return err
	}
	c.Connected = true
	return nil
}

func (c *S7Client) Close() {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	c.handler.Close()
	c.Connected = false
}

func (c *S7Client) IsConnected() bool {
	return c.Connected
}

func (c *S7Client) reconnectIfNeeded() error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	if !c.Connected || time.Since(c.LastPing) > 5*time.Second {
		c.handler.Close()
		err := c.handler.Connect()
		if err != nil {
			c.Connected = false
			return err
		}
		c.Connected = true
		c.LastPing = time.Now()
	}
	return nil
}

func (c *S7Client) Read() (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if err := c.reconnectIfNeeded(); err != nil {
		return nil, err
	}
	for _, v := range c.Tags {
		if !v.Parsed {
			continue
		}
		var resultError error
		byts := make([]byte, 256)
		if s7tag, ok := v.Mate.(S7Tag); ok {
			switch s7tag.Area {
			case AreaDB:
				 resultError= c.client.AGReadDB(s7tag.DBNumber, s7tag.Start, s7tag.Length, byts)
			case AreaTM:
				resultError= c.client.AGReadTM(s7tag.Start, s7tag.Length, byts)
			case AreaCT:
				resultError= c.client.AGReadCT(s7tag.Start, s7tag.Length, byts)
			case AreaMK:
				resultError= c.client.AGReadMB(s7tag.Start, s7tag.Length, byts)
			case AreaPA:
				resultError= c.client.AGReadEB(s7tag.Start, s7tag.Length, byts)
			case AreaPE:
				resultError= c.client.AGReadAB(s7tag.Start, s7tag.Length, byts)
			default:
				log.Println("tag ")
			}
			log.Printf("resultError:%v\n %+v %s", resultError, s7tag, hex.EncodeToString(byts[0:s7tag.Length]))
			value, err := ParseValueFromBuffer(s7tag, byts)
			if resultError == nil && err == nil {
				result[v.Name] = value
				v.Quality = "Good"
				v.Value = value
				v.Timestamp = time.Now()
			} else {
				result[v.Name] = nil
				v.Quality = "Bad"
				v.Value = nil
				v.Timestamp = time.Now()
				log.Println("resultError:", resultError, "parse error:", err)
			}

		} else {
			log.Println("tag is not s7 tag")
		}
		result[v.Name] = v.Value
	}
	return result, nil
}

func (c *S7Client) Write(name string, value interface{}) error {
	c.ChWrite <- map[string]interface{}{
		name: value,
	}
	return <-c.ChWriteResult
}

func (c *S7Client) write(values map[string]interface{}) error {
	for k, v := range values {
		if t, ok := c.Tags[k]; !ok {
			return fmt.Errorf("tag %s is not define", k)
		} else {
			if !t.Writable {
				return fmt.Errorf("tag %s is readonly", k)
			} else {
				return c.WriteTag(k, v)
			}
		}
	}
	return nil
}
func (c *S7Client) WriteTag(name string, value interface{}) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	tag, ok := c.Tags[name]
	if !ok {
		return fmt.Errorf("tag %s is not defined", name)
	}
	t, ok := tag.Mate.(S7Tag)
	if !ok {
		return fmt.Errorf("tag %s Mate not S7Tag", name)
	}

	// bool 位操作写入
	if tag.Datatype == "bool" && t.WordLen == Bit {
		buffer := make([]byte, 1)
		var err error
		switch t.Area {
		case AreaDB:
			err = c.client.AGReadDB(t.DBNumber, t.Start, 1, buffer)
		case AreaMK:
			err = c.client.AGReadMB(t.Start, 1, buffer)
		case AreaPE:
			err = c.client.AGReadEB(t.Start, 1, buffer)
		case AreaPA:
			err = c.client.AGReadAB(t.Start, 1, buffer)
		default:
			return fmt.Errorf("不支持的区域写入 bool")
		}
		if err != nil {
			return fmt.Errorf("读取原始字节失败: %v", err)
		}

		bit := uint(t.Bit)
		v, ok := value.(bool)
		if !ok {
			return fmt.Errorf("期望写入 bool 类型")
		}
		if v {
			buffer[0] |= (1 << bit)
		} else {
			buffer[0] &^= (1 << bit)
		}

		switch t.Area {
		case AreaDB:
			return c.client.AGWriteDB(t.DBNumber, t.Start, 1, buffer)
		case AreaMK:
			return c.client.AGWriteMB(t.Start, 1, buffer)
		case AreaPE:
			return c.client.AGWriteEB(t.Start, 1, buffer)
		case AreaPA:
			return c.client.AGWriteAB(t.Start, 1, buffer)
		default:
			return fmt.Errorf("不支持的区域写入 bool")
		}
	}

	// 普通类型写入（非位）
	data, err := tag.ConvertToBytes(value)
	if err != nil {
		return err
	}
	length := len(data)

	switch t.Area {
	case AreaDB:
		return c.client.AGWriteDB(t.DBNumber, t.Start, length, data)
	case AreaMK:
		return c.client.AGWriteMB(t.Start, length, data)
	case AreaPE:
		return c.client.AGWriteEB(t.Start, length, data)
	case AreaPA:
		return c.client.AGWriteAB(t.Start, length, data)
	default:
		return fmt.Errorf("未知区域: %d", t.Area)
	}
}
func (c *S7Client) Start() error {
	go func() {
		ticker := time.NewTicker(time.Duration(c.Interval) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ts := time.Now()
				if values, err := c.Read(); err != nil {
					c.FailCount++
					if c.FailCount > 5 {
						for _, v := range c.Tags {
							v.Quality = "Bad"
							v.Timestamp = ts
						}
					}
					log.Println("read error:", err)
				} else {
					log.Printf("%+v", values)
					c.FailCount = 0
				}
				c.Update()
			case command := <-c.ChCommand:
				switch command {
				case "stop":
					return
				default:
					break
				}
			case writeValues := <-c.ChWrite:
				log.Println("write value ", writeValues)
				c.ChWriteResult <- c.write(writeValues)
			}
		}
	}()
	return nil
}

func (c *S7Client) Stop() error {
	c.ChCommand <- "stop"
	return nil
}

func (c *S7Client) Reconfig() error {
	return nil
}
