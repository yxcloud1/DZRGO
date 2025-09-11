package handler

import (
	"fmt"
	"net/http"
	"sync"

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
func removeClient(clientId string , id uuid.UUID) {
	clientsMu.Lock()
	if c, ok := clients[string(clientId)]; ok {
		if conn, exists := c[id]; exists {
			conn.Close()
			delete(c, id)
		}
	}
	clientsMu.Unlock()
}

func sendToClient(typeId string, epid string, msg string) error {
	if(msg == "") {
		return nil
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
