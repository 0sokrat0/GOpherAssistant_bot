package cmd

import (
	"context"

	"github.com/0sokrat0/GOpherAssistant_bot/internal/config"
	"github.com/theartofdevel/logging"
)

func main() {
	ctx := context.Background()

	cfg := config.GetConfig()

	logger := logging.NewLogger(
		logging.WithLevel(logging.LevelInfo.String()),
		logging.WithIsJSON(cfg.App.IsLogJSON),
	)
	cfg = logging.ContextWithLogger(ctx)
}
