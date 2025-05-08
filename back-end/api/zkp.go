package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
)

// GenerateProofHandler handles the generation of Zero-Knowledge Proofs
func GenerateProofHandler(c *fiber.Ctx) error {
	type request struct {
		Data string `json:"data"`
	}
	type response struct {
		Proof string `json:"proof"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	zkpService := blockchain.ZKPService{}
	proof, err := zkpService.GenerateProof(req.Data)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(response{Proof: proof})
}

// VerifyProofHandler handles the verification of Zero-Knowledge Proofs
func VerifyProofHandler(c *fiber.Ctx) error {
	type request struct {
		Data  string `json:"data"`
		Proof string `json:"proof"`
	}
	type response struct {
		Valid bool `json:"valid"`
	}

	var req request
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request format")
	}

	zkpService := blockchain.ZKPService{}
	valid, err := zkpService.VerifyProof(req.Data, req.Proof)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(response{Valid: valid})
}