package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	utils "github.com/kouleen/portable-chat/pkg/util"
)

func StreamEvents(c *gin.Context) {
	ctx := c.Request.Context()
	user := utils.GetAuthorization(ctx)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Unauthorized"})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "streaming unsupported"})
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	_, ch, cancel := utils.SubscribeSSE(user.ID)
	defer cancel()

	writeEvent := func(event *utils.SSEEvent) bool {
		payload, err := json.Marshal(event.Data)
		if err != nil {
			payload = []byte(`{}`)
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\n", event.Type); err != nil {
			return false
		}
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
			return false
		}
		flusher.Flush()
		return true
	}

	if !writeEvent(&utils.SSEEvent{
		Type: "connected",
		Data: gin.H{"userId": user.ID, "at": time.Now().UnixMilli()},
	}) {
		return
	}

	pingTicker := time.NewTicker(25 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pingTicker.C:
			if !writeEvent(&utils.SSEEvent{Type: "ping", Data: gin.H{"at": time.Now().UnixMilli()}}) {
				return
			}
		case event, ok := <-ch:
			if !ok {
				return
			}
			if !writeEvent(event) {
				return
			}
		}
	}
}
