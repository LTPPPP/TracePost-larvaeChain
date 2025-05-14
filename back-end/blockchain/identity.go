// identity.go
package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"
)

// IdentityClient provides decentralized digital identity capabilities
type IdentityClient struct {
	// Base blockchain client
	BaseClient *BlockchainClient
	
	// Identity registry on the blockchain
	RegistryContract string
	
	// Local identity cache
	IdentityCache map[string]*DecentralizedID
	
	// Advanced identity capabilities for 2025
	SSIClient    *SSIClient
	W3CDIDClient *W3CDIDClient
}

// DecentralizedID represents a decentralized digital identity
type DecentralizedID struct {
	DID           string                 // Decentralized Identifier (did:tracepost:...)
	ControllerDID string                 // Controller DID (optional, for delegated identities)
	PublicKey     string                 // Public key
	MetaData      map[string]interface{} // Additional metadata
	Status        string                 // "active", "revoked", "suspended"
	Created       time.Time              // Creation timestamp
	Updated       time.Time              // Last update timestamp
	Proof         *IdentityProof         // Cryptographic proof
}

// IdentityProof represents a cryptographic proof for verifying identity claims
type IdentityProof struct {
	Type               string    // Proof type (e.g., "EcdsaSecp256k1Signature2025")
	Created            time.Time // When the proof was created
	VerificationMethod string    // Method used to verify the proof
	ProofPurpose       string    // Purpose of the proof
	ProofValue         string    // The actual proof value (signature)
}

// IdentityClaim represents a verifiable claim about an identity
type IdentityClaim struct {
	ID           string                 // Unique identifier for the claim
	Type         string                 // Claim type
	Issuer       string                 // DID of the claim issuer
	Subject      string                 // DID of the claim subject
	IssuanceDate time.Time              // When the claim was issued
	ExpiryDate   time.Time              // When the claim expires
	Claims       map[string]interface{} // The actual claims
	Proof        *IdentityProof         // Cryptographic proof
	Status       string                 // "valid", "revoked", "expired"
}

// VerificationResult represents the result of verifying a claim
type VerificationResult struct {
	IsValid        bool      // Whether the claim is valid
	ValidationTime time.Time // When the validation was performed
	Errors         []string  // Any validation errors
}

// NewIdentityClient creates a new identity client
func NewIdentityClient(baseClient *BlockchainClient, registryContract string) *IdentityClient {
	client := &IdentityClient{
		BaseClient:       baseClient,
		RegistryContract: registryContract,
		IdentityCache:    make(map[string]*DecentralizedID),
	}
	
	// Initialize SSI client
	client.SSIClient = NewSSIClient(client, "tracepost")
	
	// Initialize W3C DID client
	client.W3CDIDClient = NewW3CDIDClient(client)
	
	return client
}

// CreateDecentralizedID creates a new decentralized identity
func (ic *IdentityClient) CreateDecentralizedID(entityType, entityName string, metadata map[string]interface{}) (*DecentralizedID, error) {
	// Generate a new key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Generate DID based on public key
	pubKeyBytes := elliptic.Marshal(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)
	
	// Create DID in format did:tracepost:<entity-type>:<hex-encoded-public-key>
	did := fmt.Sprintf("did:tracepost:%s:%s", entityType, pubKeyHex[:16])
	
	// Create identity
	now := time.Now()
	identity := &DecentralizedID{
		DID:       did,
		PublicKey: pubKeyHex,
		MetaData: map[string]interface{}{
			"name": entityName,
			"type": entityType,
		},
		Status:  "active",
		Created: now,
		Updated: now,
	}
	
	// Add any additional metadata
	for k, v := range metadata {
		identity.MetaData[k] = v
	}
	
	// Create proof
	digest := sha256.Sum256([]byte(did + pubKeyHex))
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create proof: %v", err)
	}
	
	signature := append(r.Bytes(), s.Bytes()...)
	signatureB64 := base64.StdEncoding.EncodeToString(signature)
	
	identity.Proof = &IdentityProof{
		Type:               "EcdsaSecp256k1Signature2025",
		Created:            now,
		VerificationMethod: did + "#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         signatureB64,
	}
	
	// Register on blockchain
	_, err = ic.BaseClient.submitTransaction("REGISTER_DID", map[string]interface{}{
		"did":        did,
		"public_key": pubKeyHex,
		"metadata":   identity.MetaData,
		"status":     identity.Status,
		"created":    identity.Created,
		"proof":      identity.Proof,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register DID on blockchain: %v", err)
	}
	
	// Cache identity
	ic.IdentityCache[did] = identity
	
	return identity, nil
}

// ResolveDID resolves a DID to retrieve the associated DID document
func (ic *IdentityClient) ResolveDID(did string) (*DecentralizedID, error) {
	// Check cache first
	if identity, exists := ic.IdentityCache[did]; exists {
		return identity, nil
	}
	
	// In a real implementation, this would query the blockchain
	// For now, we'll just return an error since the DID isn't in our cache
	return nil, errors.New("DID not found")
}

