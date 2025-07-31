package main

import (
	"acetek-mes/conf"
	"acetek-mes/handler"
	"acetek-mes/tcpserver"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yxcloud1/go-comm/db"
	"github.com/yxcloud1/go-comm/logger"
	"github.com/yxcloud1/go-comm/winservice"
)

func init() {
	db.SetOption(conf.Conf().DB.Type, conf.Conf().DB.Url)
}

func startApi() {
	r := gin.Default()
	r.Use(handler.GlobalMiddleware())
	path := conf.Conf().Api.Path
	if path == "" {
		path = "/"
	}
	path = path + "/:type/:id"
	r.POST(path, handler.LimsDataCollection)
	go r.Run(conf.Conf().Api.ListenAddr)

}

func stop() error {
	logger.TxtLog("Stopping application...")
	tcpserver.Stop()
	time.Sleep(time.Second)
	logger.TxtLog("Application stopped.")
	return nil
}

func start() error {
	startApi()
	handlers := make(map[string]func(clientAddr string, message string))
	handlers[":9100"] = receiveCallback9100
	handlers[":9002"] = receiveCallback9002
	tcpserver.Start(handlers)
	return nil
}
func receiveCallback9100(addr string, content string) {
	if content != "" {
		sendToLimsDCApi(addr, "9100", content)
	}
}

func receiveCallback9002(addr string, content string) {
	if content != "" {
		sendToLimsDCApi(addr, "9002", content)
	}
}

func findDeviceByIP(addr string , port string) (string, string){
	if res, err := db.DB().ExecuteQuery("exec sp_lims_query_device_by_ip @addr = ? , @type = ? ", addr, port);err != nil{
	return addr, port
	}else{
		if len(res) > 0{
			res1 := addr
			res2 := port
			if v, ok := res[0]["type"]; ok{
				res1 = fmt.Sprintf("%v", v)
			}
			if v, ok := res[0]["addr"]; ok{
				res2 = fmt.Sprintf("%v", v)
			}
			return res1, res2
		}else{
			return addr, port
		}
	}
}

func sendToLimsDCApi(addr string, port string, content string) error {
	url := conf.Conf().Api.ListenAddr
	if url == "" {
		url = ":8000"
	}
	if url[0] == ':' {
		url = "http://127.0.0.1" + url
	} else {
		url = "http://" + url
	}
	deviceType, deviceID := findDeviceByIP(addr, port)
	url = fmt.Sprintf("%s%s/%s/%s", url, conf.Conf().Api.Path, deviceType, deviceID)

	byts, _ := json.Marshal(map[string]interface{}{
		"data": content,
	})
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(byts))
	if err != nil {
		fmt.Printf("POST失败: %v\n", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	winservice.RunAsService("LimsDataCollection", "LimsDataCollection", "LIMS 数据采集", start, stop)
}
