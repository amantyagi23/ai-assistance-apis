package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amantyagi23/ai-assistance/internal/config"
)

type GenerateRequest struct {
	Model    string                  `json:"model"`
	Prompt   string                  `json:"prompt"`
	Stream   *bool                   `json:"stream"`
	Options  *map[string]interface{} `json:"options,omitempty"`
	System   *string                 `json:"system,omitempty"`
	Context  *[]int                  `json:"context,omitempty"`
	Template *string                 `json:"template,omitempty"`
}

type GenerateResponse struct {
	Model              string `json:"model"`
	CreatedAt          string `json:"created_at"`
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	Context            []int  `json:"context"`
	TotalDuration      int64  `json:"total_duration"`
	LoadDuration       int64  `json:"load_duration"`
	PromptEvalCount    int    `json:"prompt_eval_count"`
	PromptEvalDuration int64  `json:"prompt_eval_duration"`
	EvalCount          int    `json:"eval_count"`
	EvalDuration       int64  `json:"eval_duration"`
}

func OllamaService(req GenerateRequest) (GenerateResponse, error) {

	var result GenerateResponse

	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return result, err
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	var url string = config.APPConfig().OllamaHostPath + "/api/generate"
	// Send POST request to Ollama
	resp, err := client.Post(
		url,
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return result, err
	}

	defer resp.Body.Close()

	// Decode response JSON
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, err
	}

	fmt.Print(result)
	return result, nil
}
