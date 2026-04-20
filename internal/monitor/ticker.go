package monitor

import (
	"context"
	"time"
)

func (s *Service) RunClusterProbe(ctx context.Context) error {
	ticker := time.NewTicker(s.Cfg.Monitoring.APIProbeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, err := s.K8s.Discovery().ServerVersion()
			if err != nil {
				s.Logger.Warn().Err(err).Msg("connection to the cluster lost")
			}
		case <-ctx.Done():
			return nil
		}
	}
}
