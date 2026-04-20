package monitor

import "time"

type ObjectState struct {
	LastStatus string
	LastAlert  time.Time
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

	if state.LastStatus == healthyStatus || time.Since(state.LastAlert) > s.Cfg.Monitoring.AlertRepeatInterval {
		state.LastStatus = currentStatus
		state.LastAlert = time.Now()
		return true, "❌ " + key + " is experiencing issues (status: " + currentStatus + ")"
	}

	return false, ""
}
