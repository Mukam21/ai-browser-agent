package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIClient struct {
	apiKey    string
	model     string
	baseURL   string
	client    *http.Client
	maxTokens int
}

func NewOpenAIClient(apiKey, model string, maxTokens int) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	if maxTokens == 0 {
		maxTokens = 1000
	}

	return &OpenAIClient{
		apiKey:    apiKey,
		model:     model,
		baseURL:   "https://api.openai.com/v1",
		maxTokens: maxTokens,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *OpenAIClient) GetModel() string {
	return c.model
}

func (c *OpenAIClient) Chat(ctx context.Context, prompt string) (string, error) {
	request := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": "Ты - автономный AI агент для управления браузером. Ты умеешь анализировать веб-страницы и планировать действия.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.1,
		"max_tokens":  c.maxTokens,
	}

	respBody, err := c.makeRequest("chat/completions", request)
	if err != nil {
		return "", fmt.Errorf("OpenAI API error: %w", err)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("OpenAI error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	content := result.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	return content, nil
}

func (c *OpenAIClient) ChatStructured(ctx context.Context, prompt string, output interface{}) error {
	systemMsg := `Ты - автономный AI агент. ВСЕГДА отвечай ТОЛЬКО в формате JSON без дополнительного текста.
Формат ответа должен быть валидным JSON.`

	fullPrompt := fmt.Sprintf(`%s

%s`, systemMsg, prompt)

	response, err := c.Chat(ctx, fullPrompt)
	if err != nil {
		return err
	}

	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return fmt.Errorf("не удалось извлечь JSON из ответа: %s", response)
	}

	if err := json.Unmarshal([]byte(jsonStr), output); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w, response: %s", err, jsonStr)
	}

	return nil
}

func (c *OpenAIClient) makeRequest(endpoint string, data interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+"/"+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func extractJSON(text string) string {
	start := strings.Index(text, "{")
	if start == -1 {
		start = strings.Index(text, "[")
		if start == -1 {
			return ""
		}
	}

	braceCount := 0
	bracketCount := 0
	inString := false
	escape := false

	for i := start; i < len(text); i++ {
		ch := text[i]

		if escape {
			escape = false
			continue
		}

		if ch == '\\' {
			escape = true
			continue
		}

		if ch == '"' && !escape {
			inString = !inString
			continue
		}

		if !inString {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 && bracketCount == 0 {
					return text[start : i+1]
				}
			} else if ch == '[' {
				bracketCount++
			} else if ch == ']' {
				bracketCount--
				if bracketCount == 0 && braceCount == 0 {
					return text[start : i+1]
				}
			}
		}
	}

	return text[start:]
}
