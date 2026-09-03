package utils

import (
	"sync"

	"github.com/gorilla/websocket"
)

var webSocketClients sync.Map

type WebSocketClient struct {
	UserID int64
	Conn   *websocket.Conn
}

func SetWebSocketClient(id int64, conn *websocket.Conn) {
	webSocketClients.Store(id, &WebSocketClient{
		UserID: id,
		Conn:   conn,
	})
}

func GetWebSocketClient(id int64) *websocket.Conn {
	value, ok := webSocketClients.Load(id)
	if !ok {
		return nil
	}
	client := value.(*WebSocketClient)
	return client.Conn
}

func DeleteWebSocketClient(id int64) {
	webSocketClients.Delete(id)
}
