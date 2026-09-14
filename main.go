package main

import (
	"log"
	"net/http"

	"wxcloudrun-golang/service"

	"github.com/gin-gonic/gin"
)

type OptimizeRequest struct {
	TargetJob  string `json:"targetJob" binding:"required"`
	Experience string `json:"experience" binding:"required"`
}

func main() {
	router := gin.Default()

	// 健康检查接口，供微信云托管监控使用
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "Service is running"})
	})

	// 统一 API 路由组
	api := router.Group("/api")
	{
		api.POST("/optimize", handleOptimize)
	}

	// 启动服务，微信云托管默认监听 80 端口
	if err := router.Run(":80"); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}

func handleOptimize(c *gin.Context) {
	var req OptimizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	// 调用 AI 服务进行润色
	optimizedExp, err := service.OptimizeResume(req.TargetJob, req.Experience)
	if err != nil {
		log.Printf("AI Optimize Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "AI 处理失败"})
		return
	}

	// 返回成功结果
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": map[string]string{
			"optimizedExperience": optimizedExp,
		},
	})
}