// CreateVerifiableClaim creates a verifiable claim about an identity
func (ic *IdentityClient) CreateVerifiableClaim(
	issuerDID string,
	subjectDID string,
	claimType string,
	claims map[string]interface{},
	expiryDays int,
) (*IdentityClaim, error) {
	// Resolve issuer DID
	_, err := ic.ResolveDID(issuerDID)
	if err != nil {
		return nil, fmt.Errorf("issuer DID not found: %v", err)
	}
	
	// Resolve subject DID
	_, err = ic.ResolveDID(subjectDID)
	if err != nil {
		return nil, fmt.Errorf("subject DID not found: %v", err)
	}
	
	// Create claim
	now := time.Now()
	claim := &IdentityClaim{
		ID:           fmt.Sprintf("claim:%s:%d", issuerDID, now.Unix()),
		Type:         claimType,
		Issuer:       issuerDID,
		Subject:      subjectDID,
		IssuanceDate: now,
		ExpiryDate:   now.AddDate(0, 0, expiryDays),
		Claims:       claims,
		Status:       "valid",
	}
	
	// Register on blockchain
	_, err = ic.BaseClient.submitTransaction("REGISTER_CLAIM", map[string]interface{}{
		"claim_id":      claim.ID,
		"claim_type":    claim.Type,
		"issuer":        claim.Issuer,
		"subject":       claim.Subject,
		"issuance_date": claim.IssuanceDate,
		"expiry_date":   claim.ExpiryDate,
		"claims":        claim.Claims,
		"status":        claim.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register claim on blockchain: %v", err)
	}
	
	return claim, nil
}

// VerifyClaim verifies a claim
func (ic *IdentityClient) VerifyClaim(claim *IdentityClaim) (*VerificationResult, error) {
	result := &VerificationResult{
		IsValid:        false,
		ValidationTime: time.Now(),
		Errors:         []string{},
	}
	
	// Check if claim has expired
	if claim.ExpiryDate.Before(time.Now()) {
		result.Errors = append(result.Errors, "claim has expired")
		return result, nil
	}
	
	// Check if claim is revoked
	if claim.Status != "valid" {
		result.Errors = append(result.Errors, fmt.Sprintf("claim status is %s", claim.Status))
		return result, nil
	}
	
	// In a real implementation, this would verify the cryptographic proof
	// For now, we'll just assume the proof is valid if we've gotten this far
	result.IsValid = true
	
	return result, nil
}

// RevokeClaim revokes a claim
func (ic *IdentityClient) RevokeClaim(claimID string, issuerDID string) error {
	// In a real implementation, this would verify the issuer has permission to revoke
	// For now, we'll just update the status on the blockchain
	
	_, err := ic.BaseClient.submitTransaction("REVOKE_CLAIM", map[string]interface{}{
		"claim_id": claimID,
		"issuer":   issuerDID,
		"revoked_at": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to revoke claim on blockchain: %v", err)
	}
	
	return nil
}

// VerifySignature verifies a signature against a verification method
func (ic *IdentityClient) VerifySignature(message, signature string, verificationMethod *W3CVerificationMethod) (bool, error) {
	if verificationMethod == nil {
		return false, errors.New("verification method is nil")
	}
	
	// Extract public key from verification method
	if verificationMethod.Type == "JsonWebKey2020" && verificationMethod.PublicKeyJwk != nil {
		// Extract x and y coordinates from JWK
		xEncoded, xOk := verificationMethod.PublicKeyJwk["x"].(string)
		yEncoded, yOk := verificationMethod.PublicKeyJwk["y"].(string)
		
		if !xOk || !yOk {
			return false, errors.New("invalid public key format in verification method")
		}
		
		b64url := base64url{}
		xBytes, err := b64url.Decode(xEncoded)
		if err != nil {
			return false, fmt.Errorf("failed to decode x coordinate: %v", err)
		}
		
		yBytes, err := b64url.Decode(yEncoded)
		if err != nil {
			return false, fmt.Errorf("failed to decode y coordinate: %v", err)
		}
		
		x := new(big.Int).SetBytes(xBytes)
		y := new(big.Int).SetBytes(yBytes)
		
		publicKey := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     x,
			Y:     y,
		}
		
		// Decode the signature
		signatureBytes, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			return false, fmt.Errorf("failed to decode signature: %v", err)
		}
		
		// The signature should be in the format of r || s
		sigLen := len(signatureBytes)
		if sigLen%2 != 0 {
			return false, errors.New("invalid signature length")
		}
		
		rBytes := signatureBytes[:sigLen/2]
		sBytes := signatureBytes[sigLen/2:]
		
		var r, s big.Int
		r.SetBytes(rBytes)
		s.SetBytes(sBytes)
		
		// Create message hash
		messageHash := sha256.Sum256([]byte(message))
		
		// Verify the signature
		return ecdsa.Verify(publicKey, messageHash[:], &r, &s), nil
	} else if verificationMethod.Type == "Ed25519VerificationKey2020" && verificationMethod.PublicKeyMultibase != "" {
		// Ed25519 signature verification would be implemented here
		return false, errors.New("Ed25519 signature verification not implemented")
	}
	
	return false, fmt.Errorf("unsupported verification method type: %s", verificationMethod.Type)
}

