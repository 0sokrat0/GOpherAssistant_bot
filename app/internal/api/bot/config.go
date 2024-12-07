package bot

import (
	"context"
	"errors"
	"time"
)

type Config struct {
	Token   string
	Timeout time.Duration
}

// Service определяет интерфейс для взаимодействия с AI
type Service interface {
	ChatCompletion(ctx context.Context, prompt string) (string, error)
}

var (
	defaultTimeout   = 10 * time.Second
	ErrTokenRequired = errors.New("bot token is required")
)

// NewConfig создает новую конфигурацию бота.
func NewConfig(token string, timeout time.Duration) *Config {
	return &Config{
		Token:   token,
		Timeout: timeout,
	}
}

// Validate проверяет, что конфигурация корректна.
func (c *Config) Validate() error {
	if c.Token == "" {
		return ErrTokenRequired
	}
	if c.Timeout <= 0 {
		c.Timeout = defaultTimeout
	}
	return nil
}
