package monitor

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/uudecode/k8s-bot/internal/config"
	"github.com/uudecode/k8s-bot/internal/notifier"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type ObjectState struct {
	LastStatus string
	LastAlert  time.Time
}

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

	nodes, err := s.K8s.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
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
	err = s.ListArgoApplications()
	if err != nil {
		return err
	}
	if err := s.SendStartupReport(); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to send startup report")
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		s.WatchNodes(ctx)
		return nil
	})

	g.Go(func() error {
		s.WatchArgoApps(ctx)
		return nil
	})

	g.Go(func() error {
		ticker := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-ticker.C:
				_, err := s.K8s.Discovery().ServerVersion()
				if err != nil {
					s.Logger.Warn().Msg("Connection to the cluster lost!")
				}
			case <-ctx.Done():
				return nil
			}
		}
	})

	return g.Wait()
}

func (s *Service) ListArgoApplications() error {
	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	list, err := s.Dynamic.Resource(gvr).Namespace("argocd").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to list ArgoCD applications")
		return err
	}

	for _, item := range list.Items {
		s.Logger.Info().
			Str("app_name", item.GetName()).
			Str("sync_status", item.Object["status"].(map[string]interface{})["sync"].(map[string]interface{})["status"].(string)).
			Msg("ArgoCD application found")
	}

	return nil
}

func (s *Service) WatchNodes(ctx context.Context) {
	watcher, err := s.K8s.CoreV1().Nodes().Watch(ctx, metav1.ListOptions{})
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to create node watcher")
		return
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		node, ok := event.Object.(*corev1.Node)
		if !ok {
			continue
		}

		if event.Type == "MODIFIED" {
			for _, cond := range node.Status.Conditions {
				if cond.Type == "Ready" {
					shouldSend, msg := s.updateStateAndCheck("node:"+node.Name, string(cond.Status), "True")
					if shouldSend {
						s.Notifier.Send(msg)
					}
				}
			}
		}
	}
}

func (s *Service) WatchArgoApps(ctx context.Context) {
	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	watcher, err := s.Dynamic.Resource(gvr).Namespace("argocd").Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return
	}
	defer watcher.Stop()
	s.Logger.Info().
		Msg("Starting ArgoCD application watcher")
	for event := range watcher.ResultChan() {
		app := event.Object.(*unstructured.Unstructured)
		status, _ := app.Object["status"].(map[string]interface{})
		healthObj, _ := status["health"].(map[string]interface{})
		health := healthObj["status"].(string)
		s.Logger.Info().
			Str("event_type", string(event.Type)).
			Str("app_name", app.GetName()).
			Str("health", health).
			Msg("Watcher update")

		shouldSend, msg := s.updateStateAndCheck("app:"+app.GetName(), health, "Healthy")
		if shouldSend {
			s.Notifier.Send("🚨 *ArgoCD Alert:* " + msg)
		}
	}
}

func (s *Service) updateStateAndCheck(key, currentStatus, healthyStatus string) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.objectStates[key]
	if !exists {
		s.objectStates[key] = &ObjectState{LastStatus: currentStatus, LastAlert: time.Time{}}
		return false, ""
	}

	if currentStatus == healthyStatus {
		if state.LastStatus != healthyStatus {
			state.LastStatus = currentStatus
			return true, "✅ " + key + " recovered (status: " + currentStatus + ")"
		}
		return false, ""
	}

	if state.LastStatus == healthyStatus || time.Since(state.LastAlert) > 5*time.Minute {
		state.LastStatus = currentStatus
		state.LastAlert = time.Now()
		return true, "❌ " + key + " is experiencing issues (status: " + currentStatus + ")"
	}

	return false, ""
}

func (s *Service) SendStartupReport() error {
	var report string
	report += "🚀 *Bot started and ready!*\n\n"

	nodes, err := s.K8s.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	report += "🖥 *Node status:*\n"
	for _, node := range nodes.Items {
		status := "✅ Ready"
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status != "True" {
				status = "❌ NotReady"
			}
		}
		report += "- " + node.Name + ": " + status + "\n"
	}

	report += "\n📦 *ArgoCD applications:*\n"
	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	list, err := s.Dynamic.Resource(gvr).Namespace("argocd").List(context.Background(), metav1.ListOptions{})
	if err == nil {
		for _, item := range list.Items {
			syncStatus := item.Object["status"].(map[string]interface{})["sync"].(map[string]interface{})["status"].(string)
			report += "- " + item.GetName() + ": " + syncStatus + "\n"
		}
	}

	return s.Notifier.Send(report)
}
