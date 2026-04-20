package monitor

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
