package serve

import (
	"github.com/gin-gonic/gin"
)

// StartServer 启动 Gin 服务器
func StartServer() error {
	r := gin.Default()
	r.Static("/static", "./src/templates/static") // 访问静态资源
	r.LoadHTMLGlob("src/templates/*")

	r.GET("/", showIndexPage)

	// 设置处理工作流请求的API端点
	r.POST("/api/process", processWorkflow)

	return r.Run(":8080")
}
