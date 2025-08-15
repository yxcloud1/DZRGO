package driver

import (
	"acetek-mes/model"
	"acetek-mes/valconv"
	"log"
	"net/url"

	"github.com/yxcloud1/go-comm/db"
	"github.com/yxcloud1/go-comm/logger"
)

type DriverMgr struct {
	drivers []string
	clients map[string]IDriver
}

func NewDriverMgr(drivers []string) (*DriverMgr, error) {
	result := &DriverMgr{
		clients: make(map[string]IDriver),
		drivers: drivers,
	}
	result.load()
	return result, nil
}

func (mgr *DriverMgr) Start() {
	for k, v := range mgr.clients{
		log.Println("start plc driver ", k)
		v.Start()
	}
}

func (mgr *DriverMgr) load() error {

	var drivers[] model.DCDriver
	tx := db.DB().Conn().Debug().Find(&drivers, mgr.drivers)
	if tx.Error != nil{
		return tx.Error
	}
	for _, v := range drivers{
		var items[] model.DCItem
		tx = db.DB().Conn().Debug().Where(&model.DCItem{
			DriverID: v.ID,
		}).Find(&items)
		if tx.Error != nil{
			return tx.Error
		}
		var tags []*Tag
		for _, item := range items{
			tags = append(tags, &Tag{
				Name: item.ID,
				Address: item.Address,
				Datatype: item.DataType,
				Writable: item.Writable,
				Comment: item.Description,
				Value: func()interface{}{
					if item.DataType != "string" && item.Value == ""{
						return nil
					}else{
						if val, err:= valconv.StringToTargetType(item.Value, item.DataType);err != nil{
							return val
						}else{
							return val
						}
					}
				}(),
			})
		}
		if u, err := url.Parse(v.Url); err != nil{
			logger.TxtErr(err)
			continue;
		}else{
			if c,  err := NewDriver(v.ID, v.Name, u.Scheme, v.Url, tags);err == nil{
				mgr.clients[v.ID] = c
			}else{
				logger.TxtErr(err)
			}
		}
	}
	return nil
}

func (mgr *DriverMgr) Stop() error {

	return nil
}
