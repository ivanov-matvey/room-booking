package infohandler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	infohandler "github.com/ivanov-matvey/room-booking/internal/http/handlers/info"

	"github.com/stretchr/testify/assert"
)

func TestGetInfo(t *testing.T) {
	h := infohandler.New()
	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	w := httptest.NewRecorder()
	h.GetInfo(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
