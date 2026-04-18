//go:build integration

package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_Register_Success(t *testing.T) {
	skipIfNoServer(t)

	email := fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	body := map[string]any{
		"email":    email,
		"password": "password123",
		"role":     "user",
	}

	resp := doRequest(t, "POST", "/register", "", body)
	assert.Equal(t, 201, resp.StatusCode)

	result := readJSON(t, resp)
	user, ok := result["user"].(map[string]any)
	require.True(t, ok, "user not in response")
	assert.Equal(t, email, user["email"])
	assert.Equal(t, "user", user["role"])
	assert.NotEmpty(t, user["id"])
}

func TestIntegration_Register_MissingFields(t *testing.T) {
	skipIfNoServer(t)

	resp := doRequest(t, "POST", "/register", "", map[string]any{
		"email": "missing_password@example.com",
		"role":  "user",
	})
	assert.Equal(t, 400, resp.StatusCode)
}

func TestIntegration_Register_InvalidRole(t *testing.T) {
	skipIfNoServer(t)

	resp := doRequest(t, "POST", "/register", "", map[string]any{
		"email":    "test@example.com",
		"password": "password123",
		"role":     "superuser",
	})
	assert.Equal(t, 400, resp.StatusCode)
}

func TestIntegration_Login_Success(t *testing.T) {
	skipIfNoServer(t)

	email := fmt.Sprintf("login_test_%d@example.com", time.Now().UnixNano())
	password := "securepassword"

	regResp := doRequest(t, "POST", "/register", "", map[string]any{
		"email":    email,
		"password": password,
		"role":     "user",
	})
	require.Equal(t, 201, regResp.StatusCode)
	readJSON(t, regResp)

	resp := doRequest(t, "POST", "/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	assert.Equal(t, 200, resp.StatusCode)

	result := readJSON(t, resp)
	token, ok := result["token"].(string)
	require.True(t, ok, "token not in response")
	assert.NotEmpty(t, token)
}

func TestIntegration_Login_WrongPassword(t *testing.T) {
	skipIfNoServer(t)

	email := fmt.Sprintf("wrong_pass_%d@example.com", time.Now().UnixNano())

	regResp := doRequest(t, "POST", "/register", "", map[string]any{
		"email":    email,
		"password": "correctpassword",
		"role":     "user",
	})
	require.Equal(t, 201, regResp.StatusCode)
	readJSON(t, regResp)

	resp := doRequest(t, "POST", "/login", "", map[string]any{
		"email":    email,
		"password": "wrongpassword",
	})
	assert.Equal(t, 401, resp.StatusCode)
}

func TestIntegration_Login_UnknownEmail(t *testing.T) {
	skipIfNoServer(t)

	resp := doRequest(t, "POST", "/login", "", map[string]any{
		"email":    "nobody_" + fmt.Sprintf("%d", time.Now().UnixNano()) + "@example.com",
		"password": "password",
	})
	assert.Equal(t, 401, resp.StatusCode)
}

func TestIntegration_LoginToken_WorksForAuth(t *testing.T) {
	skipIfNoServer(t)

	email := fmt.Sprintf("auth_test_%d@example.com", time.Now().UnixNano())
	password := "mypassword"

	regResp := doRequest(t, "POST", "/register", "", map[string]any{
		"email":    email,
		"password": password,
		"role":     "user",
	})
	require.Equal(t, 201, regResp.StatusCode)
	readJSON(t, regResp)

	loginResp := doRequest(t, "POST", "/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	require.Equal(t, 200, loginResp.StatusCode)
	loginResult := readJSON(t, loginResp)
	token := loginResult["token"].(string)

	resp := doRequest(t, "GET", "/rooms/list", token, nil)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestIntegration_SlotsEndpoint_NoSchedule(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getToken(t, "admin")
	userToken := getToken(t, "user")

	roomBody := map[string]any{
		"name": fmt.Sprintf("No Schedule Room %d", time.Now().UnixNano()),
	}
	resp := doRequest(t, "POST", "/rooms/create", adminToken, roomBody)
	require.Equal(t, 201, resp.StatusCode)
	roomResult := readJSON(t, resp)
	roomData := roomResult["room"].(map[string]any)
	roomID := roomData["id"].(string)

	date := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	resp = doRequest(t, "GET", "/rooms/"+roomID+"/slots/list?date="+date, userToken, nil)
	assert.Equal(t, 200, resp.StatusCode)
	result := readJSON(t, resp)
	slots := result["slots"].([]any)
	assert.Empty(t, slots)
}

func TestIntegration_Pagination(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getToken(t, "admin")

	resp := doRequest(t, "GET", "/bookings/list?page=1&pageSize=5", adminToken, nil)
	assert.Equal(t, 200, resp.StatusCode)
	result := readJSON(t, resp)
	pagination := result["pagination"].(map[string]any)
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(5), pagination["pageSize"])
}

func TestIntegration_Pagination_InvalidPage(t *testing.T) {
	skipIfNoServer(t)

	adminToken := getToken(t, "admin")

	resp := doRequest(t, "GET", "/bookings/list?page=0", adminToken, nil)
	assert.Equal(t, 400, resp.StatusCode)
}
