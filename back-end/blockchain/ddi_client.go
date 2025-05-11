package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"time"
)

// DDIClientConfig represents configuration for a DDI client
type DDIClientConfig struct {
	PrivateKeyPEM string
	DID           string
	ContractAddress string
}

// DDIClient provides client-side functions for DID operations
type DDIClient struct {
	privateKey *ecdsa.PrivateKey
	did        string
	contractAddress string
	blockchainClient *BlockchainClient
}

// DDIPermission represents a permission in the DDI system
type DDIPermission struct {
	Action     string   `json:"action"`
	Resource   string   `json:"resource"`
	Conditions []string `json:"conditions,omitempty"`
	Expiry     int64    `json:"expiry,omitempty"`
}

// DDIVerifiableCredential represents a verifiable credential in the DDI system
type DDIVerifiableCredential struct {
	ID           string                 `json:"id"`
	Type         []string               `json:"type"`
	Issuer       string                 `json:"issuer"`
	IssuanceDate time.Time              `json:"issuanceDate"`
	ExpiryDate   time.Time              `json:"expirationDate,omitempty"`
	Subject      map[string]interface{} `json:"credentialSubject"`
	Proof        DDIProof               `json:"proof,omitempty"`
}

// DDIProof represents a cryptographic proof
type DDIProof struct {
	Type               string    `json:"type"`
	Created            time.Time `json:"created"`
	VerificationMethod string    `json:"verificationMethod"`
	ProofPurpose       string    `json:"proofPurpose"`
	ProofValue         string    `json:"proofValue"`
}

// NewDDIClient creates a new DDI client
func NewDDIClient(config DDIClientConfig, blockchainClient *BlockchainClient) (*DDIClient, error) {
	// Parse private key from PEM
	if config.PrivateKeyPEM == "" {
		return nil, errors.New("private key is required")
	}
	
	block, _ := pem.Decode([]byte(config.PrivateKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing private key")
	}
	
	privateKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	
	return &DDIClient{
		privateKey: privateKey,
		did:        config.DID,
		contractAddress: config.ContractAddress,
		blockchainClient: blockchainClient,
	}, nil
}

// GenerateProof generates a proof for DID authentication
func (dc *DDIClient) GenerateProof() (string, error) {
	// Create a message to sign (DID + current date)
	message := dc.did + time.Now().Format("2006-01-02")
	messageHash := sha256.Sum256([]byte(message))
	
	// Sign the message
	r, s, err := ecdsa.Sign(rand.Reader, dc.privateKey, messageHash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %v", err)
	}
	
	// Combine r and s to create the signature
	signature := append(r.Bytes(), s.Bytes()...)
	
	// Encode the signature as base64
	proofValue := base64.StdEncoding.EncodeToString(signature)
	
	return proofValue, nil
}

// CheckPermission checks if the DID has a specific permission
func (dc *DDIClient) CheckPermission(action, resource string) (bool, error) {
	if dc.blockchainClient == nil {
		return false, errors.New("blockchain client not initialized")
	}
	
	if dc.contractAddress == "" {
		return false, errors.New("contract address not specified")
	}
	
	// Create function call parameters
	functionSignature := "hasPermission(string,string,string)"
	params := []interface{}{
		dc.did,
		action,
		resource,
	}
	
	// Call contract
	result, err := dc.blockchainClient.CallContract(
		dc.contractAddress,
		functionSignature,
		params,
	)
	
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %v", err)
	}
	
	// Parse result as boolean
	hasPermission, ok := result.(bool)
	if !ok {
		return false, errors.New("unexpected result type from contract")
	}
	
	return hasPermission, nil
}

