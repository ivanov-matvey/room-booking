//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getServerURL() string {
	url := os.Getenv("TEST_SERVER_URL")
	if url == "" {
		return "http://localhost:8080"
	}
	return url
}

func skipIfNoServer(t *testing.T) {
	t.Helper()
	url := getServerURL() + "/_info"
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Skip("Test server not available, skipping integration tests")
	}
}

func getToken(t *testing.T, role string) string {
	t.Helper()
	body := fmt.Sprintf(`{"role": "%s"}`, role)
	resp, err := http.Post(getServerURL()+"/dummyLogin", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	token, ok := result["token"]
	require.True(t, ok, "token not in response")
	return token
}

func doRequest(t *testing.T, method, path, token string, body any) *http.Response {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, getServerURL()+path, reqBody)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}

func readJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}

func TestIntegration_HealthCheck(t *testing.T) {
	skipIfNoServer(t)

	resp := doRequest(t, "GET", "/_info", "", nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIntegration_DummyLogin(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getToken(t, "admin")
	assert.NotEmpty(t, adminToken)

	userToken := getToken(t, "user")
	assert.NotEmpty(t, userToken)
}

func TestIntegration_DummyLogin_InvalidRole(t *testing.T) {
	skipIfNoServer(t)

	body := `{"role": "superadmin"}`
	resp, err := http.Post(getServerURL()+"/dummyLogin", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestIntegration_CreateRoom_ListRooms_CreateSchedule_GetSlots_CreateBooking_CancelBooking(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getToken(t, "admin")
	userToken := getToken(t, "user")

	roomBody := map[string]any{
		"name":        fmt.Sprintf("Integration Test Room %d", time.Now().UnixNano()),
		"description": "Test room for integration tests",
		"capacity":    10,
	}
	resp := doRequest(t, "POST", "/rooms/create", adminToken, roomBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	roomResult := readJSON(t, resp)

	roomData, ok := roomResult["room"].(map[string]any)
	require.True(t, ok, "room not in response")
	roomID, ok := roomData["id"].(string)
	require.True(t, ok, "room id not in response")
	assert.NotEmpty(t, roomID)

	resp = doRequest(t, "GET", "/rooms/list", userToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	roomsResult := readJSON(t, resp)
	rooms, ok := roomsResult["rooms"].([]any)
	require.True(t, ok, "rooms not in response")
	assert.True(t, len(rooms) > 0)

	now := time.Now().UTC()
	daysUntilMonday := int(time.Monday - now.Weekday())
	if daysUntilMonday <= 0 {
		daysUntilMonday += 7
	}
	nextMonday := now.AddDate(0, 0, daysUntilMonday)

	scheduleBody := map[string]any{
		"daysOfWeek": []int{1, 2, 3, 4, 5},
		"startTime":  "09:00",
		"endTime":    "18:00",
	}
	resp = doRequest(t, "POST", "/rooms/"+roomID+"/schedule/create", adminToken, scheduleBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	scheduleResult := readJSON(t, resp)
	scheduleData, ok := scheduleResult["schedule"].(map[string]any)
	require.True(t, ok, "schedule not in response")
	assert.NotEmpty(t, scheduleData["id"])

	dateStr := nextMonday.Format("2006-01-02")
	resp = doRequest(t, "GET", "/rooms/"+roomID+"/slots/list?date="+dateStr, userToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	slotsResult := readJSON(t, resp)
	slots, ok := slotsResult["slots"].([]any)
	require.True(t, ok, "slots not in response")
	assert.True(t, len(slots) > 0, "should have slots on Monday")

	firstSlot, ok := slots[0].(map[string]any)
	require.True(t, ok)
	slotID, ok := firstSlot["id"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, slotID)

	bookingBody := map[string]any{
		"slotId":               slotID,
		"createConferenceLink": true,
	}
	resp = doRequest(t, "POST", "/bookings/create", userToken, bookingBody)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	bookingResult := readJSON(t, resp)
	bookingData, ok := bookingResult["booking"].(map[string]any)
	require.True(t, ok, "booking not in response")
	bookingID, ok := bookingData["id"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, bookingID)
	assert.Equal(t, "active", bookingData["status"])
	assert.NotNil(t, bookingData["conferenceLink"])

	resp = doRequest(t, "POST", "/bookings/create", userToken, bookingBody)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	resp = doRequest(t, "GET", "/bookings/my", userToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	myBookingsResult := readJSON(t, resp)
	myBookings, ok := myBookingsResult["bookings"].([]any)
	require.True(t, ok)
	assert.True(t, len(myBookings) > 0)

	resp = doRequest(t, "GET", "/bookings/list", adminToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	allBookingsResult := readJSON(t, resp)
	allBookings, ok := allBookingsResult["bookings"].([]any)
	require.True(t, ok)
	assert.True(t, len(allBookings) > 0)

	resp = doRequest(t, "POST", "/bookings/"+bookingID+"/cancel", userToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	cancelResult := readJSON(t, resp)
	cancelledBooking, ok := cancelResult["booking"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cancelled", cancelledBooking["status"])

	resp = doRequest(t, "POST", "/bookings/"+bookingID+"/cancel", userToken, nil)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestIntegration_AdminCannotBook(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getToken(t, "admin")

	bookingBody := map[string]any{
		"slotId": "00000000-0000-0000-0000-000000000099",
	}

	resp := doRequest(t, "POST", "/bookings/create", adminToken, bookingBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestIntegration_UserCannotCreateRoom(t *testing.T) {
	skipIfNoServer(t)

	userToken := getToken(t, "user")

	roomBody := map[string]any{
		"name": "Should Fail Room",
	}

	resp := doRequest(t, "POST", "/rooms/create", userToken, roomBody)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestIntegration_Unauthenticated(t *testing.T) {
	skipIfNoServer(t)

	resp := doRequest(t, "GET", "/rooms/list", "", nil)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestIntegration_ScheduleAlreadyExists(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getToken(t, "admin")

	roomBody := map[string]any{
		"name": fmt.Sprintf("Schedule Test Room %d", time.Now().UnixNano()),
	}
	resp := doRequest(t, "POST", "/rooms/create", adminToken, roomBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	roomResult := readJSON(t, resp)
	roomData := roomResult["room"].(map[string]any)
	roomID := roomData["id"].(string)

	scheduleBody := map[string]any{
		"daysOfWeek": []int{1},
		"startTime":  "09:00",
		"endTime":    "10:00",
	}
	resp = doRequest(t, "POST", "/rooms/"+roomID+"/schedule/create", adminToken, scheduleBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp = doRequest(t, "POST", "/rooms/"+roomID+"/schedule/create", adminToken, scheduleBody)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}
