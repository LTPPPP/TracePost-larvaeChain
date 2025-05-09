package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"strings"
)

// DDIAuthMiddleware verifies decentralized digital identity authentication
// It checks if the request includes a valid DID and proof of identity
func DDIAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get DID from header
		didHeader := c.Get("X-DID")
		if didHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID header is required")
		}

		// Get DID proof from header
		didProofHeader := c.Get("X-DID-Proof")
		if didProofHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID proof is required")
		}

		// Initialize blockchain client
		cfg := config.GetConfig()
		blockchainClient := blockchain.NewBlockchainClient(
			cfg.BlockchainNodeURL,
			"", // Private key is not needed for verification
			cfg.BlockchainAccount,
			cfg.BlockchainChainID,
			cfg.BlockchainConsensus,
		)

		// Create identity client
		identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)

		// Verify the DID and proof
		isValid, err := identityClient.VerifyDIDProof(didHeader, didProofHeader)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify DID: "+err.Error())
		}

		if !isValid {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid DID or proof")
		}

		// Set the verified DID in context for later use
		c.Locals("did", didHeader)

		// Continue to the next middleware or route handler
		return c.Next()
	}
}

// DDIPermissionMiddleware checks if the authenticated DID has the required permissions
func DDIPermissionMiddleware(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the verified DID from context
		did, ok := c.Locals("did").(string)
		if !ok || did == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID authentication required")
		}

		// Initialize blockchain client
		cfg := config.GetConfig()
		blockchainClient := blockchain.NewBlockchainClient(
			cfg.BlockchainNodeURL,
			"", // Private key is not needed for verification
			cfg.BlockchainAccount,
			cfg.BlockchainChainID,
			cfg.BlockchainConsensus,
		)

		// Create identity client
		identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)

		// Check permissions for each required permission
		for _, permission := range requiredPermissions {
			hasPermission, err := identityClient.VerifyPermission(did, permission)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify permission: "+err.Error())
			}

			if !hasPermission {
				// Format the required permissions for error message
				readablePermissions := strings.Join(requiredPermissions, "', '")
				readablePermissions = "'" + readablePermissions + "'"
				
				return fiber.NewError(
					fiber.StatusForbidden,
					"DID '"+did+"' does not have sufficient permissions. Required permission(s): "+readablePermissions,
				)
			}
		}

		// All permissions verified, continue to the next middleware or route handler
		return c.Next()
	}
}

// Combined DDI auth and permission middleware for convenience
func DDIProtect(requiredPermissions ...string) []fiber.Handler {
	return []fiber.Handler{
		DDIAuthMiddleware(),
		DDIPermissionMiddleware(requiredPermissions...),
	}
}
