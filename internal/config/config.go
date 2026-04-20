package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		LogLevel string `mapstructure:"log_level"`
	} `mapstructure:"app"`
	Monitoring struct {
		APIProbeInterval    time.Duration `mapstructure:"api_probe_interval"`
		AlertRepeatInterval time.Duration `mapstructure:"alert_repeat_interval"`
		NodesEnabled        bool          `mapstructure:"nodes_enabled"`
		ArgoEnabled         bool          `mapstructure:"argocd_enabled"`
	} `mapstructure:"monitoring"`
	ArgoCD struct {
		Name string `mapstructure:"name"`
	} `mapstructure:"argocd"`
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
	v.AddConfigPath("/etc/k8s-bot")

	v.SetEnvPrefix("K8SBOT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	if cfg.Monitoring.APIProbeInterval <= 0 {
		cfg.Monitoring.APIProbeInterval = time.Minute
	}
	if cfg.Monitoring.AlertRepeatInterval <= 0 {
		cfg.Monitoring.AlertRepeatInterval = 5 * time.Minute
	}
	return &cfg, nil
}
