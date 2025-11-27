package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var Models map[string]string

func getModels() map[string]string {
	if Models != nil {
		return Models
	}
	Models = make(map[string]string)
	models := os.Getenv("MODELS")
	if len(models) == 0 {
		slog.Error("environment variable MODELS is not set.\neg: export MODELS=name1=https://xxx//webhook/xxx/chat;name2=https://xxx//webhook/xxx/chat")
		os.Exit(-1)
	}
	for _, line := range strings.Split(models, ";") {
		fields := strings.Split(line, "=")
		if len(fields) != 2 {
			slog.Warn("skip invalid model config", slog.String("model", line))
			continue
		}
		Models[fields[0]] = fields[1]
	}
	if len(Models) == 0 {
		slog.Error("environment variable MODELS is invalid format.\neg: export MODELS=name1=https://xxx//webhook/xxx/chat;name2=https://xxx//webhook/xxx/chat")
		os.Exit(-1)
	}
	return Models
}

// 模型列表接口
func modelsHandler(c *gin.Context) {
	slog.Debug("modelsHandler")
	response := ModelsResponse{
		Object: "list",
		Data:   nil,
	}
	for model := range getModels() {
		response.Data = append(response.Data, Model{
			ID:      model,
			Object:  "model",
			Created: time.Now().Unix(),
		})
	}
	c.JSON(http.StatusOK, response)
}
