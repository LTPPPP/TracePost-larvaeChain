// hsm.go
package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

// HSMType defines the type of HSM implementation
type HSMType string

const (
	// HSMTypeSoftware is a software-based HSM simulation (for development/testing)
	HSMTypeSoftware HSMType = "software"
	// HSMTypeAWS is AWS CloudHSM implementation
	HSMTypeAWS HSMType = "aws"
	// HSMTypeAzure is Azure KeyVault implementation
	HSMTypeAzure HSMType = "azure"
	// HSMTypeGCP is Google Cloud KMS implementation
	HSMTypeGCP HSMType = "gcp"
	// HSMTypeThales is Thales Luna HSM implementation
	HSMTypeThales HSMType = "thales"
)

// HSMConfig contains configuration for the HSM
type HSMConfig struct {
	// Type of HSM to use
	Type HSMType
	
	// Connection details
	Endpoint   string
	Region     string
	ProjectID  string
	
	// Authentication
	APIKey     string
	APISecret  string
	
	// Key management
	KeyID      string
	Slot       int
	UserPin    string
	
	// Performance
	CacheDuration time.Duration
}

// HSMService provides hardware security module integration
type HSMService struct {
	config HSMConfig
	mutex  sync.RWMutex
	
	// Cache of key handles and session information
	keyCache   map[string]interface{}
	sessionID  string
	sessionExp time.Time
}

// HSMKeyInfo contains metadata about keys stored in HSM
type HSMKeyInfo struct {
	KeyID      string
	Algorithm  string
	KeySize    int
	Created    time.Time
	LastUsed   time.Time
	Usage      string
	Exportable bool
}

// NewHSMService creates a new HSM service
func NewHSMService(config HSMConfig) (*HSMService, error) {
	// Validate config
	if config.Type == "" {
		return nil, errors.New("HSM type is required")
	}
	
	// Set default cache duration if not specified
	if config.CacheDuration == 0 {
		config.CacheDuration = 15 * time.Minute
	}
	
	// Initialize HSM service
	service := &HSMService{
		config:   config,
		keyCache: make(map[string]interface{}),
	}
	
	// Initialize connection to HSM
	if err := service.initializeConnection(); err != nil {
		return nil, fmt.Errorf("failed to initialize HSM connection: %w", err)
	}
	
	return service, nil
}

// initializeConnection establishes a connection to the HSM
func (h *HSMService) initializeConnection() error {
	switch h.config.Type {
	case HSMTypeSoftware:
		// No actual connection for software HSM
		h.sessionID = fmt.Sprintf("sim-%x", time.Now().UnixNano())
		h.sessionExp = time.Now().Add(h.config.CacheDuration)
		return nil
		
	case HSMTypeAWS:
		// Connect to AWS CloudHSM
		return h.initAWSConnection()
		
	case HSMTypeAzure:
		// Connect to Azure KeyVault
		return h.initAzureConnection()
		
	case HSMTypeGCP:
		// Connect to Google Cloud KMS
		return h.initGCPConnection()
		
	case HSMTypeThales:
		// Connect to Thales Luna HSM
		return h.initThalesConnection()
		
	default:
		return fmt.Errorf("unsupported HSM type: %s", h.config.Type)
	}
}

// initAWSConnection initializes connection to AWS CloudHSM
func (h *HSMService) initAWSConnection() error {
	// AWS CloudHSM connection implementation would go here
	// In production, this would use AWS SDK to establish a connection
	
	// For now, we'll just simulate success and set session information
	h.sessionID = fmt.Sprintf("aws-%x", time.Now().UnixNano())
	h.sessionExp = time.Now().Add(h.config.CacheDuration)
	return nil
}

// initAzureConnection initializes connection to Azure KeyVault
func (h *HSMService) initAzureConnection() error {
	// Azure KeyVault connection implementation would go here
	// In production, this would use Azure SDK to establish a connection
	
	// For now, we'll just simulate success and set session information
	h.sessionID = fmt.Sprintf("azure-%x", time.Now().UnixNano())
	h.sessionExp = time.Now().Add(h.config.CacheDuration)
	return nil
}

