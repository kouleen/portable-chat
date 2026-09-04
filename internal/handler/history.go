package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/kouleen/portable-chat/internal/model"
	"github.com/kouleen/portable-chat/internal/service"
	utils "github.com/kouleen/portable-chat/pkg/util"
)

var chatUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ListHistory(c *gin.Context) {
	ctx := c.Request.Context()
	user := utils.GetAuthorization(ctx)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Unauthorized"})
		return
	}
	var req model.CharHistoryQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	if receipt, err := service.MarkConversationRead(ctx, user.ID, req.ContactID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	} else if receipt != nil {
		pushReadReceipt(req.ContactID, user.ID, receipt)
	}
	resp, err := service.ListChatHistory(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "timestamp": time.Now().UnixMilli(), "data": resp})
}

func pushReadReceipt(senderID, readerID int64, receipt *model.ReadReceipt) {
	if receipt == nil {
		return
	}
	if senderClient := utils.GetWebSocketClient(senderID); senderClient != nil {
		_ = senderClient.WriteJSON(model.WsEvent{Type: "read", Data: receipt})
	}
	utils.PublishSSE(senderID, &utils.SSEEvent{
		Type: "read_receipt",
		Data: gin.H{
			"readerId": readerID,
			"receipt":  receipt,
		},
	})
	utils.PublishSSE(readerID, &utils.SSEEvent{
		Type: "contacts_refresh",
		Data: gin.H{
			"contactId": senderID,
		},
	})
}

func pushMessageEvents(message *model.ChatMessage) {
	if message == nil {
		return
	}
	utils.PublishSSE(message.SenderID, &utils.SSEEvent{
		Type: "contacts_refresh",
		Data: gin.H{
			"contactId": message.ReceiverID,
			"messageId": message.ID,
		},
	})
	utils.PublishSSE(message.ReceiverID, &utils.SSEEvent{
		Type: "incoming_message",
		Data: gin.H{
			"contactId":   message.SenderID,
			"contactName": message.SenderName,
			"messageId":   message.ID,
			"content":     message.Content,
			"senderId":    message.SenderID,
			"createTime":  message.CreateTime,
		},
	})
}

func ChatRoom(c *gin.Context) {
	ctx := c.Request.Context()
	user := utils.GetAuthorization(ctx)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": "Unauthorized"})
		return
	}

	var req model.CharHistoryQuery
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "timestamp": time.Now().UnixMilli(), "data": err.Error()})
		return
	}

	conn, err := chatUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := utils.SetWebSocketClient(user.ID, req.ContactID, conn)
	defer func() {
		utils.DeleteWebSocketClient(user.ID, conn)
		_ = conn.Close()
	}()

	if receipt, err := service.MarkConversationRead(ctx, user.ID, req.ContactID); err == nil && receipt != nil {
		pushReadReceipt(req.ContactID, user.ID, receipt)
	}

	_ = client.WriteJSON(model.WsEvent{
		Type: "connected",
		Data: gin.H{
			"userId":    user.ID,
			"contactId": req.ContactID,
		},
	})

	for {
		var incoming model.WsIncomingMessage
		if err := conn.ReadJSON(&incoming); err != nil {
			return
		}

		switch incoming.Type {
		case "", "message":
			message, err := service.SendChatMessage(ctx, req.ContactID, incoming.Content)
			if err != nil {
				_ = client.WriteJSON(model.WsEvent{Type: "error", Data: gin.H{"message": err.Error()}})
				continue
			}
			event := model.WsEvent{Type: "message", Data: message}
			_ = client.WriteJSON(event)
			if receiverClient := utils.GetWebSocketClient(req.ContactID); receiverClient != nil {
				_ = receiverClient.WriteJSON(event)
			}
			pushMessageEvents(message)
		case "read":
			receipt, err := service.MarkConversationRead(ctx, user.ID, req.ContactID)
			if err != nil {
				_ = client.WriteJSON(model.WsEvent{Type: "error", Data: gin.H{"message": err.Error()}})
				continue
			}
			if receipt != nil {
				pushReadReceipt(req.ContactID, user.ID, receipt)
			}
		case "ping":
			_ = client.WriteJSON(model.WsEvent{Type: "pong", Data: gin.H{"at": time.Now().UnixMilli()}})
		default:
			_ = client.WriteJSON(model.WsEvent{Type: "error", Data: gin.H{"message": "unsupported event type"}})
		}
	}
}
