package service

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/bugfan/ang-admin/models"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512 * 1024 // 512 KB
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许跨域连接
	},
}

// WSMessage 统一的 WebSocket 消息协议
type WSMessage struct {
	Type      string      `json:"type"`      // 消息类型，如: "CERT_STATUS_CHANGE", "PING", "PONG", "SYSTEM_EVENT", "CAPTURE"
	Topic     string      `json:"topic"`     // 频道/主题，如: "cert", "cluster", "logs", "traffic"
	Data      interface{} `json:"data"`      // 数据载荷
	Timestamp int64       `json:"timestamp"` // 毫秒时间戳
}

// WSClient 单个 WebSocket 连接客户端
type WSClient struct {
	hub           *WSHub
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]bool
	mu            sync.RWMutex
}

// WSHub 全局 WebSocket 客户端管理与事件分发中心
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	mu         sync.RWMutex
}

var (
	globalHub     *WSHub
	globalHubOnce sync.Once
)

// GetWSHub 获取全局单例 WSHub
func GetWSHub() *WSHub {
	globalHubOnce.Do(func() {
		globalHub = &WSHub{
			clients:    make(map[*WSClient]bool),
			broadcast:  make(chan []byte, 256),
			register:   make(chan *WSClient),
			unregister: make(chan *WSClient),
		}
		go globalHub.run()
	})
	return globalHub
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("[WSHub] 客户端连接建立, 当前在线连接数: %d\n", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("[WSHub] 客户端连接断开, 当前在线连接数: %d\n", len(h.clients))
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastMessage 全局广播消息
func (h *WSHub) BroadcastMessage(msg *WSMessage) {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().UnixMilli()
	}
	bytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WSHub] 序列化广播消息失败: %v\n", err)
		return
	}
	h.broadcast <- bytes
}

// BroadcastEvent 便捷的全局事件广播函数
func BroadcastEvent(msgType, topic string, data interface{}) {
	GetWSHub().BroadcastMessage(&WSMessage{
		Type:      msgType,
		Topic:     topic,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	})
}

// BroadcastCertStatus 广播证书状态变更事件
func BroadcastCertStatus(cert *models.Certificate) {
	if cert == nil {
		return
	}
	BroadcastEvent("CERT_STATUS_CHANGE", "cert", cert)
}

func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WSHub] 客户端异常断开: %v\n", err)
			}
			break
		}

		// 处理客户端上行消息 (如心跳 PING 或订阅特定主题)
		var inMsg WSMessage
		if err := json.Unmarshal(message, &inMsg); err == nil {
			if inMsg.Type == "PING" {
				pongMsg, _ := json.Marshal(WSMessage{
					Type:      "PONG",
					Timestamp: time.Now().UnixMilli(),
				})
				c.send <- pongMsg
			}
		}
	}
}

func (c *WSClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[WSHub] 发送消息异常: %v\n", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWS 处理 WebSocket 连接握手升级
func ServeWS(hub *WSHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WSHub] 升级 WebSocket 协议失败: %v\n", err)
		return
	}

	client := &WSClient{
		hub:           hub,
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

	// 握手成功后发送一条欢迎及已就绪消息
	readyMsg, _ := json.Marshal(WSMessage{
		Type:      "WS_CONNECTED",
		Topic:     "system",
		Data:      "ANG-Admin WebSocket Connected Successfully",
		Timestamp: time.Now().UnixMilli(),
	})
	client.send <- readyMsg
}
