package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// DDIClientConfig represents configuration for a DDI client
type DDIClientConfig struct {
	PrivateKeyPEM string
	DID           string
}

// DDIClient provides client-side functions for DID operations
type DDIClient struct {
	privateKey *ecdsa.PrivateKey
	did        string
}

// NewDDIClient creates a new DDI client
func NewDDIClient(config DDIClientConfig) (*DDIClient, error) {
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
