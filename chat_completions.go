package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 聊天完成接口
func chatCompletionsHandler(c *gin.Context) {
	slog.Debug("chatCompletionsHandler")
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("chatCompletionsHandler", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	slog.Debug("chatCompletionsHandler", "req", req)

	models := getModels()
	if len(models[req.Model]) == 0 {
		slog.Error("not found model", "model", req.Model)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model name"})
		return
	}
	// 检查是否请求流式响应
	isStream := req.Stream != nil && *req.Stream
	if isStream {
		handleStreamChatCompletion(c, req)
	} else {
		handleNonStreamChatCompletion(c, req)
	}
}

// 处理非流式聊天完成
func handleNonStreamChatCompletion(c *gin.Context, req ChatCompletionRequest) {
	content := ""
	err := n8nChat(req, func(msg N8NChatItem) error {
		content += msg.Content
		return nil
	})
	if err != nil {
		slog.Error("Error exec n8n Chat", "error", err)
		return
	}
	response := ChatCompletionResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{
			{
				Index: 0,
				Message: ChoiceMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

func n8nChat(req ChatCompletionRequest, onMessage func(msg N8NChatItem) error) error {
	var n8nAction struct {
		Action    string `json:"action"`
		SessionID string `json:"sessionId"`
		ChatInput string `json:"chatInput"`
	}
	n8nAction.Action = "sendMessage"
	n8nAction.SessionID = req.User
	var strBuilder strings.Builder
	for _, msg := range req.Messages {
		for _, content := range msg.Content {
			if msg.Role != "user" {
				slog.Warn("ignore message", "role", msg.Role, "message", content.Text)
				continue
			}
			strBuilder.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, content.Text))
		}
	}
	n8nAction.ChatInput = strBuilder.String()
	n8nAction.ChatInput = req.Messages[len(req.Messages)-1].Content[len(req.Messages[len(req.Messages)-1].Content)-1].Text
	slog.Debug("n8nAction", "action", n8nAction)
	n8nBody, err := json.Marshal(n8nAction)
	if err != nil {
		slog.Error("marshal n8n body", "error", err)
		return err
	}
	n8nReq, err := http.NewRequest(http.MethodPost, "https://n8n.cicd.getdeepin.org/webhook/d58ddd17-1d51-4dd1-a5a5-a9fa8d4a81c3/chat", bytes.NewReader(n8nBody))
	if err != nil {
		slog.Error("Error new http request", "error", err)
		return err
	}
	n8nReq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(n8nReq)
	if err != nil {
		slog.Error("Error send http request", "error", err)
		return err
	}
	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	var item N8NChatItem

	for decoder.More() {
		err = decoder.Decode(&item)
		if err != nil {
			slog.Error("Error decode chat item", "error", err)
			return err
		}
		slog.Debug("n8n item", slog.Any("item", item))
		if item.Type != "item" {
			continue
		}
		// Respond to Webhook 流模式只能返回json数据，所以要解析嵌套的json字符串
		data := []byte(item.Content)
		if json.Valid(data) {
			var webhookItem N8NChatItem
			err = json.Unmarshal(data, &webhookItem)
			if err == nil {
				item = webhookItem
			}
		}
		err = onMessage(item)
		if err != nil {
			return err
		}
	}
	return nil
}

// 处理流式聊天完成
func handleStreamChatCompletion(c *gin.Context, req ChatCompletionRequest) {
	streamID := uuid.New().String()
	// 设置响应头为流式传输
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 发送流式响应
	created := time.Now().Unix()
	// 发送开始事件
	startEvent := ChatCompletionStreamResponse{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []StreamChoice{
			{
				Index: 0,
				Delta: StreamDelta{
					Role: stringPtr("assistant"),
				},
			},
		},
	}
	err := sendStreamEvent(c.Writer, startEvent)
	if err != nil {
		slog.Error("Error sending start event", "error", err)
		return
	}

	var output string
	err = n8nChat(req, func(item N8NChatItem) error {
		// 发送内容块
		contentEvent := ChatCompletionStreamResponse{
			ID:      streamID,
			Object:  "chat.completion.chunk",
			Created: created,
			Model:   req.Model,
			Choices: []StreamChoice{
				{
					Index: 0,
					Delta: StreamDelta{
						Content: &item.Content,
					},
				},
			},
		}
		output += item.Content
		if err := sendStreamEvent(c.Writer, contentEvent); err != nil {
			slog.Error("Error sending content event", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		slog.Error("Error exec n8n Chat", "error", err)
		return
	}

	// 流结束，发送完成事件
	finishReason := "stop"
	endEvent := ChatCompletionStreamResponse{
		ID:      streamID,
		Object:  "chat.completion.chunk",
		Created: created,
		Model:   req.Model,
		Choices: []StreamChoice{
			{
				Index:        0,
				FinishReason: &finishReason,
			},
		},
	}
	sendStreamEvent(c.Writer, endEvent)
	slog.Info("chat", "output", output)
}

type N8NChatItem struct {
	Type     string
	Content  string
	Metadata map[string]interface{}
}

// 发送流式事件
func sendStreamEvent(w gin.ResponseWriter, event ChatCompletionStreamResponse) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// 发送SSE格式的数据
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	if err != nil {
		return err
	}

	// 刷新缓冲区
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	return nil
}

// 辅助函数：创建字符串指针
func stringPtr(s string) *string {
	return &s
}
