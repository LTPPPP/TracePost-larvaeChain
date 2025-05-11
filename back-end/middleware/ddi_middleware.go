package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"strings"
	"time"
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
		
		// Get timestamp from header to prevent replay attacks
		timestampHeader := c.Get("X-DID-Timestamp")
		if timestampHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID timestamp is required")
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

		// Create W3C DID client
		didClient := blockchain.NewW3CDIDClient(identityClient)
		
		// Resolve DID document for the provided DID
		didDoc, err := didClient.SupportedMethods["tracepost"].Resolve(didHeader)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Failed to resolve DID: " + err.Error())
		}
		
		// Verify the DID proof
		// The proof should be a signature of "<DID>:<timestamp>" using the private key
		// that corresponds to the public key in the DID document
		
		// Find the verification method
		var verificationMethod *blockchain.W3CVerificationMethod
		for _, vm := range didDoc.VerificationMethod {
			if strings.HasSuffix(vm.ID, "#keys-1") {
				verificationMethod = &vm
				break
			}
		}
		
		if verificationMethod == nil {
			return fiber.NewError(fiber.StatusUnauthorized, "No valid verification method found in DID document")
		}
		
		// Verify the proof
		message := didHeader + ":" + timestampHeader
		isValid, err := identityClient.VerifySignature(message, didProofHeader, verificationMethod)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify DID proof: "+err.Error())
		}

		if !isValid {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid DID proof")
		}

		// Check timestamp to prevent replay attacks
		timestamp, err := time.Parse(time.RFC3339, timestampHeader)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid timestamp format")
		}
		
		// Ensure timestamp is not too old or in the future
		// Allow for a 15-minute window
		now := time.Now().UTC()
		maxPast := now.Add(-15 * time.Minute)
		maxFuture := now.Add(15 * time.Minute)
		
		if timestamp.Before(maxPast) || timestamp.After(maxFuture) {
			return fiber.NewError(fiber.StatusUnauthorized, "Timestamp out of acceptable range")
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
