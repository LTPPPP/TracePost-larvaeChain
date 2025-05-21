package api

import (
	"encoding/json"
	
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
)

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
	options := blockchain.ZKPOptions{
		Type: blockchain.ZKPTypeMerkle,
	}
	
	proof, err := zkpService.GenerateProof(req.Data, options)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	
	proofBytes, err := json.Marshal(proof)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to serialize proof")
	}

	return c.JSON(response{Proof: string(proofBytes)})
}

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
	
	var zkpProof blockchain.ZKPProof
	if err := json.Unmarshal([]byte(req.Proof), &zkpProof); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid proof format")
	}
	
	valid, err := zkpService.VerifyProof(req.Data, &zkpProof)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(response{Valid: valid})
}