package comfyui

// PromptNode 表示提示节点的结构
type PromptNode struct {
	ClassType string `json:"class_type"`
	Inputs    Inputs `json:"inputs"`
}

// Inputs 包含每个节点的输入字段
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

// Prompt 是表示所有节点的整体结构
type Prompt struct {
	Nodes map[string]PromptNode `json:"nodes"`
}
