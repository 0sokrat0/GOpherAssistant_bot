package config

import (
	"flag"
	"log/slog"
	"os"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	App     AppConfig     `yaml:"app"`
	Bot     BotConfig     `yaml:"bot"`
	Metrics MetricsConfig `yaml:"metrics"`
	Tracing TracingConfig `yaml:"tracing"`
}

type AppConfig struct {
	Id        string `yaml:"id" env:"APP_ID"`
	Name      string `yaml:"name" env:"APP_NAME"`
	LogLevel  string `yaml:"logLevel" env:"LOG_LEVEL"`
	IsLogJSON bool   `yaml:"is_log_json" env:"IS_LOG_JSON"`
}

type BotConfig struct {
	Token string `yaml:"token" env:"BOT_TOKEN"`
}

type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" env:"METRICS_ENABLED"`
	Host    string `yaml:"host" env:"METRICS_HOST"`
	Port    int    `yaml:"port" env:"METRICS_PORT"`
}

type TracingConfig struct {
	Enabled bool   `yaml:"enabled" env:"TRACING_ENABLED"`
	Host    string `yaml:"host" env:"TRACING_HOST"`
	Port    int    `yaml:"port" env:"TRACING_PORT"`
}

const (
	flagConfigPathName = "config"
	envConfigPathName  = "CONFIG_PATH"
)

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {

	once.Do(func() {
		var configPath string
		flag.StringVar(&configPath, flagConfigPathName, "", "path to config file")
		flag.Parse()

		if path, ok := os.LookupEnv(envConfigPathName); ok {
			configPath = path
		}

		if readErr := cleanenv.ReadConfig(configPath, instance); readErr != nil {
			description, descrErr := cleanenv.GetDescription(instance, nil)
			if descrErr != nil {
				panic(descrErr)
			}

			slog.Info(description)
			slog.Error("failed to read config", slog.String("err", readErr.Error()),
				slog.String("path", configPath),
			)
			os.Exit(1)
		}

		instance = &Config{}
	})

	return instance
}
