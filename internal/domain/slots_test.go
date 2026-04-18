package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidTime(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"09:00", true},
		{"00:00", true},
		{"23:59", true},
		{"12:30", true},
		{"24:00", false},
		{"23:60", false},
		{"-1:00", false},
		{"9:00", true},
		{"abc", false},
		{"09:xx", false},
		{"", false},
		{"09", false},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			assert.Equal(t, tc.want, IsValidTime(tc.input))
		})
	}
}

func TestIsValidTimeRange(t *testing.T) {
	assert.True(t, IsValidTimeRange("09:00", "18:00"))
	assert.True(t, IsValidTimeRange("00:00", "23:59"))
	assert.False(t, IsValidTimeRange("18:00", "09:00"))
	assert.False(t, IsValidTimeRange("09:00", "09:00"))
	assert.False(t, IsValidTimeRange("bad", "18:00"))
	assert.False(t, IsValidTimeRange("09:00", "bad"))
}

func TestGenerateSlotsForDate_Success(t *testing.T) {
	roomID := uuid.UUID{1}
	schedule := &Schedule{
		RoomID:     roomID,
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}
	date := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	slots, err := GenerateSlotsForDate(schedule, date)
	require.NoError(t, err)
	assert.Len(t, slots, 2)
	assert.Equal(t, roomID, slots[0].RoomID)
}

func TestGenerateSlotsForDate_NotInSchedule(t *testing.T) {
	schedule := &Schedule{
		DaysOfWeek: []int{1, 2, 3, 4, 5},
		StartTime:  "09:00",
		EndTime:    "10:00",
	}
	date := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	slots, err := GenerateSlotsForDate(schedule, date)
	require.NoError(t, err)
	assert.Nil(t, slots)
}

func TestGenerateSlotsForDate_Sunday(t *testing.T) {
	schedule := &Schedule{
		DaysOfWeek: []int{7},
		StartTime:  "10:00",
		EndTime:    "11:00",
	}
	date := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)
	slots, err := GenerateSlotsForDate(schedule, date)
	require.NoError(t, err)
	assert.Len(t, slots, 2)
}

func TestGenerateSlotsForDate_SlotWindowTooSmall(t *testing.T) {
	schedule := &Schedule{
		DaysOfWeek: []int{1},
		StartTime:  "09:00",
		EndTime:    "09:20",
	}
	date := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	slots, err := GenerateSlotsForDate(schedule, date)
	require.NoError(t, err)
	assert.Empty(t, slots)
}

func TestGenerateSlotsForDate_ManySlots(t *testing.T) {
	schedule := &Schedule{
		DaysOfWeek: []int{1},
		StartTime:  "08:00",
		EndTime:    "18:00",
	}
	date := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	slots, err := GenerateSlotsForDate(schedule, date)
	require.NoError(t, err)
	assert.Len(t, slots, 20)
}
