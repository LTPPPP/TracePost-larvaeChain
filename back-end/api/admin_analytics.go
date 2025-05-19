package api

import (
	// "encoding/json"
	"time"
	
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/analytics"
)

// GetAdminDashboardAnalytics retrieves combined analytics for admin dashboard
// @Summary Get admin dashboard analytics
// @Description Get comprehensive analytics for the admin dashboard
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/dashboard [get]
func GetAdminDashboardAnalytics(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Get all metrics
	metrics := analyticsService.GetAllMetrics()

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Analytics data retrieved successfully",
		Data:    metrics,
	})
}

// GetSystemMetrics retrieves system performance metrics
// @Summary Get system performance metrics
// @Description Get metrics about system performance
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/system [get]
func GetSystemMetrics(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Get system metrics
	metrics := analyticsService.GetSystemMetrics()

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "System metrics retrieved successfully",
		Data:    metrics,
	})
}

// GetBlockchainAnalytics retrieves blockchain performance metrics
// @Summary Get blockchain analytics
// @Description Get analytics about blockchain performance
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/blockchain [get]
func GetBlockchainAnalytics(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Get blockchain metrics
	metrics := analyticsService.GetBlockchainMetrics()

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Blockchain analytics retrieved successfully",
		Data:    metrics,
	})
}

// GetComplianceAnalytics retrieves compliance metrics
// @Summary Get compliance analytics
// @Description Get analytics about system compliance
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/compliance [get]
func GetComplianceAnalytics(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Get compliance metrics
	metrics := analyticsService.GetComplianceMetrics()

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Compliance analytics retrieved successfully",
		Data:    metrics,
	})
}

// GetUserActivityAnalytics retrieves user activity metrics
// @Summary Get user activity analytics
// @Description Get analytics about user activity
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/users [get]
func GetUserActivityAnalytics(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Get user activity metrics
	metrics := analyticsService.GetUserActivityMetrics()

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "User activity analytics retrieved successfully",
		Data:    metrics,
	})
}

// GetBatchAnalytics retrieves batch production and tracking metrics
// @Summary Get batch analytics
// @Description Get analytics about batches and production
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/batches [get]
func GetBatchAnalytics(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Get batch metrics
	metrics := analyticsService.GetBatchMetrics()

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Batch analytics retrieved successfully",
		Data:    metrics,
	})
}

// ExportAnalyticsData exports analytics data as JSON
// @Summary Export analytics data
// @Description Export all analytics data in JSON format
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/export [get]
func ExportAnalyticsData(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Get all metrics as JSON
	jsonData, err := analyticsService.GetMetricsJSON()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Error exporting analytics data: "+err.Error())
	}

	// Set filename with current date
	filename := "tracepost_analytics_" + time.Now().Format("2006-01-02") + ".json"
	
	// Set content disposition header for download
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/json")
	
	return c.SendString(jsonData)
}

// RefreshAnalyticsData forces a refresh of analytics data
// @Summary Refresh analytics data
// @Description Force a refresh of all analytics data
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /admin/analytics/refresh [post]
func RefreshAnalyticsData(c *fiber.Ctx) error {
	// Check admin role
	role := c.Locals("role").(string)
	if role != "admin" {
		return fiber.NewError(fiber.StatusForbidden, "Only admin users can perform this action")
	}

	// Get analytics instance
	analyticsService := analytics.GetAnalytics()
	if analyticsService == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Analytics service not initialized")
	}

	// Trigger data collection
	go analyticsService.CollectAllMetrics()

	return c.JSON(SuccessResponse{
		Success: true,
		Message: "Analytics data refresh triggered",
		Data: map[string]interface{}{
			"triggered_at": time.Now(),
			"status": "processing",
		},
	})
}
