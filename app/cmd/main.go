package main

import (
	"context"
	"fmt"
	"github.com/0sokrat0/GOpherAssistant_bot/internal/service/gpt4all"
	"os"
	"strconv"

	"github.com/0sokrat0/GOpherAssistant_bot/internal/api/bot"
	"github.com/0sokrat0/GOpherAssistant_bot/internal/config"
	logging "github.com/theartofdevel/logging"
)

func main() {
	// Создаем базовый контекст
	ctx := context.Background()

	// Загружаем конфигурацию
	cfg := config.GetConfig()

	// Логируем информацию о конфигурации
	logAppConfig(ctx, cfg)

	// Инициализируем логгер на основе конфигурации
	logger := logging.NewLogger(
		logging.WithLevel(cfg.App.LogLevel),
		logging.WithIsJSON(cfg.App.IsLogJSON),
	)
	ctx = logging.ContextWithLogger(ctx, logger)

	// Запускаем бот
	err := runBot(ctx, cfg)
	if err != nil {
		logging.WithAttrs(ctx, logging.ErrAttr(err)).Error("application failed")
		os.Exit(1)
	}

	logging.L(ctx).Info("application stopped")
	os.Exit(0)
}

func logAppConfig(ctx context.Context, cfg *config.Config) {
	logging.Default().Info(
		"application initialization with configuration",
		logging.StringAttr("app_id", cfg.App.Id),
		logging.StringAttr("app_name", cfg.App.Name),
		logging.StringAttr("log_level", cfg.App.LogLevel),
		logging.BoolAttr("is_log_json", cfg.App.IsLogJSON),
		logging.StringAttr("bot_token", "***"+strconv.Itoa(len(cfg.Bot.Token))),
		logging.StringAttr("gpt4all_token", "***"+strconv.Itoa(len(cfg.GPT4All.Token))),
		logging.StringAttr("gpt4all_url", cfg.GPT4All.URL),
	)
}

func runBot(ctx context.Context, cfg *config.Config) error {
	// Проверяем конфигурацию GPT4All
	if cfg.GPT4All.Token == "" || cfg.GPT4All.URL == "" {
		return fmt.Errorf("gpt4all token or url is missing")
	}

	// Создаем GPT4All сервис
	gpt4allCfg := &gpt4all.Config{
		Token: cfg.GPT4All.Token,
		URL:   cfg.GPT4All.URL,
	}
	gpt4allSvc, err := gpt4all.NewService(gpt4allCfg)
	if err != nil {
		logging.WithAttrs(ctx, logging.ErrAttr(err)).Error("failed to create GPT4All service")
		return err
	}

	// Создаем конфигурацию бота
	botCfg := bot.NewConfig(cfg.Bot.Token, cfg.Bot.Timeout)

	// Создаем обертку для бота
	botWrapper, err := bot.NewWrapper(botCfg, gpt4allSvc)
	if err != nil {
		logging.WithAttrs(ctx, logging.ErrAttr(err)).Error("failed to create bot wrapper")
		return err
	}

	// Запускаем бота
	err = botWrapper.Start(ctx)
	if err != nil {
		logging.WithAttrs(ctx, logging.ErrAttr(err)).Error("bot stopped with error")
		return err
	}

	return nil
}
