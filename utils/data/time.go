package data

import "time"

func updateConfigTime(createdAt *time.Time, updatedAt *time.Time) {
	now := time.Now()
	if createdAt.IsZero() {
		*createdAt = now
	}
	*updatedAt = now
}
