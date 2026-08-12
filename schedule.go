package main

import "time"

type clockTime struct {
	hour   int
	minute int
}

func isWakeUpMinute(now time.Time, schedule []clockTime) bool {
	for _, item := range schedule {
		if now.Hour() == item.hour && now.Minute() == item.minute {
			return true
		}
	}
	return false
}

func nextWakeUp(now time.Time, schedule []clockTime) time.Time {
	location := now.Location()
	for dayOffset := 0; dayOffset <= 1; dayOffset++ {
		date := now.AddDate(0, 0, dayOffset)
		for _, item := range schedule {
			candidate := time.Date(date.Year(), date.Month(), date.Day(), item.hour, item.minute, 0, 0, location)
			if candidate.After(now) {
				return candidate
			}
		}
	}
	panic("schedule must contain at least one valid time")
}