// initGCPConnection initializes connection to Google Cloud KMS
func (h *HSMService) initGCPConnection() error {
	// GCP KMS connection implementation would go here
	// In production, this would use GCP SDK to establish a connection
	
	// For now, we'll just simulate success and set session information
	h.sessionID = fmt.Sprintf("gcp-%x", time.Now().UnixNano())
	h.sessionExp = time.Now().Add(h.config.CacheDuration)
	return nil
}

// initThalesConnection initializes connection to Thales Luna HSM
func (h *HSMService) initThalesConnection() error {
	// Thales Luna HSM connection implementation would go here
	// In production, this would use PKCS#11 to establish a connection
	
	// For now, we'll just simulate success and set session information
	h.sessionID = fmt.Sprintf("thales-%x", time.Now().UnixNano())
	h.sessionExp = time.Now().Add(h.config.CacheDuration)
	return nil
}

// CreateKey creates a new key in the HSM
func (h *HSMService) CreateKey(keyID, algorithm string, keySize int) (string, error) {
	// Check if session needs renewal
	if err := h.ensureValidSession(); err != nil {
		return "", err
	}
	
	// Implementation varies by HSM type
	switch h.config.Type {
	case HSMTypeSoftware:
		// For software HSM, just generate a key pair locally
		return h.createSoftwareKey(keyID, algorithm, keySize)
		
	case HSMTypeAWS:
		// For AWS CloudHSM
		return h.createAWSKey(keyID, algorithm, keySize)
		
	case HSMTypeAzure:
		// For Azure KeyVault
		return h.createAzureKey(keyID, algorithm, keySize)
		
	case HSMTypeGCP:
		// For Google Cloud KMS
		return h.createGCPKey(keyID, algorithm, keySize)
		
	case HSMTypeThales:
		// For Thales Luna HSM
		return h.createThalesKey(keyID, algorithm, keySize)
		
	default:
		return "", fmt.Errorf("unsupported HSM type: %s", h.config.Type)
	}
}

// createSoftwareKey creates a new software key (simulation for development)
func (h *HSMService) createSoftwareKey(keyID, algorithm string, keySize int) (string, error) {
	// Generate a new ECDSA key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate key pair: %w", err)
	}
	
	// Store private key in cache
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Format keyID with prefix to avoid collisions
	fullKeyID := fmt.Sprintf("sim:%s", keyID)
	h.keyCache[fullKeyID] = privateKey
	
	return fullKeyID, nil
}

// createAWSKey creates a new key in AWS CloudHSM
func (h *HSMService) createAWSKey(keyID, algorithm string, keySize int) (string, error) {
	// AWS CloudHSM key creation implementation would go here
	// This would use AWS CloudHSM API to create a key
	
	// For now, we'll just simulate success and return a formatted key ID
	return fmt.Sprintf("aws:%s", keyID), nil
}

// createAzureKey creates a new key in Azure KeyVault
func (h *HSMService) createAzureKey(keyID, algorithm string, keySize int) (string, error) {
	// Azure KeyVault key creation implementation would go here
	// This would use Azure KeyVault API to create a key
	
	// For now, we'll just simulate success and return a formatted key ID
	return fmt.Sprintf("azure:%s", keyID), nil
}

// createGCPKey creates a new key in Google Cloud KMS
func (h *HSMService) createGCPKey(keyID, algorithm string, keySize int) (string, error) {
	// Google Cloud KMS key creation implementation would go here
	// This would use Google Cloud KMS API to create a key
	
	// For now, we'll just simulate success and return a formatted key ID
	return fmt.Sprintf("gcp:%s", keyID), nil
}

// createThalesKey creates a new key in Thales Luna HSM
func (h *HSMService) createThalesKey(keyID, algorithm string, keySize int) (string, error) {
	// Thales Luna HSM key creation implementation would go here
	// This would use PKCS#11 API to create a key
	
	// For now, we'll just simulate success and return a formatted key ID
	return fmt.Sprintf("thales:%s", keyID), nil
}

