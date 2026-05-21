package monitor

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Service) WatchNodes(ctx context.Context) {
	for {
		if err := s.watchNodesOnce(ctx); err != nil {
			s.Logger.Error().Err(err).Msg("node watcher stopped with error")

			if notifyErr := s.Notifier.Send("🚨 *Node watcher stopped:*\n`" + err.Error() + "`"); notifyErr != nil {
				s.Logger.Error().Err(notifyErr).Msg("failed to send node watcher error notification")
			}
		} else {
			s.Logger.Warn().Msg("node watcher stopped")
		}

		select {
		case <-time.After(5 * time.Second):
			s.Logger.Info().Msg("restarting node watcher")
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) watchNodesOnce(ctx context.Context) error {
	watcher, err := s.K8s.CoreV1().Nodes().Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	defer watcher.Stop()

	s.Logger.Info().Msg("starting node watcher")

	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return nil
			}

			node, ok := event.Object.(*corev1.Node)
			if !ok {
				continue
			}

			if event.Type == "MODIFIED" {
				for _, cond := range node.Status.Conditions {
					if cond.Type == "Ready" {
						shouldSend, msg := s.updateStateAndCheck("node:"+node.Name, string(cond.Status), "True")
						if shouldSend {
							if notifyErr := s.Notifier.Send(msg); notifyErr != nil {
								s.Logger.Error().Err(notifyErr).Msg("failed to send node notification")
							}
						}
					}
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}
