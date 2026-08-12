package main

import (
	"testing"
	"time"
)

func TestNextWakeUp(t *testing.T) {
	location := time.FixedZone("test", 2*60*60)
	tests := []struct {
		name string
		now  string
		want string
	}{
		{"before first", "2026-08-12 04:00", "2026-08-12 05:30"},
		{"between morning slots", "2026-08-12 07:00", "2026-08-12 10:30"},
		{"exact time moves ahead", "2026-08-12 10:30", "2026-08-12 15:30"},
		{"after last", "2026-08-12 23:00", "2026-08-13 05:30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now, _ := time.ParseInLocation("2006-01-02 15:04", tt.now, location)
			want, _ := time.ParseInLocation("2006-01-02 15:04", tt.want, location)
			if got := nextWakeUp(now, wakeUpTimes); !got.Equal(want) {
				t.Fatalf("nextWakeUp(%s) = %s, want %s", now, got, want)
			}
		})
	}
}

func TestScheduleHasNo0130(t *testing.T) {
	for _, item := range wakeUpTimes {
		if item.hour == 1 && item.minute == 30 {
			t.Fatal("01:30 must be skipped")
		}
	}
}