// Sign generates a signature using a key stored in the HSM
func (h *HSMService) Sign(keyID string, data []byte) ([]byte, error) {
	// Check if session needs renewal
	if err := h.ensureValidSession(); err != nil {
		return nil, err
	}
	
	// Implementation varies by HSM type and key prefix
	if strings.HasPrefix(keyID, "sim:") {
		return h.signSoftware(keyID, data)
	} else if strings.HasPrefix(keyID, "aws:") {
		return h.signAWS(keyID, data)
	} else if strings.HasPrefix(keyID, "azure:") {
		return h.signAzure(keyID, data)
	} else if strings.HasPrefix(keyID, "gcp:") {
		return h.signGCP(keyID, data)
	} else if strings.HasPrefix(keyID, "thales:") {
		return h.signThales(keyID, data)
	} else {
		return nil, fmt.Errorf("unsupported key format: %s", keyID)
	}
}

// signSoftware signs data using a software key (simulation for development)
func (h *HSMService) signSoftware(keyID string, data []byte) ([]byte, error) {
	h.mutex.RLock()
	privateKey, ok := h.keyCache[keyID]
	h.mutex.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	
	ecdsaKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid key type in cache")
	}
	
	// Hash the data
	hash := sha256.Sum256(data)
	
	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, ecdsaKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	
	// Combine r and s into a signature
	signature := append(r.Bytes(), s.Bytes()...)
	
	return signature, nil
}

// signAWS signs data using a key in AWS CloudHSM
func (h *HSMService) signAWS(keyID string, data []byte) ([]byte, error) {
	// AWS CloudHSM signing implementation would go here
	// This would use AWS CloudHSM API to sign data
	
	// For now, we'll just simulate success and return a dummy signature
	dummySig := make([]byte, 64)
	rand.Read(dummySig)
	return dummySig, nil
}

// signAzure signs data using a key in Azure KeyVault
func (h *HSMService) signAzure(keyID string, data []byte) ([]byte, error) {
	// Azure KeyVault signing implementation would go here
	// This would use Azure KeyVault API to sign data
	
	// For now, we'll just simulate success and return a dummy signature
	dummySig := make([]byte, 64)
	rand.Read(dummySig)
	return dummySig, nil
}

// signGCP signs data using a key in Google Cloud KMS
func (h *HSMService) signGCP(keyID string, data []byte) ([]byte, error) {
	// Google Cloud KMS signing implementation would go here
	// This would use Google Cloud KMS API to sign data
	
	// For now, we'll just simulate success and return a dummy signature
	dummySig := make([]byte, 64)
	rand.Read(dummySig)
	return dummySig, nil
}

// signThales signs data using a key in Thales Luna HSM
func (h *HSMService) signThales(keyID string, data []byte) ([]byte, error) {
	// Thales Luna HSM signing implementation would go here
	// This would use PKCS#11 API to sign data
	
	// For now, we'll just simulate success and return a dummy signature
	dummySig := make([]byte, 64)
	rand.Read(dummySig)
	return dummySig, nil
}

// Verify verifies a signature against data
func (h *HSMService) Verify(keyID string, data, signature []byte) (bool, error) {
	// Check if session needs renewal
	if err := h.ensureValidSession(); err != nil {
		return false, err
	}
	
	// Implementation varies by HSM type and key prefix
	if strings.HasPrefix(keyID, "sim:") {
		return h.verifySoftware(keyID, data, signature)
	} else if strings.HasPrefix(keyID, "aws:") {
		return h.verifyAWS(keyID, data, signature)
	} else if strings.HasPrefix(keyID, "azure:") {
		return h.verifyAzure(keyID, data, signature)
	} else if strings.HasPrefix(keyID, "gcp:") {
		return h.verifyGCP(keyID, data, signature)
	} else if strings.HasPrefix(keyID, "thales:") {
		return h.verifyThales(keyID, data, signature)
	} else {
		return false, fmt.Errorf("unsupported key format: %s", keyID)
	}
}

