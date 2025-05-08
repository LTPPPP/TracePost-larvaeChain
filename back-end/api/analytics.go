package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
)

// AnalyzeTransactionHandler handles transaction analysis requests
func AnalyzeTransactionHandler(c *fiber.Ctx) error {
	type request struct {
		TxID string `json:"txID"`
	}
	type response struct {
		Insights map[string]interface{} `json:"insights"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	analyticsService := blockchain.AnalyticsService{}
	insights, err := analyticsService.AnalyzeTransaction(req.TxID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(response{Insights: insights})
}

// PredictRiskHandler handles risk prediction requests
func PredictRiskHandler(c *fiber.Ctx) error {
	type request struct {
		TxID string `json:"txID"`
	}
	type response struct {
		RiskLevel string `json:"riskLevel"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	analyticsService := blockchain.AnalyticsService{}
	riskLevel, err := analyticsService.PredictRisk(req.TxID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(response{RiskLevel: riskLevel})
}