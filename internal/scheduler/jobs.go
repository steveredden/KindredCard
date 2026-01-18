/*
 * Copyright (C) 2026 Steve Redden
 *
 * KindredCard is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 */

package scheduler

import (
	"time"

	"github.com/steveredden/KindredCard/internal/logger"
)

func (s *Scheduler) globalCleanup(runNow bool) {
	now := time.Now()
	currentTime := now.Format("15:04") // HH:MM format

	// cleanup at 11:59PM
	if !runNow && currentTime != "23:59" {
		return
	}

	logger.Info("[SCHEDULER] Performing Global Clean-up at %s", currentTime)

	logger.Info("[SCHEDULER] Deleting expired sessions")
	s.db.CleanupExpiredSessions()

	logger.Info("[SCHEDULER] Global Clean-up complete!")
}