// verifySoftware verifies a signature using a software key (simulation for development)
func (h *HSMService) verifySoftware(keyID string, data, signature []byte) (bool, error) {
	h.mutex.RLock()
	key, ok := h.keyCache[keyID]
	h.mutex.RUnlock()
	
	if !ok {
		return false, fmt.Errorf("key not found: %s", keyID)
	}
	
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return false, errors.New("invalid key type in cache")
	}
	
	// Get the public key
	publicKey := &ecdsaKey.PublicKey
	
	// Hash the data
	hash := sha256.Sum256(data)
	
	// Split the signature into r and s
	if len(signature) != 64 {
		return false, errors.New("invalid signature length")
	}
	
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])
	
	// Verify the signature
	return ecdsa.Verify(publicKey, hash[:], r, s), nil
}

// verifyAWS verifies a signature using a key in AWS CloudHSM
func (h *HSMService) verifyAWS(keyID string, data, signature []byte) (bool, error) {
	// AWS CloudHSM verification implementation would go here
	// This would use AWS CloudHSM API to verify a signature
	
	// For now, we'll just simulate success
	return true, nil
}

// verifyAzure verifies a signature using a key in Azure KeyVault
func (h *HSMService) verifyAzure(keyID string, data, signature []byte) (bool, error) {
	// Azure KeyVault verification implementation would go here
	// This would use Azure KeyVault API to verify a signature
	
	// For now, we'll just simulate success
	return true, nil
}

// verifyGCP verifies a signature using a key in Google Cloud KMS
func (h *HSMService) verifyGCP(keyID string, data, signature []byte) (bool, error) {
	// Google Cloud KMS verification implementation would go here
	// This would use Google Cloud KMS API to verify a signature
	
	// For now, we'll just simulate success
	return true, nil
}

// verifyThales verifies a signature using a key in Thales Luna HSM
func (h *HSMService) verifyThales(keyID string, data, signature []byte) (bool, error) {
	// Thales Luna HSM verification implementation would go here
	// This would use PKCS#11 API to verify a signature
	
	// For now, we'll just simulate success
	return true, nil
}

// GetPublicKey retrieves the public key associated with a key in the HSM
func (h *HSMService) GetPublicKey(keyID string) ([]byte, error) {
	// Check if session needs renewal
	if err := h.ensureValidSession(); err != nil {
		return nil, err
	}
	
	// Implementation varies by HSM type and key prefix
	if strings.HasPrefix(keyID, "sim:") {
		return h.getPublicKeySoftware(keyID)
	} else if strings.HasPrefix(keyID, "aws:") {
		return h.getPublicKeyAWS(keyID)
	} else if strings.HasPrefix(keyID, "azure:") {
		return h.getPublicKeyAzure(keyID)
	} else if strings.HasPrefix(keyID, "gcp:") {
		return h.getPublicKeyGCP(keyID)
	} else if strings.HasPrefix(keyID, "thales:") {
		return h.getPublicKeyThales(keyID)
	} else {
		return nil, fmt.Errorf("unsupported key format: %s", keyID)
	}
}

// getPublicKeySoftware gets the public key for a software key (simulation for development)
func (h *HSMService) getPublicKeySoftware(keyID string) ([]byte, error) {
	h.mutex.RLock()
	key, ok := h.keyCache[keyID]
	h.mutex.RUnlock()
	
	if !ok {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}
	
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("invalid key type in cache")
	}
	
	// Get the public key
	publicKey := &ecdsaKey.PublicKey
	
	// Encode the public key
	return elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y), nil
}

// getPublicKeyAWS gets the public key for a key in AWS CloudHSM
func (h *HSMService) getPublicKeyAWS(keyID string) ([]byte, error) {
	// AWS CloudHSM public key retrieval implementation would go here
	// This would use AWS CloudHSM API to get a public key
	
	// For now, we'll just simulate success and return a dummy public key
	dummyKey := make([]byte, 65) // typical size for a compressed ECDSA public key
	dummyKey[0] = 4              // uncompressed point format
	rand.Read(dummyKey[1:])
	return dummyKey, nil
}

// getPublicKeyAzure gets the public key for a key in Azure KeyVault
func (h *HSMService) getPublicKeyAzure(keyID string) ([]byte, error) {
	// Azure KeyVault public key retrieval implementation would go here
	// This would use Azure KeyVault API to get a public key
	
	// For now, we'll just simulate success and return a dummy public key
	dummyKey := make([]byte, 65) // typical size for a compressed ECDSA public key
	dummyKey[0] = 4              // uncompressed point format
	rand.Read(dummyKey[1:])
	return dummyKey, nil
}

