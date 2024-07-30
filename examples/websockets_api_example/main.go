package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ServerAddress is the address of the ComfyUI server.
const ServerAddress = "127.0.0.1:8188"

// ClientID is a unique identifier for the client.
var ClientID = uuid.New().String()

// PromptNode represents a node in the prompt structure.
type PromptNode struct {
	ClassType string `json:"class_type"`
	Inputs    Inputs `json:"inputs"`
}

// Inputs holds the input fields for each node.
type Inputs struct {
	Cfg            int           `json:"cfg,omitempty"`
	Denoise        float64       `json:"denoise,omitempty"`
	LatentImage    []interface{} `json:"latent_image,omitempty"`
	Model          []interface{} `json:"model,omitempty"`
	Negative       []interface{} `json:"negative,omitempty"`
	Positive       []interface{} `json:"positive,omitempty"`
	SamplerName    string        `json:"sampler_name,omitempty"`
	Scheduler      string        `json:"scheduler,omitempty"`
	Seed           int64         `json:"seed,omitempty"`
	Steps          int           `json:"steps,omitempty"`
	CkptName       string        `json:"ckpt_name,omitempty"`
	BatchSize      int           `json:"batch_size,omitempty"`
	Height         int           `json:"height,omitempty"`
	Width          int           `json:"width,omitempty"`
	Clip           []interface{} `json:"clip,omitempty"`
	Text           string        `json:"text,omitempty"`
	FilenamePrefix string        `json:"filename_prefix,omitempty"`
	Images         []interface{} `json:"images,omitempty"`
	Samples        []interface{} `json:"samples,omitempty"`
	Vae            []interface{} `json:"vae,omitempty"`
}

// Prompt is the overall structure representing all nodes.
type Prompt struct {
	Nodes map[string]PromptNode `json:"nodes"`
}

// queuePrompt sends a prompt to the server and returns the response.
func queuePrompt(prompt Prompt) (map[string]interface{}, error) {
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

// getImage retrieves an image from the server based on filename and type.
func getImage(filename, subfolder, folderType string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/view?filename=%s&subfolder=%s&type=%s", ServerAddress, filename, subfolder, folderType)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// getHistory retrieves the history of the prompt execution.
func getHistory(promptID string) (map[string]interface{}, error) {
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

// getImages listens for messages from the WebSocket and processes them.
func getImages(ws *websocket.Conn, prompt Prompt) (map[string][][]byte, error) {
	result, err := queuePrompt(prompt)
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

	history, err := getHistory(promptID)
	if err != nil {
		return nil, err
	}

	for nodeID, nodeOutput := range history[promptID].(map[string]interface{})["outputs"].(map[string]interface{}) {
		if nodeOutputMap, ok := nodeOutput.(map[string]interface{}); ok {
			if images, exists := nodeOutputMap["images"]; exists {
				var imagesOutput [][]byte
				for _, imgData := range images.([]interface{}) {
					imgDetails := imgData.(map[string]interface{})
					image, err := getImage(imgDetails["filename"].(string), imgDetails["subfolder"].(string), imgDetails["type"].(string))
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

// saveImage saves the image data to a file with the specified filename.
func saveImage(data []byte, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Generate a random seed for the prompt
	randomSeed := rng.Int63n(1000000000) // Example range for seed

	// Define the prompt text
	promptText := fmt.Sprintf(`
	{
		"3": {
			"class_type": "KSampler",
			"inputs": {
				"cfg": 8,
				"denoise": 1,
				"latent_image": [
					"5",
					0
				],
				"model": [
					"4",
					0
				],
				"negative": [
					"7",
					0
				],
				"positive": [
					"6",
					0
				],
				"sampler_name": "euler",
				"scheduler": "normal",
				"seed": %d,
				"steps": 20
			}
		},
		"4": {
			"class_type": "CheckpointLoaderSimple",
			"inputs": {
				"ckpt_name": "v1-5-pruned-emaonly.ckpt"
			}
		},
		"5": {
			"class_type": "EmptyLatentImage",
			"inputs": {
				"batch_size": 1,
				"height": 512,
				"width": 512
			}
		},
		"6": {
			"class_type": "CLIPTextEncode",
			"inputs": {
				"clip": [
					"4",
					1
				],
				"text": "masterpiece best quality man"
			}
		},
		"7": {
			"class_type": "CLIPTextEncode",
			"inputs": {
				"clip": [
					"4",
					1
				],
				"text": "bad hands"
			}
		},
		"8": {
			"class_type": "VAEDecode",
			"inputs": {
				"samples": [
					"3",
					0
				],
				"vae": [
					"4",
					2
				]
			}
		},
		"9": {
			"class_type": "SaveImage",
			"inputs": {
				"filename_prefix": "ComfyUI",
				"images": [
					"8",
					0
				]
			}
		}
	}`, randomSeed)

	// Parse the prompt text
	var prompt Prompt
	if err := json.Unmarshal([]byte(promptText), &prompt); err != nil {
		fmt.Println("Error parsing prompt:", err)
		return
	}

	// Establish WebSocket connection
	wsURL := fmt.Sprintf("ws://%s/ws?clientId=%s", ServerAddress, ClientID)
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println("Error connecting to WebSocket:", err)
		return
	}
	defer ws.Close()

	// Get images after executing the prompt
	images, err := getImages(ws, prompt)
	if err != nil {
		fmt.Println("Error getting images:", err)
		return
	}

	// Save output images to local files
	for nodeID, imageDataList := range images {
		for i, imageData := range imageDataList {
			filename := fmt.Sprintf("%s_image_%d.png", nodeID, i) // Create a unique filename
			if err := saveImage(imageData, filename); err != nil {
				fmt.Printf("Error saving image %s: %v\n", filename, err)
			} else {
				fmt.Printf("Saved image: %s\n", filename)
			}
		}
	}
}
