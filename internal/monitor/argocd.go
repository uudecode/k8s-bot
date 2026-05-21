package monitor

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (s *Service) WatchArgoApps(ctx context.Context) {
	for {
		if err := s.watchArgoAppsOnce(ctx); err != nil {
			s.Logger.Error().Err(err).Msg("ArgoCD application watcher stopped with error")

			if notifyErr := s.Notifier.Send("🚨 *ArgoCD watcher stopped:*\n`" + err.Error() + "`"); notifyErr != nil {
				s.Logger.Error().Err(notifyErr).Msg("failed to send ArgoCD watcher error notification")
			}
		} else {
			s.Logger.Warn().Msg("ArgoCD application watcher stopped")
		}

		select {
		case <-time.After(5 * time.Second):
			s.Logger.Info().Msg("restarting ArgoCD application watcher")
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) watchArgoAppsOnce(ctx context.Context) error {
	gvr := schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}
	watcher, err := s.Dynamic.Resource(gvr).Namespace(s.Cfg.ArgoCD.Name).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	s.Logger.Info().Msg("starting ArgoCD application watcher")

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}

			app, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}

			status, _ := app.Object["status"].(map[string]interface{})
			healthObj, _ := status["health"].(map[string]interface{})
			health, _ := healthObj["status"].(string)
			if health == "" {
				s.Logger.Warn().
					Str("app_name", app.GetName()).
					Msg("ArgoCD application health status is empty")
				continue
			}

			s.Logger.Debug().
				Str("event_type", string(event.Type)).
				Str("app_name", app.GetName()).
				Str("health", health).
				Msg("Watcher update")

			shouldSend, msg := s.updateStateAndCheck("app:"+app.GetName(), health, "Healthy")
			if shouldSend {
				if notifyErr := s.Notifier.Send("🚨 *ArgoCD Alert:* " + msg); notifyErr != nil {
					s.Logger.Error().Err(notifyErr).Msg("failed to send ArgoCD notification")
				}
			}

		case <-ctx.Done():
			return nil
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
		status, ok := item.Object["status"].(map[string]interface{})
		if !ok {
			s.Logger.Warn().Str("app_name", item.GetName()).Msg("ArgoCD application has no status")
			continue
		}

		syncObj, ok := status["sync"].(map[string]interface{})
		if !ok {
			s.Logger.Warn().Str("app_name", item.GetName()).Msg("ArgoCD application has no sync status")
			continue
		}

		syncStatus, _ := syncObj["status"].(string)
		if syncStatus == "" {
			syncStatus = "unknown"
		}

		s.Logger.Info().
			Str("app_name", item.GetName()).
			Str("sync_status", syncStatus).
			Msg("ArgoCD application found")
	}

	return nil
}
