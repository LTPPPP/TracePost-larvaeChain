package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/db"
	"github.com/LTPPPP/TracePost-larvaeChain/models"
	"github.com/stretchr/testify/assert"
)

// mockUser creates a user with admin role for testing
func setupAdminUser() *models.User {
	return &models.User{
		ID:        1,
		Username:  "admin_user",
		FullName:  "Admin User",
		Email:     "admin@example.com",
		Role:      "admin",
		CompanyID: 1,
		IsActive:  true,
	}
}

// TestLockUnlockUser tests the lock/unlock user functionality
func TestLockUnlockUser(t *testing.T) {
	// Setup fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	// Mock request
	reqBody := LockUserRequest{
		IsActive: false,
		Reason:   "Account suspended due to suspicious activity",
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create a request
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/2/status", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Mock admin authentication
	admin := setupAdminUser()

	// Setup context with admin role
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("role", admin.Role)
	c.Params("userId", "2")
	
	// Execute the handler
	resp := LockUnlockUser(c)
	
	// Check response
	assert.Equal(t, fiber.StatusOK, resp.Status())
	
	var response SuccessResponse
	err := json.Unmarshal(resp.Body(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "locked")
	
	// Test the reverse (unlock)
	reqBody.IsActive = true
	reqBody.Reason = "Account restored after verification"
	reqJSON, _ = json.Marshal(reqBody)
	
	req = httptest.NewRequest(http.MethodPut, "/api/v1/admin/users/2/status", bytes.NewReader(reqJSON))
	c = app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("role", admin.Role)
	c.Params("userId", "2")
	
	resp = LockUnlockUser(c)
	assert.Equal(t, fiber.StatusOK, resp.Status())
	
	json.Unmarshal(resp.Body(), &response)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "unlocked")
}

// TestGetUsersByRole tests retrieving users by role
func TestGetUsersByRole(t *testing.T) {
	// Setup fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})
	
	// Create a request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?role=hatchery_manager", nil)
	
	// Mock admin authentication
	admin := setupAdminUser()
	
	// Setup context with admin role
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("role", admin.Role)
	c.Request().URI().QueryArgs().Set("role", "hatchery_manager")
	
	// Execute the handler
	resp := GetUsersByRole(c)
	
	// Check response
	assert.Equal(t, fiber.StatusOK, resp.Status())
	
	var response SuccessResponse
	err := json.Unmarshal(resp.Body(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "retrieved")
}

// TestApproveHatchery tests approving a hatchery account
func TestApproveHatchery(t *testing.T) {
	// Setup fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	// Mock request
	reqBody := ApproveHatcheryRequest{
		IsApproved: true,
		Comment:    "All verification requirements met",
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create a request
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/hatcheries/1/approve", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Mock admin authentication
	admin := setupAdminUser()

	// Setup context with admin role
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("role", admin.Role)
	c.Params("hatcheryId", "1")
	
	// Execute the handler
	resp := ApproveHatchery(c)
	
	// Check response
	assert.Equal(t, fiber.StatusOK, resp.Status())
	
	var response SuccessResponse
	err := json.Unmarshal(resp.Body(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "approved")
}

// TestRevokeCertificate tests revoking a compliance certificate
func TestRevokeCertificate(t *testing.T) {
	// Setup fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	// Mock request
	reqBody := RevokeCertificateRequest{
		Reason: "Environmental standards violation detected",
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create a request
	req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/certificates/1/revoke", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Mock admin authentication
	admin := setupAdminUser()

	// Setup context with admin role
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("role", admin.Role)
	c.Params("docId", "1")
	
	// Execute the handler
	resp := RevokeCertificate(c)
	
	// Check response
	assert.Equal(t, fiber.StatusOK, resp.Status())
	
	var response SuccessResponse
	err := json.Unmarshal(resp.Body(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "revoked")
}

// TestCheckStandardCompliance tests compliance checking functionality
func TestCheckStandardCompliance(t *testing.T) {
	// Setup fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	// Mock request
	reqBody := StandardCheckRequest{
		BatchID:   1,
		Standards: []string{"FDA", "ASC"},
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create a request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/compliance/check", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Mock admin authentication
	admin := setupAdminUser()

	// Setup context with admin role
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("role", admin.Role)
	
	// Execute the handler
	resp := CheckStandardCompliance(c)
	
	// Check response
	assert.Equal(t, fiber.StatusOK, resp.Status())
	
	var response SuccessResponse
	err := json.Unmarshal(resp.Body(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "Compliance")
}

// TestIssueDID tests issuing a decentralized identifier
func TestIssueDID(t *testing.T) {
	// Setup fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler,
	})

	// Mock request
	reqBody := DIDRequest{
		EntityType: "hatchery",
		EntityID:   1,
		Claims: map[string]interface{}{
			"certification": "organic",
		},
	}
	reqJSON, _ := json.Marshal(reqBody)

	// Create a request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/identity/issue", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	
	// Mock admin authentication
	admin := setupAdminUser()

	// Setup context with admin role
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Locals("role", admin.Role)
	
	// Execute the handler
	resp := IssueDID(c)
	
	// Check response
	assert.Equal(t, fiber.StatusOK, resp.Status())
	
	var response SuccessResponse
	err := json.Unmarshal(resp.Body(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "issued")
}
