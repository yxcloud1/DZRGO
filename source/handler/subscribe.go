package handler

import (
	"acetek-mes/redishelper"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许任何跨域
		},
	}
	clientsMu sync.Mutex
	clients   = make(map[string]map[uuid.UUID]*websocket.Conn)
	sdmanger  = NewManager()
)

func addClient(clientId string, id uuid.UUID, conn *websocket.Conn) {
	clientsMu.Lock()
	if _, ok := clients[clientId]; !ok {
		clients[clientId] = make(map[uuid.UUID]*websocket.Conn)
	}
	clients[clientId][id] = conn
	clientsMu.Unlock()
}

// 移除客户端
func removeClient(clientId string, id uuid.UUID) {
	clientsMu.Lock()
	if c, ok := clients[string(clientId)]; ok {
		if conn, exists := c[id]; exists {
			conn.Close()
			delete(c, id)
		}
	}
	clientsMu.Unlock()
}

func sendToResis(typeId string, epid string, msg string) error {
	switch typeId {
	case "PH计", "电导率测试仪":
		vals := strings.Split(msg, "\r\n")
		if len(vals) >= 4 {
			if val, err := parseNumber(vals[3]); err == nil {
				return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now())
			}
		}
	case "快速水份仪":
		vals := strings.Split(msg, "\r\n")
		for _, v := range vals {
			if strings.HasPrefix(v, "End Result") {
				if val, err := parseNumber(vals[0]); err == nil {
					return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now())
				}
			}
		}
	case "BERTHOLD微波水分仪":
		vals := strings.Split(msg, "\t")
		if len(vals) >= 16 && vals[1] == "RUN" {
			if val, err := parseNumber(vals[15]); err == nil {
				return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now())
			}
		}
	case "耀华电子磅"://连续采集15次数据一致，才进行更新
		if val, err := parseNumber(msg); err == nil {
			if sdmanger.AddData(epid, val, 15) {
				sdmanger.Reset(epid)
				return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now())
			}
		}
	default:
	}
	return nil
}

func sendToClient(typeId string, epid string, msg string) error {
	if msg == "" {
		return nil
	}
	if err := sendToResis(typeId, epid, msg); err != nil {
		fmt.Printf("发送消息到Redis失败: %v\n", err)
	}
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clientID := fmt.Sprintf("%s_%s", typeId, epid)
	c, ok := clients[clientID]
	if !ok {
		return fmt.Errorf("客户端 %s 不存在", clientID)
	}
	for id, conn := range c {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			fmt.Printf("发送消息到客户端 %s 失败: %v\n", clientID, err)
			conn.Close()
			delete(c, id)
		}
	}
	if len(c) == 0 {
		delete(clients, clientID)
	}
	return nil
}

func SubscribeLimsDataCollection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	paramType := c.Param("type")
	paramID := c.Param("id")
	id, _ := uuid.NewUUID()
	clientID := fmt.Sprintf("%s_%s", paramType, paramID)

	addClient(clientID, id, conn)
	fmt.Println("客户端连接:", clientID)

	// 读循环（保持连接，直到出错/关闭）
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("客户端断开:", clientID)
			removeClient(clientID, id)
			break
		}
	}

}

func parseNumber(s string) (float64, error) {
	// 匹配整数或小数（支持正负号）

	re := regexp.MustCompile(`[-+]?\d*\.?\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0, fmt.Errorf("未找到数字")
	}
	return strconv.ParseFloat(match, 64)
}
