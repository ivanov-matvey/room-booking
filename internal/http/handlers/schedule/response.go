package schedulehandler

import "github.com/google/uuid"

// ScheduleResponse is the API representation of a schedule.
type ScheduleResponse struct {
	ID         *uuid.UUID `json:"id,omitempty" swaggertype:"string" format:"uuid"`
	RoomID     uuid.UUID  `json:"roomId" swaggertype:"string" format:"uuid"`
	DaysOfWeek []int      `json:"daysOfWeek"`
	StartTime  string     `json:"startTime"`
	EndTime    string     `json:"endTime"`
} //@name Schedule

// CreateScheduleResponse is the response body for CreateSchedule.
type CreateScheduleResponse struct {
	Schedule *ScheduleResponse `json:"schedule"`
} //@name CreateScheduleResponse
