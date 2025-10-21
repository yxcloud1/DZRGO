package handler

import (
	"acetek-mes/redishelper"
	"fmt"
	"log"
	"net"
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
	clients   = make(map[string]map[uuid.UUID]*SafeConn)
	sdmanger  = NewManager()
)

type SafeConn struct {
	conn *websocket.Conn
	mut  sync.Mutex
	send chan []byte
}

func BCDToInt(b byte) int {
	high := (b >> 4) & 0x0F
	low := b & 0x0F
	return int(high*10 + low)
}

func addClient(clientId string, id uuid.UUID, conn *SafeConn) {
	log.Println("websocket connect", clientId, id)
	clientsMu.Lock()
	if _, ok := clients[clientId]; !ok {
		clients[clientId] = make(map[uuid.UUID]*SafeConn)
	}
	clients[clientId][id] = conn
	clientsMu.Unlock()
}

// 移除客户端
func removeClient(clientId string, id uuid.UUID) {
	log.Println("客户端断开连接:", clientId, id)
	clientsMu.Lock()
	if c, ok := clients[string(clientId)]; ok {
		if conn, exists := c[id]; exists {
			conn.conn.Close()
			delete(c, id)
		}
	}
	clientsMu.Unlock()
}

func sendToResis(typeId string, epid string, msg string) (error, interface{}) {
	switch typeId {
	case "PH计", "电导率测试仪":
		vals := strings.Split(msg, "\r\n")
		if len(vals) >= 4 {
			if val, err := parseNumber(vals[3]); err == nil {
				return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now()), val
			}
		}
	case "快速水份仪":
		vals := strings.Split(msg, "\r\n")
		for _, v := range vals {
			if strings.HasPrefix(v, "End Result") {
				if val, err := parseNumber(vals[0]); err == nil {
					return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now()), val
				}
			}
		}
	case "BERTHOLD微波水分仪":
		items := strings.Split(msg, "\r\n")
		vals := strings.Split(items[len(items)-1], "\t")
		if len(vals) >= 16 && vals[1] == "RUN" {
			if val, err := parseNumber(vals[13]); err == nil {
				return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now()), val
			}
		}
	case "耀华电子磅": //连续采集15次数据一致，才进行更新
		if val, err := parseNumber(msg); err == nil {
			if sdmanger.AddData(epid, val, 15) && val > 10.0 {
				//sdmanger.Reset(epid)
				return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now()), val
			}
		}
	case "HT9800", "XK3168":
		byts := []byte(msg)
		if len(byts) == 5 && byts[0] == 0xFF {
			// 0 0 0 0 0 0 0 0
			//超载 +-
			//  NC
			if byts[1]&0b10010000 == 0b00010000 /*7不超载；4稳定*/ {
				k := 1.0
				if byts[1]&0b00100000 == 0b00100000 { //负数
					k = -1.0
				}
				flag := byts[1] & 0b00000111
				flag = flag - 1
				for i := 0; i < int(flag); i++ {
					k = k * 0.1
				}

				val := 0.0
				val += float64(BCDToInt(byts[2]))
				val += float64(BCDToInt(byts[3]) * 100)
				val += float64(BCDToInt(byts[4]) * 10000)
				val = val * k
				///log.Println("----", typeId, epid, val)
				if sdmanger.AddData(epid, val, 15) && val > 10.0 {
					//sdmanger.Reset(epid)
					return redishelper.Instance().SetRealtime(epid, "value", val, "Good", time.Now()), val
				}
			}
		}
	default:
		return nil, msg
	}
	return nil, nil
}

func sendToClient(typeId string, epid string, msg string) error {
	if msg == "" {
		return nil
	}
	err, val := sendToResis(typeId, epid, msg)
	if err != nil {
		log.Printf("发送消息到Redis失败: %v\n", err)
	}
	clientID := fmt.Sprintf("%s_%s", typeId, epid)
	clientsMu.Lock()
	defer clientsMu.Unlock()
	c, ok := clients[clientID]
	if !ok {
		return fmt.Errorf("客户端 %s 不存在", clientID)
	}
	for _, conn := range c {
		message := []byte(msg)
		if val != nil {
			message = []byte(fmt.Sprintf("%v", val))
		log.Printf("发送消息到客户端 %s \n", clientID)
		conn.send <- message
		log.Printf("发送消息到客户端 %s 完成\n", clientID)
		}
		//if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		//	fmt.Printf("发送消息到客户端 %s 失败: %v\n", clientID, err)
		//	conn.Close()
		//	delete(c, id)
		//}
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

	safeConn := &SafeConn{
		conn: conn,
		send: make(chan []byte, 100),
	}
	addClient(clientID, id, safeConn)

	go func(sconn *SafeConn, cid string, i uuid.UUID) {
		tck := time.NewTicker(time.Second * 10)
		tck2:= time.NewTicker(time.Millisecond * 500)
		defer func() {
			removeClient(cid, i)
			tck.Stop()
			tck2.Stop()
			sconn.conn.Close()
		}()
		for {
			select {
			case msg, ok := <-sconn.send:
				sconn.mut.Lock()
				if !ok {
					sconn.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
					sconn.mut.Unlock()
					return
				}
				sconn.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := sconn.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					fmt.Println("❌ write error:", err)
					sconn.mut.Unlock()
					return
				}
				sconn.mut.Unlock()
			case <-tck.C:
				sconn.mut.Lock()
				sconn.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := sconn.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					fmt.Println("ping error:", err)
					sconn.mut.Unlock()
					return
				}
				sconn.mut.Unlock()
			case <- tck2.C:
				sconn.mut.Lock()
				sconn.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 500))
				if _, _, err := sconn.conn.ReadMessage(); err != nil {
					if ne, ok := err.(net.Error); ok && ne.Timeout(){
						sconn.mut.Unlock()
						continue;
					}else{
						fmt.Println("read error:", err)
						sconn.mut.Unlock()
						return
					}
				}
				sconn.mut.Unlock()
			}
		}
	}(safeConn, clientID, id)

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
