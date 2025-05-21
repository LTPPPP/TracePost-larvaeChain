package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/LTPPPP/TracePost-larvaeChain/blockchain"
	"github.com/LTPPPP/TracePost-larvaeChain/config"
	"strings"
	"time"
)

func DDIAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		didHeader := c.Get("X-DID")
		if didHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID header is required")
		}

		didProofHeader := c.Get("X-DID-Proof")
		if didProofHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID proof is required")
		}
		
		timestampHeader := c.Get("X-DID-Timestamp")
		if timestampHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID timestamp is required")
		}

		cfg := config.GetConfig()
		blockchainClient := blockchain.NewBlockchainClient(
			cfg.BlockchainNodeURL,
			"", // Private key is not needed for verification
			cfg.BlockchainAccount,
			cfg.BlockchainChainID,
			cfg.BlockchainConsensus,
		)

		identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)

		didClient := blockchain.NewW3CDIDClient(identityClient)
		
		didDoc, err := didClient.SupportedMethods["tracepost"].Resolve(didHeader)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Failed to resolve DID: " + err.Error())
		}
		
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
		
		message := didHeader + ":" + timestampHeader
		isValid, err := identityClient.VerifySignature(message, didProofHeader, verificationMethod)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify DID proof: "+err.Error())
		}

		if !isValid {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid DID proof")
		}

		timestamp, err := time.Parse(time.RFC3339, timestampHeader)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid timestamp format")
		}
		
		now := time.Now().UTC()
		maxPast := now.Add(-15 * time.Minute)
		maxFuture := now.Add(15 * time.Minute)
		
		if timestamp.Before(maxPast) || timestamp.After(maxFuture) {
			return fiber.NewError(fiber.StatusUnauthorized, "Timestamp out of acceptable range")
		}

		c.Locals("did", didHeader)

		return c.Next()
	}
}

func DDIPermissionMiddleware(requiredPermissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		did, ok := c.Locals("did").(string)
		if !ok || did == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "DID authentication required")
		}

		cfg := config.GetConfig()
		blockchainClient := blockchain.NewBlockchainClient(
			cfg.BlockchainNodeURL,
			"", // Private key is not needed for verification
			cfg.BlockchainAccount,
			cfg.BlockchainChainID,
			cfg.BlockchainConsensus,
		)

		identityClient := blockchain.NewIdentityClient(blockchainClient, cfg.IdentityRegistryContract)

		for _, permission := range requiredPermissions {
			hasPermission, err := identityClient.VerifyPermission(did, permission)
			if err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to verify permission: "+err.Error())
			}

			if !hasPermission {
				readablePermissions := strings.Join(requiredPermissions, "', '")
				readablePermissions = "'" + readablePermissions + "'"
				
				return fiber.NewError(
					fiber.StatusForbidden,
					"DID '"+did+"' does not have sufficient permissions. Required permission(s): "+readablePermissions,
				)
			}
		}

		return c.Next()
	}
}

func DDIProtect(requiredPermissions ...string) []fiber.Handler {
	return []fiber.Handler{
		DDIAuthMiddleware(),
		DDIPermissionMiddleware(requiredPermissions...),
	}
}
