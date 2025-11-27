package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Debug("Error loading .env file", "error", err)
	}
	gin.SetMode(os.Getenv("GIN_MODE"))
	if gin.IsDebugging() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	slog.Info("find models", slog.Any("models", getModels()))

	// 创建Gin路由
	r := gin.Default()

	// 添加中间件
	r.Use(authMiddleware())

	// 设置路由
	v1 := r.Group("/v1")
	{
		// 聊天完成接口
		v1.POST("/chat/completions", chatCompletionsHandler)
		// 模型列表接口
		v1.GET("/models", modelsHandler)
	}

	// 启动服务器
	port := ":8080"
	if len(os.Getenv("PORT")) > 0 {
		port = os.Getenv("PORT")
	}

	slog.Info("Server starting on port", "port", port)
	slog.Info("Available endpoints:")
	slog.Info("  POST /v1/chat/completions")
	slog.Info("  GET  /v1/models")
	r.Run(port)
}
