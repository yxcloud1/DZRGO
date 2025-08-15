package handler

import (
	"acetek-mes/model"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yxcloud1/go-comm/db"
)

func GetRealClientAddr(c *gin.Context) (ip string, port string) {
	// 1. 优先从 Gin 内置方法获取真实 IP（会处理 X-Forwarded-For / X-Real-IP）
	ip = c.ClientIP()

	// 2. 尝试从 IIS 自定义头获取端口
	port = c.GetHeader("X-Client-Port") // 需要 IIS 反代配置转发 {REMOTE_PORT}

	// 3. 如果端口没取到，降级从 RemoteAddr 获取（直连或代理端口）
	if port == "" {
		host, p, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr))
		if err == nil {
			port = p
			// 如果 c.ClientIP() 和 host 一致，说明是直连，端口就是客户端端口
			// 如果不一致，说明有反代，端口是代理服务器到 Gin 的端口
			_ = host // 这里我们不覆盖 IP
		}
	}

	return ip, port
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

func saveDcLog(url string, client_ip string, deviceType string, deviceId string, request string) (int, error) {
	log := &model.LimsDcRequestLog{
		RequestUrl: url,
		DeviceType: deviceType,
		DeviceID:   deviceId,
		Request:    request,
		ClientIP:   client_ip,
	}
	if tx := db.DB().Conn().Save(log); tx.Error != nil {
		return 0, tx.Error
	} else {
		return log.ID, nil
	}
}
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
		ip = c.ClientIP() // fallback to default
	}
	return ip
}

func LimsDataCollection(c *gin.Context) {
	paramType := c.Param("type")
	paramID := c.Param("id")
	data := make(map[string]string)
	body, _ := c.GetRawData()
	response := gin.H{}
	rawID, _ := saveDcLog(getOriginalURL(c), getClientIP(c), paramType, paramID, string(body))
	if err := json.Unmarshal(body, &data); err != nil {
		response["error"] = err.Error()
		c.JSON(400, response)
		return
	}
	switch paramType {
	case "TP_TOLEDO":
		values := parseReceiveData(data["data"], "\t")
		response = saveReciveeData(rawID, paramType, paramID, data["data"], values)
	case "TP_WT":
		values := parseReceiveData(data["data"], " ")
		response = saveReciveeData(rawID, paramType, paramID, data["data"], values)
	default:
		response["type"] = paramType
		response["id"] = paramID
		response["data"] = string(body)
		values := parseReceiveData(data["data"], " ")
		response = saveReciveeData(rawID, paramType, paramID, data["data"], values)
	}

	c.JSON(http.StatusOK, response)
}

func LimsDataCollection2(c *gin.Context) {
	ip, port := GetRealClientAddr(c)
	log.Println("Real Client IP:", ip, "Port:", port)
	remoteAddr := fmt.Sprintf("%s:%s", ip, port)
	paramType, paramID :=  findDeviceByIP(remoteAddr, "")
	body, _ := c.GetRawData()

	context := string(body)
	response := gin.H{}
	rawID, _ := saveDcLog(getOriginalURL(c), remoteAddr, paramType, paramID, string(body))
	lines := strings.Split(context, "\n")
	for _, v := range lines {

		switch paramType {
		case "TP_TOLEDO":
			values := parseReceiveData(v, "\t")
			response = saveReciveeData(rawID, paramType, paramID, v, values)
		case "TP_WT":
			values := parseReceiveData(v, " ")
			response = saveReciveeData(rawID, paramType, paramID, v, values)
		default:
			response["type"] = paramType
			response["id"] = paramID
			response["data"] = string(body)
			values := parseReceiveData(v, " ")
			response = saveReciveeData(rawID, paramType, paramID, v, values)
		}
	}
	c.JSON(http.StatusOK, response)
}

func parseReceiveData(data string, splitStr string) []string {
	return strings.Split(data, splitStr)
}

func saveReciveeData(rawid int, deviceType string, deviceId string, data string, values []string) map[string]any {
	command := `EXEC sp_lims_save_dc_data @deviceType= ? ,@deviceID= ? ,@sampleID= ? ,@rawID= ? ,@rawData= ? , @item_codes= ? ,
								     @item_value1= ? , @item_value2= ? , @item_value3= ? , @item_value4= ? , @item_value5= ? ,
									 @item_value6= ? , @item_value7= ? , @item_value8= ? , @item_value9= ? , @item_value10= ? ,
									 @item_value11= ? , @item_value12= ? , @item_value13= ? , @item_value14= ? , @item_value15= ? `
	var params []interface{}
	params = append(params, deviceType, deviceId, "", rawid, data, "")
	for i := 0; i < 15; i++ {
		if len(values) > i {
			params = append(params, values[i])
		} else {
			params = append(params, nil)
		}
	}
	err := db.DB().ExecuteSQL(command, params...)
	if err != nil {
		log.Println(err)
	}
	return gin.H{
		"type": deviceType,
		"id":   deviceId,
	}
}
