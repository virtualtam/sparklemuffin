// Copyright (c) VirtualTam
// SPDX-License-Identifier: MIT

package synchronizing

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/ksuid"
)

const (
	defaultSynchronizationInterval = 1 * time.Hour
	defaultTaskTimeout             = 5 * time.Minute
)

// A Scheduler periodically synchronizes all syndication feeds.
type Scheduler struct {
	s           *Service
	locker      sync.Locker
	interval    time.Duration
	taskTimeout time.Duration
}

// NewScheduler initializes and returns a Scheduler.
func NewScheduler(service *Service, locker sync.Locker) *Scheduler {
	return &Scheduler{
		s:           service,
		locker:      locker,
		interval:    defaultSynchronizationInterval,
		taskTimeout: defaultTaskTimeout,
	}
}

// Run periodically synchronizes all syndication feeds.
func (sc *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(sc.interval)
	log.Info().Dur("interval_seconds", sc.interval).Msg("feeds: synchronization scheduler started")

	for {
		<-ticker.C

		go func() {
			jobID := ksuid.New().String()

			sc.locker.Lock()
			defer sc.locker.Unlock()

			taskCtx, cancel := context.WithTimeout(ctx, sc.taskTimeout)
			defer cancel()

			if err := sc.s.Synchronize(taskCtx, jobID); err != nil {
				log.
					Error().
					Err(err).
					Str("job_id", jobID).
					Msg("feeds: failed to synchronize data")
			}
		}()
	}
}
