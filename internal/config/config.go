package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		LogLevel string `mapstructure:"log_level"`
	}
	Monitoring struct {
		CheckInterval string `mapstructure:"check_interval"`
		NodesEnabled  bool   `mapstructure:"nodes_enabled"`
		ArgoEnabled   bool   `mapstructure:"argocd_enabled"`
	} `mapstructure:"monitoring"`
	Cluster struct {
		APIURL string `mapstructure:"api_url"`
		Token  string `mapstructure:"token"`
	} `mapstructure:"cluster"`
	Telegram struct {
		Token  string `mapstructure:"token"`
		ChatID int64  `mapstructure:"chat_id"`
		Proxy  string `mapstructure:"proxy"`
	} `mapstructure:"telegram"`
}

func NewConfig() (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	v.AddConfigPath("./internal/app")

	v.SetEnvPrefix("K8SBOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
