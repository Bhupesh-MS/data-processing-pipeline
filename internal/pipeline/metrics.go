package pipeline

import "time"

type RuntimeMetrics struct {
	StartedAt  time.Time
	FinishedAt time.Time
}

func (m RuntimeMetrics) Duration() time.Duration {
	if m.StartedAt.IsZero() || m.FinishedAt.IsZero() {
		return 0
	}
	return m.FinishedAt.Sub(m.StartedAt)
}
