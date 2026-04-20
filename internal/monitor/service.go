package monitor

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
	"github.com/uudecode/k8s-bot/internal/config"
	"github.com/uudecode/k8s-bot/internal/notifier"
	"golang.org/x/sync/errgroup"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Service struct {
	Cfg          *config.Config
	Logger       zerolog.Logger
	K8s          *kubernetes.Clientset
	Dynamic      dynamic.Interface
	Notifier     notifier.Notifier
	mu           sync.RWMutex
	objectStates map[string]*ObjectState
}

func NewMonitorService(cfg *config.Config, log zerolog.Logger, k8s *kubernetes.Clientset, dynamic dynamic.Interface, notifier notifier.Notifier) *Service {
	return &Service{
		Cfg:          cfg,
		Logger:       log,
		K8s:          k8s,
		Dynamic:      dynamic,
		Notifier:     notifier,
		mu:           sync.RWMutex{},
		objectStates: make(map[string]*ObjectState, 1024),
	}
}

func (s *Service) Run(ctx context.Context) error {
	s.Logger.Info().Msg("Bot started, starting cluster check...")

	nodes, err := s.K8s.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to get node list")
		return err
	}
	for _, node := range nodes.Items {
		s.Logger.Info().
			Str("node_name", node.Name).
			Str("version", node.Status.NodeInfo.KubeletVersion).
			Msg("Node found")
	}
	err = s.ListArgoApplications(ctx)
	if err != nil {
		return err
	}
	if err := s.SendStartupReport(); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to send startup report")
	}

	g, ctx := errgroup.WithContext(ctx)

	if s.Cfg.Monitoring.NodesEnabled {
		g.Go(func() error {
			s.WatchNodes(ctx)
			return nil
		})
	}

	if s.Cfg.Monitoring.ArgoEnabled {
		g.Go(func() error {
			s.WatchArgoApps(ctx)
			return nil
		})
	}

	g.Go(func() error {
		return s.RunClusterProbe(ctx)
	})

	return g.Wait()
}
