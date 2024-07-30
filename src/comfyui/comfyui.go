package comfyui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ServerAddress 定义 ComfyUI 服务器地址（需要传入）
var ServerAddress string

// ClientID 是客户端的唯一标识符
var ClientID = uuid.New().String()

// QueuePrompt 发送提示到 ComfyUI 服务器
func QueuePrompt(prompt Prompt) (map[string]interface{}, error) {
	requestPayload := map[string]interface{}{
		"prompt":    prompt,
		"client_id": ClientID,
	}
	data, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/prompt", ServerAddress), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetImage 根据文件名和类型从服务器获取图像
func GetImage(filename, subfolder, folderType string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/view?filename=%s&subfolder=%s&type=%s", ServerAddress, filename, subfolder, folderType)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// GetHistory 获取提示执行的历史记录
func GetHistory(promptID string) (map[string]interface{}, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/history/%s", ServerAddress, promptID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetImages 监听 WebSocket 消息并处理它们
func GetImages(ws *websocket.Conn, prompt Prompt) (map[string][][]byte, error) {
	result, err := QueuePrompt(prompt)
	if err != nil {
		return nil, err
	}
	promptID := result["prompt_id"].(string)

	outputImages := make(map[string][][]byte)

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			return nil, err
		}

		var msgData map[string]interface{}
		if err := json.Unmarshal(message, &msgData); err != nil {
			continue // Ignore invalid messages
		}

		if msgData["type"] == "executing" {
			data := msgData["data"].(map[string]interface{})
			if data["node"] == nil && data["prompt_id"] == promptID {
				break // Execution is done
			}
		}
	}

	history, err := GetHistory(promptID)
	if err != nil {
		return nil, err
	}

	for nodeID, nodeOutput := range history[promptID].(map[string]interface{})["outputs"].(map[string]interface{}) {
		if nodeOutputMap, ok := nodeOutput.(map[string]interface{}); ok {
			if images, exists := nodeOutputMap["images"]; exists {
				var imagesOutput [][]byte
				for _, imgData := range images.([]interface{}) {
					imgDetails := imgData.(map[string]interface{})
					image, err := GetImage(imgDetails["filename"].(string), imgDetails["subfolder"].(string), imgDetails["type"].(string))
					if err != nil {
						return nil, err
					}
					imagesOutput = append(imagesOutput, image)
				}
				outputImages[nodeID] = imagesOutput
			}
		}
	}

	return outputImages, nil
}