// GetActorPermissions retrieves the permissions for an actor based on their DID
func (ic *IdentityClient) GetActorPermissions(actorDID string) (map[string]bool, error) {
	// In a real implementation, this would query the blockchain for all valid claims
	// about the actor and then map those to a set of permissions
	
	// For now, we'll just return a mock set of permissions
	permissions := map[string]bool{
		"create_batch":        true,
		"update_batch_status": true,
		"record_event":        true,
		"record_environment":  true,
		"upload_document":     true,
	}
	
	return permissions, nil
}

// VerifyPermission checks if an actor has a specific permission
func (ic *IdentityClient) VerifyPermission(actorDID string, permission string) (bool, error) {
	permissions, err := ic.GetActorPermissions(actorDID)
	if err != nil {
		return false, err
	}
	
	hasPermission, exists := permissions[permission]
	if !exists {
		return false, nil
	}
	
	return hasPermission, nil
}

// VerifyDIDProof verifies a DID proof
func (ic *IdentityClient) VerifyDIDProof(did, proofValue string) (bool, error) {
	// Resolve the DID to get the DID document
	didDoc, err := ic.ResolveDID(did)
	if err != nil {
		return false, fmt.Errorf("failed to resolve DID: %v", err)
	}
	
	// Check if DID is active
	if didDoc.Status != "active" {
		return false, fmt.Errorf("DID is not active (status: %s)", didDoc.Status)
	}
	
	// Decode the proof value (base64)
	signatureBytes, err := base64.StdEncoding.DecodeString(proofValue)
	if err != nil {
		return false, fmt.Errorf("failed to decode proof: %v", err)
	}
	
	// In a real implementation, this would verify the signature against the public key
	// from the DID document. For now, we'll implement a simplified version.
	
	// Extract public key from DID document
	pubKeyHex := didDoc.PublicKey
	
	// Convert hex-encoded public key to bytes
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode public key: %v", err)
	}
	
	// Parse public key
	x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
	if x == nil {
		return false, errors.New("failed to unmarshal public key")
	}
	
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	
	// Create challenge message (typically a nonce or timestamp)
	// In a real implementation, this would be part of the proof
	challenge := []byte(did + time.Now().Format("2006-01-02"))
	challengeHash := sha256.Sum256(challenge)
	
	// Verify signature
	// In ECDSA, the signature is typically two values: r and s
	// For simplicity, assuming the first half of signature is r and second half is s
	sigLen := len(signatureBytes)
	if sigLen%2 != 0 {
		return false, errors.New("invalid signature length")
	}
	
	rBytes := signatureBytes[:sigLen/2]
	sBytes := signatureBytes[sigLen/2:]
	
	var r, s big.Int
	r.SetBytes(rBytes)
	s.SetBytes(sBytes)
	
	// Verify the signature
	return ecdsa.Verify(publicKey, challengeHash[:], &r, &s), nil
}

// UpdateDIDPermissions updates the permissions for a DID
func (ic *IdentityClient) UpdateDIDPermissions(did string, permissions map[string]bool) error {
	// Resolve the DID to check if it exists
	_, err := ic.ResolveDID(did)
	if err != nil {
		return fmt.Errorf("failed to resolve DID: %v", err)
	}
	
	// Update permissions on blockchain
	_, err = ic.BaseClient.submitTransaction("UPDATE_DID_PERMISSIONS", map[string]interface{}{
		"did":         did,
		"permissions": permissions,
		"updated_at":  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to update DID permissions on blockchain: %v", err)
	}
	
	return nil
}

// VerifyPermissionBatch efficiently checks multiple permissions at once
func (ic *IdentityClient) VerifyPermissionBatch(actorDID string, permissions []string) (map[string]bool, error) {
	// Get all permissions for the actor
	allPermissions, err := ic.GetActorPermissions(actorDID)
	if err != nil {
		return nil, err
	}
	
	// Check each requested permission
	result := make(map[string]bool)
	for _, permission := range permissions {
		hasPermission, exists := allPermissions[permission]
		if !exists {
			result[permission] = false
		} else {
			result[permission] = hasPermission
		}
	}
	
	return result, nil
}

// Add support for Decentralized Digital Identities (DDIs)
func SetupDDI() error {
	fmt.Println("Setting up Decentralized Digital Identities...")
	// Add logic to initialize and manage DDIs
	return nil
}