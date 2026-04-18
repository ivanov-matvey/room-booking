package schedulehandler

// CreateScheduleRequest is the request body for CreateSchedule.
type CreateScheduleRequest struct {
	DaysOfWeek []int  `json:"daysOfWeek"`
	StartTime  string `json:"startTime" example:"09:00"`
	EndTime    string `json:"endTime" example:"18:00"`
} //@name CreateScheduleRequest
