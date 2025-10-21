package handler

import (
	"acetek-mes/dataservice"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	delayedTasks = make(map[string]*DelayedMessage, 0)
)

func getOriginalURL(c *gin.Context) string {
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}

	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	uri := c.Request.RequestURI
	return fmt.Sprintf("%s://%s%s", proto, host, uri)
}

func getClientIP(c *gin.Context) string {
	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.ClientIP()               // fallback to default
		if !strings.Contains(ip, ":") { // IPv6 address
			ip = c.Request.RemoteAddr
		}
	}
	return ip
}

func LimsDataCollection(c *gin.Context) {
	paramType := c.Param("type")
	paramID := c.Param("id")
	data := make(map[string]string)
	body, _ := c.GetRawData()

	response := gin.H{}
	rawID, _ := dataservice.SaveDcLog(getOriginalURL(c), getClientIP(c), paramType, paramID, string(body), body)
	if err := json.Unmarshal(body, &data); err != nil {
		response["error"] = err.Error()
		c.JSON(400, response)
		return
	}
	sendToClient(paramType, paramID, strings.Trim(data["data"], "\r\n"))
	response = dataservice.SaveReciveeData(rawID, paramType, paramID, data["data"], parseReceiveData(data["data"], "\r\n"))
	c.JSON(http.StatusOK, response)
}

func LimsDataCollection2(c *gin.Context) {
	remoteAddr := getClientIP(c)
	paramType, paramID, endFlag, dealy := dataservice.FindDeviceByIP(remoteAddr, "")
	body, _ := c.GetRawData()

	context := string(body)
	if context == "wn00000.0kg\r\n" {
		c.Status(http.StatusOK)
		return
	}
	if len(body) == 5 && body[0] == 0xFF && body[2] == 0 && body[3] == 0 && body[4] == 0 {
		c.Status(http.StatusOK)
		return
	}
	key := fmt.Sprintf("%s_%s", paramType, paramID)
	//log.Println("*******", paramType, paramID, dealy, context)
	if dealy > 0 {
		if dt, ok := delayedTasks[key]; !ok {
			delayedTasks[key] = NewDelayedTask(paramType, paramID, endFlag, time.Millisecond*time.Duration(dealy),
				func(d *DelayedMessage) {
					msg := strings.Trim(d.message, "\r\n")
					if msg != "" {
						sendToClient(paramType, paramID, msg)
						dataservice.SaveReciveeData(0, d.deviceType, d.deviceId, d.message, parseReceiveData(d.message, "\r\n"))
					}
				})
		} else {
			dt.delay = time.Millisecond * time.Duration(dealy)
			dt.endFlag = endFlag
		}
		dt := delayedTasks[key]
		dt.Receive(context, body)
	} else {
		msg := strings.Trim(context, "\r\n")
		if msg != "" {
			sendToClient(paramType, paramID, string(body))
			dataservice.SaveReciveeData(0, paramType, paramID, msg, parseReceiveData(msg, "\r\n"))
		}
	}
	if _, err := dataservice.SaveDcLog(getOriginalURL(c), remoteAddr, paramType, paramID, string(body), body); err != nil {
		log.Println("保存日志失败:", err)
	}
	c.Status(http.StatusOK)
}

func parseReceiveData(data string, splitStr string) []string {
	data = strings.Trim(data, "\n\r")
	return strings.Split(data, splitStr)
}
