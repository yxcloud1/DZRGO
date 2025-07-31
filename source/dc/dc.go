package main

import (
	"acetek-mes/conf"
	"acetek-mes/driver"
	_ "acetek-mes/driver/s7"
	"acetek-mes/model"
	"acetek-mes/redishelper"
	"log"

	"github.com/yxcloud1/go-comm/db"
)

var (
	driverMgr *driver.DriverMgr
)

func init() {
	db.SetOption(conf.Conf().DB.Type, conf.Conf().DB.Url)
	log.Println("init redis ", redishelper.Instance().InitFromURL(conf.Conf().RedisConfig.Url))
	db.DB().Conn().Debug().AutoMigrate(model.GetEntitys()...)

	var err error;
	if driverMgr , err = driver.NewDriverMgr(conf.Conf().DataCollection.Drivers); err == nil{
		driverMgr.Start()
	}else{
		log.Println(err)
	}

}

func main() {
	select {} // 阻塞主线程
}
