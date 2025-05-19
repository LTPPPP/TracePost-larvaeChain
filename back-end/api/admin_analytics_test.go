package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LTPPPP/TracePost-larvaeChain/analytics"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func initAnalyticsForTest() {
	analytics.InitAnalytics()
}

func setupAdminAnalyticsTestApp() *fiber.App {
	app := fiber.New()
	
	// Initialize analytics for testing
	initAnalyticsForTest()
	
	// Setup routes for testing
	app.Get("/admin/analytics/dashboard", func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // Mock admin role
		return GetAdminDashboardAnalytics(c)
	})
	
	app.Get("/admin/analytics/system", func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // Mock admin role
		return GetSystemMetrics(c)
	})
	
	app.Get("/admin/analytics/blockchain", func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // Mock admin role
		return GetBlockchainAnalytics(c)
	})
	
	app.Get("/admin/analytics/compliance", func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // Mock admin role
		return GetComplianceAnalytics(c)
	})
	
	app.Get("/admin/analytics/users", func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // Mock admin role
		return GetUserActivityAnalytics(c)
	})
	
	app.Get("/admin/analytics/batches", func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // Mock admin role
		return GetBatchAnalytics(c)
	})
	
	app.Post("/admin/analytics/refresh", func(c *fiber.Ctx) error {
		c.Locals("role", "admin") // Mock admin role
		return RefreshAnalyticsData(c)
	})
	
	return app
}

func TestGetAdminDashboardAnalytics(t *testing.T) {
	// Setup app with test routes
	app := setupAdminAnalyticsTestApp()
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/dashboard", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert structure
	assert.Equal(t, true, result["success"])
	assert.NotNil(t, result["data"])
	
	// Check data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	
	// Check main categories
	assert.NotNil(t, data["system"])
	assert.NotNil(t, data["compliance"])
	assert.NotNil(t, data["blockchain"])
	assert.NotNil(t, data["user_activity"])
	assert.NotNil(t, data["batch"])
	assert.NotNil(t, data["timestamp"])
}

func TestGetSystemMetrics(t *testing.T) {
	// Setup app with test routes
	app := setupAdminAnalyticsTestApp()
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/system", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert structure
	assert.Equal(t, true, result["success"])
	assert.NotNil(t, result["data"])
	
	// Check data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	
	// Check system metrics fields
	assert.NotNil(t, data["active_users"])
	assert.NotNil(t, data["total_batches"])
	assert.NotNil(t, data["blockchain_tx_count"])
	assert.NotNil(t, data["api_requests_per_hour"])
	assert.NotNil(t, data["avg_response_time_ms"])
	assert.NotNil(t, data["system_health"])
	assert.NotNil(t, data["server_cpu_usage"])
	assert.NotNil(t, data["server_memory_usage"])
	assert.NotNil(t, data["db_connections"])
	assert.NotNil(t, data["last_updated"])
}

func TestGetBlockchainAnalytics(t *testing.T) {
	// Setup app with test routes
	app := setupAdminAnalyticsTestApp()
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/blockchain", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert structure
	assert.Equal(t, true, result["success"])
	assert.NotNil(t, result["data"])
	
	// Check data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	
	// Check blockchain metrics fields
	assert.NotNil(t, data["total_nodes"])
	assert.NotNil(t, data["active_nodes"])
	assert.NotNil(t, data["network_health"])
	assert.NotNil(t, data["consensus_status"])
	assert.NotNil(t, data["average_block_time_ms"])
	assert.NotNil(t, data["transactions_per_second"])
	assert.NotNil(t, data["pending_transactions"])
	assert.NotNil(t, data["node_latencies"])
	assert.NotNil(t, data["chain_health"])
	assert.NotNil(t, data["cross_chain_transactions"])
	assert.NotNil(t, data["last_updated"])
}

func TestGetComplianceAnalytics(t *testing.T) {
	// Setup app with test routes
	app := setupAdminAnalyticsTestApp()
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/compliance", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert structure
	assert.Equal(t, true, result["success"])
	assert.NotNil(t, result["data"])
	
	// Check data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	
	// Check compliance metrics fields
	assert.NotNil(t, data["total_certificates"])
	assert.NotNil(t, data["valid_certificates"]) 
	assert.NotNil(t, data["expired_certificates"])
	assert.NotNil(t, data["revoked_certificates"])
	assert.NotNil(t, data["company_compliance"])
	assert.NotNil(t, data["standards_compliance"])
	assert.NotNil(t, data["regional_compliance"])
	assert.NotNil(t, data["last_updated"])
}

func TestGetUserActivityAnalytics(t *testing.T) {
	// Setup app with test routes
	app := setupAdminAnalyticsTestApp()
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/users", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert structure
	assert.Equal(t, true, result["success"])
	assert.NotNil(t, result["data"])
	
	// Check data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	
	// Check user activity metrics fields
	assert.NotNil(t, data["active_users_by_role"])
	assert.NotNil(t, data["login_frequency"])
	assert.NotNil(t, data["api_endpoint_usage"])
	assert.NotNil(t, data["most_active_users"])
	assert.NotNil(t, data["user_growth"])
	assert.NotNil(t, data["last_updated"])
}

func TestGetBatchAnalytics(t *testing.T) {
	// Setup app with test routes
	app := setupAdminAnalyticsTestApp()
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/batches", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert structure
	assert.Equal(t, true, result["success"])
	assert.NotNil(t, result["data"])
	
	// Check data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	
	// Check batch metrics fields
	assert.NotNil(t, data["total_batches_produced"])
	assert.NotNil(t, data["active_batches"])
	assert.NotNil(t, data["batches_by_status"])
	assert.NotNil(t, data["batches_by_region"])
	assert.NotNil(t, data["batches_by_species"])
	assert.NotNil(t, data["batches_by_hatchery"])
	assert.NotNil(t, data["last_updated"])
}

func TestRefreshAnalyticsData(t *testing.T) {
	// Setup app with test routes
	app := setupAdminAnalyticsTestApp()
	
	// Create request
	req := httptest.NewRequest(http.MethodPost, "/admin/analytics/refresh", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert structure
	assert.Equal(t, true, result["success"])
	assert.NotNil(t, result["data"])
	
	// Check data structure
	data, ok := result["data"].(map[string]interface{})
	assert.True(t, ok)
	
	// Check response fields
	assert.NotNil(t, data["triggered_at"])
	assert.Equal(t, "processing", data["status"])
}

func TestAccessDeniedForNonAdmin(t *testing.T) {
	// Setup app with test routes but without admin role
	app := fiber.New()
	
	// Initialize analytics for testing
	initAnalyticsForTest()
	
	// Setup non-admin route for testing
	app.Get("/admin/analytics/dashboard", func(c *fiber.Ctx) error {
		c.Locals("role", "viewer") // Non-admin role
		return GetAdminDashboardAnalytics(c)
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/admin/analytics/dashboard", nil)
	resp, err := app.Test(req)
	
	// Assert no error in making request
	assert.Nil(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	
	// Read response body
	body, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	assert.Nil(t, err)
	
	// Assert error structure
	assert.Equal(t, false, result["success"])
	assert.Contains(t, result["message"], "error")
}
