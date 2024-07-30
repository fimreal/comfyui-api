package serve

import (
	"encoding/json"
	"net/http"

	"github.com/fimreal/comfyui-api/src/comfyui"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// showIndexPage 渲染首页
func showIndexPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

// processWorkflow 处理工作流请求
func processWorkflow(c *gin.Context) {
	var workflow struct {
		Workflow string `json:"workflow"`
		Server   string `json:"server"` // 新增字段，用于接收服务器地址
	}

	if err := c.ShouldBindJSON(&workflow); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置 ComfyUI 服务器地址
	comfyui.ServerAddress = workflow.Server // 从请求中获取服务器地址

	// 解析工作流 JSON
	var prompt comfyui.Prompt
	if err := json.Unmarshal([]byte(workflow.Workflow), &prompt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workflow JSON"})
		return
	}

	// 创建 WebSocket 连接
	wsURL := "ws://" + comfyui.ServerAddress + "/ws?clientId=" + comfyui.ClientID
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to WebSocket: " + err.Error()})
		return
	}
	defer ws.Close()

	// 使用 WebSocket 和轮询获取图像
	outputImages, err := comfyui.GetImages(ws, prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Workflow processed successfully!",
		"output":  outputImages,
	})
}
