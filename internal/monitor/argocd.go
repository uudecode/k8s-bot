package monitor

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

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
		s.Logger.Debug().
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

func (s *Service) ListArgoApplications(ctx context.Context) error {
	gvr := schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}

	list, err := s.Dynamic.Resource(gvr).Namespace(s.Cfg.ArgoCD.Name).List(ctx, metav1.ListOptions{})
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
