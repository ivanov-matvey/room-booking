package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func IsValidTime(t string) bool {
	parts := strings.SplitN(t, ":", 2)
	if len(parts) != 2 {
		return false
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return h >= 0 && h <= 23 && m >= 0 && m <= 59
}

func GenerateSlotsForDate(schedule *Schedule, date time.Time) ([]Slot, error) {
	weekday := int(date.Weekday())
	dow := weekday
	if dow == 0 {
		dow = 7
	}

	found := false
	for _, d := range schedule.DaysOfWeek {
		if d == dow {
			found = true
			break
		}
	}
	if !found {
		return nil, nil
	}

	startH, startM, err := parseHHMM(schedule.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %w", err)
	}
	endH, endM, err := parseHHMM(schedule.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %w", err)
	}

	var slots []Slot
	current := time.Date(date.Year(), date.Month(), date.Day(), startH, startM, 0, 0, time.UTC)
	end := time.Date(date.Year(), date.Month(), date.Day(), endH, endM, 0, 0, time.UTC)

	for current.Before(end) {
		next := current.Add(30 * time.Minute)
		if next.After(end) {
			break
		}
		slots = append(slots, Slot{
			ID:        uuid.New(),
			RoomID:    schedule.RoomID,
			StartTime: current,
			EndTime:   next,
		})
		current = next
	}
	return slots, nil
}

func IsValidTimeRange(start, end string) bool {
	sh, sm, err1 := parseHHMM(start)
	eh, em, err2 := parseHHMM(end)
	if err1 != nil || err2 != nil {
		return false
	}
	return sh*60+sm < eh*60+em
}

func parseHHMM(t string) (int, int, error) {
	parts := strings.SplitN(t, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %q", t)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hours in %q: %w", t, err)
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minutes in %q: %w", t, err)
	}
	return h, m, nil
}
