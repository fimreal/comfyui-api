package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

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
	Text           string        `json:"text,omitempty"` // Text input field
	FilenamePrefix string        `json:"filename_prefix,omitempty"`
	Images         []interface{} `json:"images,omitempty"`
	Samples        []interface{} `json:"samples,omitempty"`
	Vae            []interface{} `json:"vae,omitempty"`
}

// Prompt is the overall structure representing all nodes.
type Prompt struct {
	Nodes map[string]PromptNode `json:"nodes"`
}

// API endpoint and configuration constants
const (
	apiEndpoint           = "http://127.0.0.1:6006/prompt"
	defaultCfg            = 8
	defaultDenoise        = 1.0
	defaultSamplerName    = "euler"
	defaultScheduler      = "normal"
	defaultSeedMin        = int64(0)                   // Changed to int64
	defaultSeedMax        = int64(9223372036854775807) // Maximum int64 value
	defaultSteps          = 20
	defaultBatchSize      = 1
	defaultHeight         = 512
	defaultWidth          = 512
	defaultCkptName       = "majicmixRealistic_v7.safetensors"
	defaultFilenamePrefix = "ComfyUI"
)

// queuePrompt sends the prompt to the defined API endpoint.
func queuePrompt(prompt Prompt) error {
	data, err := json.Marshal(prompt)
	if err != nil {
		return fmt.Errorf("failed to marshal prompt: %w", err)
	}

	resp, err := http.Post(apiEndpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to send prompt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}
	return nil
}

// newPrompt initializes a new Prompt with configurable values.
func newPrompt(rng *rand.Rand, text string) Prompt {
	return Prompt{
		Nodes: map[string]PromptNode{
			"3": {
				ClassType: "KSampler",
				Inputs: Inputs{
					Cfg:         defaultCfg,
					Denoise:     defaultDenoise,
					LatentImage: []interface{}{"5", 0},
					Model:       []interface{}{"4", 0},
					Negative:    []interface{}{"7", 0},
					Positive:    []interface{}{"6", 0},
					SamplerName: defaultSamplerName,
					Scheduler:   defaultScheduler,
					Seed:        rng.Int63n(defaultSeedMax), // Generate seed directly in range [0, defaultSeedMax)
					Steps:       defaultSteps,
				},
			},
			"4": {
				ClassType: "CheckpointLoaderSimple",
				Inputs: Inputs{
					CkptName: defaultCkptName,
				},
			},
			"5": {
				ClassType: "EmptyLatentImage",
				Inputs: Inputs{
					BatchSize: defaultBatchSize,
					Height:    defaultHeight,
					Width:     defaultWidth,
				},
			},
			"6": {
				ClassType: "CLIPTextEncode",
				Inputs: Inputs{
					Clip: []interface{}{"4", 1},
					Text: text, // Use the provided text
				},
			},
			"7": {
				ClassType: "CLIPTextEncode",
				Inputs: Inputs{
					Clip: []interface{}{"4", 1},
					Text: "bad hands",
				},
			},
			"8": {
				ClassType: "VAEDecode",
				Inputs: Inputs{
					Samples: []interface{}{"3", 0},
					Vae:     []interface{}{"4", 2},
				},
			},
			"9": {
				ClassType: "SaveImage",
				Inputs: Inputs{
					FilenamePrefix: defaultFilenamePrefix,
					Images:         []interface{}{"8", 0},
				},
			},
		},
	}
}

func main() {
	// Create a new random number generator with a seed based on the current time
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get user input text
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <text>")
		return
	}
	text := os.Args[1]

	prompt := newPrompt(rng, text)

	// Send the prompt to the API
	if err := queuePrompt(prompt); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Prompt queued successfully!")
	}
}
