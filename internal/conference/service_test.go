package conference_test

import (
	"context"
	"testing"

	"github.com/ivanov-matvey/room-booking/internal/conference"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStub_CreateConferenceLink(t *testing.T) {
	svc := &conference.Stub{}
	bookingID := uuid.New()

	link, err := svc.CreateConferenceLink(context.Background(), bookingID)
	require.NoError(t, err)
	assert.NotEmpty(t, link)
	assert.Contains(t, link, bookingID.String())
}

func TestStub_Idempotent(t *testing.T) {
	svc := &conference.Stub{}
	bookingID := uuid.New()

	link1, err := svc.CreateConferenceLink(context.Background(), bookingID)
	require.NoError(t, err)

	link2, err := svc.CreateConferenceLink(context.Background(), bookingID)
	require.NoError(t, err)

	assert.Equal(t, link1, link2)
}

func TestStub_ImplementsInterface(t *testing.T) {
	var svc conference.Service = &conference.Stub{}
	assert.NotNil(t, svc)
}