// getPublicKeyGCP gets the public key for a key in Google Cloud KMS
func (h *HSMService) getPublicKeyGCP(keyID string) ([]byte, error) {
	// Google Cloud KMS public key retrieval implementation would go here
	// This would use Google Cloud KMS API to get a public key
	
	// For now, we'll just simulate success and return a dummy public key
	dummyKey := make([]byte, 65) // typical size for a compressed ECDSA public key
	dummyKey[0] = 4              // uncompressed point format
	rand.Read(dummyKey[1:])
	return dummyKey, nil
}

// getPublicKeyThales gets the public key for a key in Thales Luna HSM
func (h *HSMService) getPublicKeyThales(keyID string) ([]byte, error) {
	// Thales Luna HSM public key retrieval implementation would go here
	// This would use PKCS#11 API to get a public key
	
	// For now, we'll just simulate success and return a dummy public key
	dummyKey := make([]byte, 65) // typical size for a compressed ECDSA public key
	dummyKey[0] = 4              // uncompressed point format
	rand.Read(dummyKey[1:])
	return dummyKey, nil
}

// ListKeys lists all keys in the HSM
func (h *HSMService) ListKeys() ([]HSMKeyInfo, error) {
	// Check if session needs renewal
	if err := h.ensureValidSession(); err != nil {
		return nil, err
	}
	
	// Implementation varies by HSM type
	switch h.config.Type {
	case HSMTypeSoftware:
		return h.listSoftwareKeys()
		
	case HSMTypeAWS:
		return h.listAWSKeys()
		
	case HSMTypeAzure:
		return h.listAzureKeys()
		
	case HSMTypeGCP:
		return h.listGCPKeys()
		
	case HSMTypeThales:
		return h.listThalesKeys()
		
	default:
		return nil, fmt.Errorf("unsupported HSM type: %s", h.config.Type)
	}
}

// listSoftwareKeys lists all software keys (simulation for development)
func (h *HSMService) listSoftwareKeys() ([]HSMKeyInfo, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	keys := make([]HSMKeyInfo, 0, len(h.keyCache))
	for keyID := range h.keyCache {
		keys = append(keys, HSMKeyInfo{
			KeyID:      keyID,
			Algorithm:  "ECDSA",
			KeySize:    256,
			Created:    time.Now().Add(-24 * time.Hour), // Simulate created 1 day ago
			LastUsed:   time.Now(),
			Usage:      "sign,verify",
			Exportable: false,
		})
	}
	
	return keys, nil
}

// listAWSKeys lists all keys in AWS CloudHSM
func (h *HSMService) listAWSKeys() ([]HSMKeyInfo, error) {
	// AWS CloudHSM key listing implementation would go here
	// This would use AWS CloudHSM API to list keys
	
	// For now, we'll just simulate success and return a dummy list
	return []HSMKeyInfo{
		{
			KeyID:      "aws:dummy-key-1",
			Algorithm:  "ECDSA",
			KeySize:    256,
			Created:    time.Now().Add(-24 * time.Hour),
			LastUsed:   time.Now(),
			Usage:      "sign,verify",
			Exportable: false,
		},
	}, nil
}

// listAzureKeys lists all keys in Azure KeyVault
func (h *HSMService) listAzureKeys() ([]HSMKeyInfo, error) {
	// Azure KeyVault key listing implementation would go here
	// This would use Azure KeyVault API to list keys
	
	// For now, we'll just simulate success and return a dummy list
	return []HSMKeyInfo{
		{
			KeyID:      "azure:dummy-key-1",
			Algorithm:  "ECDSA",
			KeySize:    256,
			Created:    time.Now().Add(-24 * time.Hour),
			LastUsed:   time.Now(),
			Usage:      "sign,verify",
			Exportable: false,
		},
	}, nil
}

