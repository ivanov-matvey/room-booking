package conference

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type Service interface {
	CreateConferenceLink(ctx context.Context, bookingID uuid.UUID) (string, error)
}

type Stub struct {
	mu    sync.Mutex
	links map[uuid.UUID]string
}

func (s *Stub) CreateConferenceLink(_ context.Context, bookingID uuid.UUID) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.links == nil {
		s.links = make(map[uuid.UUID]string)
	}

	if link, ok := s.links[bookingID]; ok {
		return link, nil
	}

	link := fmt.Sprintf("https://meet.example.com/%s", bookingID.String())
	s.links[bookingID] = link
	return link, nil
}