// VerifyTransaction signs and verifies a transaction using DDI permissions
func (dc *DDIClient) VerifyTransaction(action, resource string, data interface{}) (bool, error) {
	if dc.blockchainClient == nil {
		return false, errors.New("blockchain client not initialized")
	}
	
	// Check permission first
	hasPermission, err := dc.CheckPermission(action, resource)
	if err != nil {
		return false, err
	}
	
	if !hasPermission {
		return false, fmt.Errorf("DID %s does not have permission for %s on %s", dc.did, action, resource)
	}
	
	// Create message to sign
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return false, fmt.Errorf("failed to marshal data: %v", err)
	}
	
	// Create message hash
	messageHash := sha256.Sum256(dataBytes)
	
	// Sign the message
	r, s, err := ecdsa.Sign(rand.Reader, dc.privateKey, messageHash[:])
	if err != nil {
		return false, fmt.Errorf("failed to sign message: %v", err)
	}
	
	// Combine r and s to create the signature
	signature := append(r.Bytes(), s.Bytes()...)
		// Encode the signature as hex
	signatureHex := hex.EncodeToString(signature)
	
	// Log the signature hex for debugging purposes
	log.Printf("Transaction signed with signature: %s", signatureHex)
	
	// Create function call parameters for verification
	functionSignature := "verifyTransaction(string,string,string,bytes,bytes32)"
	params := []interface{}{
		dc.did,
		action,
		resource,
		signature,
		messageHash[:],
	}
	
	// Call contract for verification
	result, err := dc.blockchainClient.CallContract(
		dc.contractAddress,
		functionSignature,
		params,
	)
	
	if err != nil {
		return false, fmt.Errorf("failed to verify transaction: %v", err)
	}
	
	// Parse result as boolean
	isValid, ok := result.(bool)
	if !ok {
		return false, errors.New("unexpected result type from contract")
	}
	
	return isValid, nil
}

// CreateVerifiableCredential creates a new verifiable credential
func (dc *DDIClient) CreateVerifiableCredential(subjectDID string, claims map[string]interface{}, expiryDays int) (*DDIVerifiableCredential, error) {
	// Create unique ID for credential
	credentialID := fmt.Sprintf("urn:uuid:%s", generateUUID())
	
	// Create credential
	now := time.Now().UTC()
	expiry := now.AddDate(0, 0, expiryDays)
	
	credential := &DDIVerifiableCredential{
		ID: credentialID,
		Type: []string{"VerifiableCredential", "TracePostCredential"},
		Issuer: dc.did,
		IssuanceDate: now,
		ExpiryDate: expiry,
		Subject: map[string]interface{}{
			"id": subjectDID,
		},
	}
	
	// Add claims to subject
	for k, v := range claims {
		credential.Subject[k] = v
	}
	
	// Create proof
	credentialBytes, err := json.Marshal(credential)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential: %v", err)
	}
	
	// Hash credential
	hash := sha256.Sum256(credentialBytes)
	
	// Sign hash
	r, s, err := ecdsa.Sign(rand.Reader, dc.privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign credential: %v", err)
	}
	
	// Create proof
	signature := append(r.Bytes(), s.Bytes()...)
	proofValue := base64.StdEncoding.EncodeToString(signature)
	
	credential.Proof = DDIProof{
		Type: "EcdsaSecp256k1Signature2019",
		Created: now,
		VerificationMethod: dc.did + "#keys-1",
		ProofPurpose: "assertionMethod",
		ProofValue: proofValue,
	}
	
	return credential, nil
}

// GenerateKeyPair generates a new ECDSA key pair for DID creation
func GenerateKeyPair() (string, string, error) {
	// Generate a new ECDSA private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key pair: %v", err)
	}
	
	// Marshal private key to PEM
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal private key: %v", err)
	}
	
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})
	
	// Generate public key hex
	pubKeyBytes := elliptic.Marshal(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)
	
	return string(privateKeyPEM), pubKeyHex, nil
}

// RegisterDID registers a new DID using the generated key pair
func RegisterDID(nodeURL, accountAddr, chainID, consensusType, entityType, entityName string) (string, string, error) {	// Generate key pair
	privateKeyPEM, _, err := GenerateKeyPair()
	if err != nil {
		return "", "", err
	}
	
	// Create DID format example: did:tracepost:<entity-type>:<hex-encoded-public-key>
	// The actual DID will be created by the identity client
	
	// Initialize blockchain client
	blockchainClient := NewBlockchainClient(
		nodeURL,
		"", // We don't use the private key for the blockchain client here
		accountAddr,
		chainID,
		consensusType,
	)
	
	// Create identity client
	identityClient := NewIdentityClient(blockchainClient, "")
	
	// Create metadata
	metadata := map[string]interface{}{
		"name": entityName,
		"type": entityType,
	}
	
	// Register DID on blockchain
	didDoc, err := identityClient.CreateDecentralizedID(entityType, entityName, metadata)
	if err != nil {
		return "", "", fmt.Errorf("failed to register DID: %v", err)
	}
	
	return didDoc.DID, privateKeyPEM, nil
}

// Helper functions

// generateUUID generates a random UUID
func generateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", 
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