// listGCPKeys lists all keys in Google Cloud KMS
func (h *HSMService) listGCPKeys() ([]HSMKeyInfo, error) {
	// Google Cloud KMS key listing implementation would go here
	// This would use Google Cloud KMS API to list keys
	
	// For now, we'll just simulate success and return a dummy list
	return []HSMKeyInfo{
		{
			KeyID:      "gcp:dummy-key-1",
			Algorithm:  "ECDSA",
			KeySize:    256,
			Created:    time.Now().Add(-24 * time.Hour),
			LastUsed:   time.Now(),
			Usage:      "sign,verify",
			Exportable: false,
		},
	}, nil
}

// listThalesKeys lists all keys in Thales Luna HSM
func (h *HSMService) listThalesKeys() ([]HSMKeyInfo, error) {
	// Thales Luna HSM key listing implementation would go here
	// This would use PKCS#11 API to list keys
	
	// For now, we'll just simulate success and return a dummy list
	return []HSMKeyInfo{
		{
			KeyID:      "thales:dummy-key-1",
			Algorithm:  "ECDSA",
			KeySize:    256,
			Created:    time.Now().Add(-24 * time.Hour),
			LastUsed:   time.Now(),
			Usage:      "sign,verify",
			Exportable: false,
		},
	}, nil
}

// DeleteKey deletes a key from the HSM
func (h *HSMService) DeleteKey(keyID string) error {
	// Check if session needs renewal
	if err := h.ensureValidSession(); err != nil {
		return err
	}
	
	// Implementation varies by HSM type and key prefix
	if strings.HasPrefix(keyID, "sim:") {
		return h.deleteSoftwareKey(keyID)
	} else if strings.HasPrefix(keyID, "aws:") {
		return h.deleteAWSKey(keyID)
	} else if strings.HasPrefix(keyID, "azure:") {
		return h.deleteAzureKey(keyID)
	} else if strings.HasPrefix(keyID, "gcp:") {
		return h.deleteGCPKey(keyID)
	} else if strings.HasPrefix(keyID, "thales:") {
		return h.deleteThalesKey(keyID)
	} else {
		return fmt.Errorf("unsupported key format: %s", keyID)
	}
}

// deleteSoftwareKey deletes a software key (simulation for development)
func (h *HSMService) deleteSoftwareKey(keyID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if _, ok := h.keyCache[keyID]; !ok {
		return fmt.Errorf("key not found: %s", keyID)
	}
	
	delete(h.keyCache, keyID)
	return nil
}

// deleteAWSKey deletes a key from AWS CloudHSM
func (h *HSMService) deleteAWSKey(keyID string) error {
	// AWS CloudHSM key deletion implementation would go here
	// This would use AWS CloudHSM API to delete a key
	
	// For now, we'll just simulate success
	return nil
}

// deleteAzureKey deletes a key from Azure KeyVault
func (h *HSMService) deleteAzureKey(keyID string) error {
	// Azure KeyVault key deletion implementation would go here
	// This would use Azure KeyVault API to delete a key
	
	// For now, we'll just simulate success
	return nil
}

// deleteGCPKey deletes a key from Google Cloud KMS
func (h *HSMService) deleteGCPKey(keyID string) error {
	// Google Cloud KMS key deletion implementation would go here
	// This would use Google Cloud KMS API to delete a key
	
	// For now, we'll just simulate success
	return nil
}

// deleteThalesKey deletes a key from Thales Luna HSM
func (h *HSMService) deleteThalesKey(keyID string) error {
	// Thales Luna HSM key deletion implementation would go here
	// This would use PKCS#11 API to delete a key
	
	// For now, we'll just simulate success
	return nil
}

// ensureValidSession ensures the HSM session is valid, renewing if necessary
func (h *HSMService) ensureValidSession() error {
	h.mutex.RLock()
	sessionExpired := time.Now().After(h.sessionExp)
	h.mutex.RUnlock()
	
	if sessionExpired {
		h.mutex.Lock()
		defer h.mutex.Unlock()
		
		// Double-check after acquiring lock
		if time.Now().After(h.sessionExp) {
			// Re-initialize connection
			if err := h.initializeConnection(); err != nil {
				return fmt.Errorf("failed to renew HSM connection: %w", err)
			}
		}
	}
	
	return nil
}
