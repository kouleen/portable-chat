package utils

import (
	"fmt"
	"sync"
	"time"
)

type SSEEvent struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

type sseUserHub struct {
	mu      sync.RWMutex
	clients map[string]chan *SSEEvent
}

var sseClients sync.Map

func SubscribeSSE(userID int64) (string, <-chan *SSEEvent, func()) {
	hubAny, _ := sseClients.LoadOrStore(userID, &sseUserHub{
		clients: make(map[string]chan *SSEEvent),
	})
	hub := hubAny.(*sseUserHub)

	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	ch := make(chan *SSEEvent, 16)

	hub.mu.Lock()
	hub.clients[clientID] = ch
	hub.mu.Unlock()

	cancel := func() {
		hub.mu.Lock()
		client, ok := hub.clients[clientID]
		if ok {
			delete(hub.clients, clientID)
			close(client)
		}
		empty := len(hub.clients) == 0
		hub.mu.Unlock()
		if empty {
			sseClients.Delete(userID)
		}
	}

	return clientID, ch, cancel
}

func PublishSSE(userID int64, event *SSEEvent) {
	if event == nil {
		return
	}
	hubAny, ok := sseClients.Load(userID)
	if !ok {
		return
	}
	hub := hubAny.(*sseUserHub)

	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for _, ch := range hub.clients {
		select {
		case ch <- event:
		default:
		}
	}
}
