package backoff

import (
	"context"
	"time"
)

// ExponentialBackoff implements an exponential backoff waiting time
type ExponentialBackoff struct {
	currentWaitTime time.Duration
	maxWaitTime     time.Duration
}

// NewExponential creates a new ExponentialBackoff
func NewExponential(initialWaitTime, maxWaitTime time.Duration) ExponentialBackoff {
	return ExponentialBackoff{
		currentWaitTime: initialWaitTime,
		maxWaitTime:     maxWaitTime,
	}
}

// Wait wait the time that's needed.
// Returns quickly if the context is cancelled
func (b *ExponentialBackoff) Wait(ctx context.Context) {
	select {
	case <-ctx.Done(): // wait no longer if context is cancelled
	case <-time.After(b.currentWaitTime):
	}
	// backoff
	b.currentWaitTime = minDuration(b.currentWaitTime*2, b.maxWaitTime)
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
