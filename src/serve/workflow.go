package serve

import (
	"fmt"

	"github.com/fimreal/comfyui-api/src/comfyui"
)

// WorkflowInput 是用户输入的工作流结构体
type WorkflowInput struct {
	Nodes map[string]comfyui.PromptNode `json:"nodes"`
}

// CompleteWorkflow 根据需要补全工作流
func CompleteWorkflow(input WorkflowInput) (comfyui.Prompt, error) {
	// 在这里您可以检查输入并补全相应的项
	// 例如，如果 "KSampler" 节点没有指定模型，则可能会使用默认值
	if _, exists := input.Nodes["3"]; !exists {
		return comfyui.Prompt{}, fmt.Errorf("KSampler node is missing")
	}

	// 进一步的逻辑和补全规则...
	var prompt comfyui.Prompt
	prompt.Nodes = input.Nodes
	return prompt, nil
}
