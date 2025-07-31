package main

import (
	"acetek-mes/conf"
	"acetek-mes/model"

	"github.com/yxcloud1/go-comm/db"
)
func init(){
	db.SetOption(conf.Conf().DB.Type, conf.Conf().DB.Url)
	//db.DB().Conn().Debug().AutoMigrate(model.GetEntitys()...)
	//model.InitializeBaseData()
	//model.BuildPosition()
	model.BuildTraverse()


}
func main() {

}