package roomhandler

import (
	"time"

	"github.com/google/uuid"
)

// RoomResponse is the API representation of a room.
type RoomResponse struct {
	ID          uuid.UUID  `json:"id" swaggertype:"string" format:"uuid"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	Capacity    *int       `json:"capacity,omitempty"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" swaggertype:"string" format:"date-time"`
} //@name Room

// ListRoomsResponse is the response body for ListRooms.
type ListRoomsResponse struct {
	Rooms []RoomResponse `json:"rooms"`
} //@name ListRoomsResponse

// CreateRoomResponse is the response body for CreateRoom.
type CreateRoomResponse struct {
	Room *RoomResponse `json:"room"`
} //@name CreateRoomResponse
