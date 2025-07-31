package s7

import (
	"log"
	"testing"
	"time"

	driver "acetek-mes/driver"
)
const
	connect_url = "s7://172.20.10.6:102?rack=2&slot=0"
var(
	tags = []*driver.Tag{
		&driver.Tag{
			Name: "DB1.DBX0.1",
			Address: "DB1.DBX0.2",
			Datatype: "bool",
			Writable: true,
		},
		&driver.Tag{
			Name: "DB1.DBW16",
			Address: "DB1.DBW16",
			Datatype: "uint16",
			Writable: true,
		},
		&driver.Tag{
			Name: "DB1.DBD32",
			Address: "DB1.DBD32",
			Datatype: "float32",
			Writable: true,
		},
	}
)
func TestRead(t *testing.T) {
	log.Println("test read")
	if c, err := NewS7Client("OOD","", connect_url, tags);err!= nil{
		t.Log(err)
	}else{
		c.Start()
		tck := time.NewTicker(time.Second)
	for{
		select{
		case <-tck.C:
			log.Println("write ", c.Write("DB1.DBX0.1", time.Now().Second()%2 == 0));
		}
	}
	}


	log.Println("read complete")
}