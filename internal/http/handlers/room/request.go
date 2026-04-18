package roomhandler

// CreateRoomRequest is the request body for CreateRoom.
type CreateRoomRequest struct {
	Name        string  `json:"name" example:"Переговорная №1"`
	Description *string `json:"description,omitempty" example:"Большая переговорная на 10 мест"`
	Capacity    *int    `json:"capacity,omitempty" example:"10"`
} //@name CreateRoomRequest
