package utils

import (
	"sync"

	"github.com/gorilla/websocket"
)

var webSocketClients sync.Map

type WebSocketClient struct {
	UserID    int64
	ContactID int64
	Conn      *websocket.Conn
	mu        sync.Mutex
}

func (c *WebSocketClient) WriteJSON(v any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.Close()
}

func SetWebSocketClient(id, contactID int64, conn *websocket.Conn) *WebSocketClient {
	client := &WebSocketClient{
		UserID:    id,
		ContactID: contactID,
		Conn:      conn,
	}
	if oldClient := GetWebSocketClient(id); oldClient != nil {
		_ = oldClient.Close()
	}
	webSocketClients.Store(id, client)
	return client
}

func GetWebSocketClient(id int64) *WebSocketClient {
	value, ok := webSocketClients.Load(id)
	if !ok {
		return nil
	}
	client := value.(*WebSocketClient)
	return client
}

func DeleteWebSocketClient(id int64, conn *websocket.Conn) {
	value, ok := webSocketClients.Load(id)
	if !ok {
		return
	}
	client := value.(*WebSocketClient)
	if conn != nil && client.Conn != conn {
		return
	}
	webSocketClients.Delete(id)
}
