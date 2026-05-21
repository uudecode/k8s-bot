package monitor

import (
	"context"
	"time"
)

func (s *Service) RunClusterProbe(ctx context.Context) error {
	ticker := time.NewTicker(s.Cfg.Monitoring.APIProbeInterval)
	defer ticker.Stop()

	s.checkClusterConnection(ctx)

	for {
		select {
		case <-ticker.C:
			s.checkClusterConnection(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) checkClusterConnection(ctx context.Context) {
	_, err := s.K8s.Discovery().ServerVersion()
	if err != nil {
		s.Logger.Error().Err(err).Msg("connection to the cluster lost")

		shouldSend, msg := s.updateStateAndCheck("cluster:api", "unreachable", "reachable")
		if shouldSend {
			if notifyErr := s.Notifier.Send("🚨 *Kubernetes API Alert:*\n" + msg + "\n\nError: `" + err.Error() + "`"); notifyErr != nil {
				s.Logger.Error().Err(notifyErr).Msg("failed to send cluster connection alert")
			}
		}

		return
	}

	shouldSend, msg := s.updateStateAndCheck("cluster:api", "reachable", "reachable")
	if shouldSend {
		s.Logger.Info().Msg("connection to the cluster restored")

		if notifyErr := s.Notifier.Send("✅ *Kubernetes API recovered:*\n" + msg); notifyErr != nil {
			s.Logger.Error().Err(notifyErr).Msg("failed to send cluster recovery notification")
		}
	}
}
