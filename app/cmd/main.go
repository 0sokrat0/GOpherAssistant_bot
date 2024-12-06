package main

import (
	"context"
	"strconv"

	"github.com/0sokrat0/GOpherAssistant_bot/internal/config"
	"github.com/theartofdevel/logging"
)

func main() {
	// Создаем базовый контекст
	ctx := context.Background()

	// Загружаем конфигурацию
	cfg := config.GetConfig()

	logging.Default().Info(
		"application initialization with configuration",
		logging.StringAttr("app_id", cfg.App.Id),
		logging.StringAttr("app_name", cfg.App.Name),
		logging.StringAttr("log_level", cfg.App.LogLevel),
		logging.BoolAttr("is_log_json", cfg.App.IsLogJSON),
		logging.StringAttr("bot_token", "***"+strconv.Itoa(len(cfg.Bot.Token))),
		logging.StringAttr("metrics_enabled", strconv.FormatBool(cfg.Metrics.Enabled)),
		logging.StringAttr("metrics_host", cfg.Metrics.Host),
		logging.StringAttr("metrics_port", strconv.Itoa(cfg.Metrics.Port)),
		logging.StringAttr("tracing_enabled", strconv.FormatBool(cfg.Tracing.Enabled)),
		logging.StringAttr("tracing_host", cfg.Tracing.Host),
		logging.StringAttr("tracing_port", strconv.Itoa(cfg.Tracing.Port)),
	)

	// Инициализируем логгер на основе конфигурации
	logger := logging.NewLogger(
		logging.WithLevel(logging.LevelInfo.String()),
		logging.WithIsJSON(cfg.App.IsLogJSON),
	)
	ctx = logging.ContextWithLogger(ctx, logger)

}
