package di

import (
	"fmt"
	"net/http"
	"net/url"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/wire"
	"github.com/uudecode/k8s-bot/internal/config"
	"github.com/uudecode/k8s-bot/internal/logger"
	"github.com/uudecode/k8s-bot/internal/monitor"
	"github.com/uudecode/k8s-bot/internal/notifier"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var SuperSet = wire.NewSet(
	config.NewConfig,
	logger.NewLogger,
	ProvideK8sClient,
	ProvideBot,
	ProvideDynamicClient,
	monitor.NewMonitorService,
	ProvideNotifier,
)

func ProvideK8sClient(cfg *config.Config) (*kubernetes.Clientset, error) {
	client, err := kubernetes.NewForConfig(&rest.Config{
		Host:            cfg.Cluster.APIURL,
		BearerToken:     cfg.Cluster.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	})
	if err != nil {
		return nil, fmt.Errorf("kubernetes: create clientset: %w", err)
	}
	return client, nil
}

func ProvideBot(cfg *config.Config) (*tgbotapi.BotAPI, error) {
	client := &http.Client{}
	if cfg.Telegram.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Telegram.Proxy)
		if err != nil {
			return nil, fmt.Errorf("telegram: parse proxy url: %w", err)
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}
	bot, err := tgbotapi.NewBotAPIWithClient(cfg.Telegram.Token, tgbotapi.APIEndpoint, client)
	if err != nil {
		return nil, fmt.Errorf("telegram: create bot client: %w", err)
	}
	return bot, nil
}

func ProvideDynamicClient(cfg *config.Config) (dynamic.Interface, error) {
	restCfg := &rest.Config{
		Host:            cfg.Cluster.APIURL,
		BearerToken:     cfg.Cluster.Token,
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
	}
	return dynamic.NewForConfig(restCfg)
}

func ProvideNotifier(bot *tgbotapi.BotAPI, cfg *config.Config) notifier.Notifier {
	return &notifier.TelegramNotifier{
		Bot:    bot,
		ChatID: cfg.Telegram.ChatID,
	}
}
