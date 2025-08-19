package main

import (
	"acetek-mes/conf"
	"acetek-mes/handler"
	"acetek-mes/model"
	"acetek-mes/tcpserver"
	"acetek-mes/udpserver"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yxcloud1/go-comm/db"
	"github.com/yxcloud1/go-comm/logger"
	"github.com/yxcloud1/go-comm/winservice"
)
var(

)

func init() {
	db.SetOption(conf.Conf().DB.Type, conf.Conf().DB.Url)
	db.DB().Conn().Debug().AutoMigrate(&model.LimsDcRequestLog{})
}

func startApi() {
	r := gin.Default()
	r.Use(handler.GlobalMiddleware())
	path := conf.Conf().Api.Path
	if path == "" {
		path = "/"
	}
	r.POST(path+"/serial", handler.LimsDataCollection2)
	path = path + "/:type/:id"
	r.POST(path, handler.LimsDataCollection)

	go r.Run(conf.Conf().Api.ListenAddr)

}

func stop() error {
	logger.TxtLog("Stopping application...")
	tcpserver.Stop()
	time.Sleep(time.Second)
	udpserver.Stop()
	time.Sleep(time.Second)
	logger.TxtLog("Application stopped.")

	return nil
}

func start() error {
	startApi()
	handlers := make(map[string]func(clientAddr string, message string, raw []byte), 0)
	handlers[":9100"] = receiveCallback9100
	tcpserver.Start(handlers)
	handlers[":9002"] = receiveCallback9002
	udpserver.Start(handlers)
	return nil
}
func receiveCallback9100(addr string, content string, raw []byte) {
	if content != "" {
		sendToLimsDCApi(addr, "9100", content, raw)
	}
}

func receiveCallback9002(addr string, content string, raw []byte) {
	if content != "" {
		sendToLimsDCApi(addr, "9002", content, raw)
	}
}
func sendToLimsDCApi(addr string, port string, content string, raw []byte) error {
	url := conf.Conf().Api.ListenAddr
	if url == "" {
		url = ":8000"
	}
	if url[0] == ':' {
		url = "http://127.0.0.1" + url
	} else {
		url = "http://" + url
	}
	//deviceType, deviceID, _, _ := dataservice.FindDeviceByIP(addr, port)
	url = fmt.Sprintf("%s%s/serial", url, conf.Conf().Api.Path)

	//byts, _ := json.Marshal(map[string]interface{}{
	//	"data": content,
	//	"raw" : raw,
	//})
	//X-Forwarded-For
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(raw))
	if err != nil {
		log.Println("创建请求失败:", err)
		return err
	}
	//req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", addr)
	client := &http.Client{}
	resp, err := client.Do(req)
	//resp, err := http.Post(url, "application/json", bytes.NewBuffer(byts))
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
