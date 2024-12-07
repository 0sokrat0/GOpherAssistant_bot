package gpt4all

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Validate проверяет корректность конфигурации.
func (c *Config) Validate() error {
	if c.Token == "" {
		return errors.New("API token is required")
	}
	if c.URL == "" {
		return errors.New("API URL is required")
	}
	return nil
}

// Service представляет клиент для работы с GPT4All API.
type Service struct {
	client *http.Client
	config *Config
}

// NewService создает новый сервис для GPT4All.
func NewService(cfg *Config) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Service{
		client: &http.Client{Timeout: 10 * time.Second},
		config: cfg,
	}, nil
}

// ChatCompletionRequest представляет запрос к GPT4All API.
type ChatCompletionRequest struct {
	Model       string              `json:"model"`
	Messages    []map[string]string `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature float64             `json:"temperature"`
}

// ChatCompletionResponse представляет ответ от GPT4All API.
type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (s *Service) ChatCompletion(ctx context.Context, prompt string) (string, error) {
	// Подготовка тела запроса
	reqBody := ChatCompletionRequest{
		Model:       "gpt-4o-mini",
		Messages:    []map[string]string{{"role": "user", "content": prompt}},
		MaxTokens:   500, // Увеличьте до необходимого значения
		Temperature: 0.7,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Создание HTTP-запроса
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", s.config.URL), bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Установка заголовков
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.Token))
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса
	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status code %d", resp.StatusCode)
	}

	// Чтение тела ответа
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	fmt.Println("Response JSON:", string(respData)) // Отладочный вывод

	// Распаковка ответа
	var respBody ChatCompletionResponse
	if err := json.Unmarshal(respData, &respBody); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Проверка наличия выбора
	if len(respBody.Choices) == 0 {
		return "", errors.New("no choices in API response")
	}

	return respBody.Choices[0].Message.Content, nil
}
