// w3c_did.go
package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// W3CDIDClient provides W3C-compliant DID operations
type W3CDIDClient struct {
	// Base identity client
	BaseClient *IdentityClient
	
	// Supported DID methods
	SupportedMethods map[string]DIDMethodHandler
	
	// DID document store
	Documents map[string]*W3CDIDDocument
}

// DIDMethodHandler is a function type for handling DID method operations
type DIDMethodHandler interface {
	Create(options map[string]interface{}) (*W3CDIDDocument, *ecdsa.PrivateKey, error)
	Resolve(did string) (*W3CDIDDocument, error)
	Update(did string, document *W3CDIDDocument, privateKey *ecdsa.PrivateKey) error
	Deactivate(did string, privateKey *ecdsa.PrivateKey) error
}

// W3CDIDDocument represents a W3C-compliant DID document
type W3CDIDDocument struct {
	Context            []string                   `json:"@context"`
	ID                 string                     `json:"id"`
	Controller         []string                   `json:"controller,omitempty"`
	AlsoKnownAs        []string                   `json:"alsoKnownAs,omitempty"`
	VerificationMethod []W3CVerificationMethod    `json:"verificationMethod,omitempty"`
	Authentication     []string                   `json:"authentication,omitempty"`
	AssertionMethod    []string                   `json:"assertionMethod,omitempty"`
	KeyAgreement       []string                   `json:"keyAgreement,omitempty"`
	CapabilityInvocation []string                 `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []string                 `json:"capabilityDelegation,omitempty"`
	Service            []W3CService               `json:"service,omitempty"`
	Created            time.Time                  `json:"created,omitempty"`
	Updated            time.Time                  `json:"updated,omitempty"`
	Proof              *W3CProof                  `json:"proof,omitempty"`
}

// W3CVerificationMethod represents a W3C-compliant verification method
type W3CVerificationMethod struct {
	ID              string                    `json:"id"`
	Type            string                    `json:"type"`
	Controller      string                    `json:"controller"`
	PublicKeyJwk    map[string]interface{}    `json:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string                 `json:"publicKeyMultibase,omitempty"`
}

// W3CService represents a W3C-compliant service endpoint
type W3CService struct {
	ID              string                    `json:"id"`
	Type            string                    `json:"type"`
	ServiceEndpoint interface{}               `json:"serviceEndpoint"`
}

// W3CProof represents a W3C-compliant proof for a DID document
type W3CProof struct {
	Type            string                    `json:"type"`
	Created         time.Time                 `json:"created"`
	VerificationMethod string                 `json:"verificationMethod"`
	ProofPurpose    string                    `json:"proofPurpose"`
	ProofValue      string                    `json:"proofValue"`
}

// TracePostDIDMethod implements the did:tracepost method
type TracePostDIDMethod struct {
	Client *W3CDIDClient
}

// NewW3CDIDClient creates a new W3C DID client
func NewW3CDIDClient(baseClient *IdentityClient) *W3CDIDClient {
	client := &W3CDIDClient{
		BaseClient:      baseClient,
		SupportedMethods: make(map[string]DIDMethodHandler),
		Documents:       make(map[string]*W3CDIDDocument),
	}
	
	// Register the tracepost DID method
	client.SupportedMethods["tracepost"] = &TracePostDIDMethod{
		Client: client,
	}
	
	return client
}

// Create creates a new W3C-compliant DID
func (w *W3CDIDClient) Create(method string, options map[string]interface{}) (*W3CDIDDocument, *ecdsa.PrivateKey, error) {
	methodHandler, ok := w.SupportedMethods[method]
	if !ok {
		return nil, nil, fmt.Errorf("unsupported DID method: %s", method)
	}
	
	// Create DID document using the method handler
	document, privateKey, err := methodHandler.Create(options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DID: %w", err)
	}
	
	// Store the document
	w.Documents[document.ID] = document
	
	return document, privateKey, nil
}

// Resolve resolves a W3C-compliant DID to a DID document
func (w *W3CDIDClient) Resolve(did string) (*W3CDIDDocument, error) {
	// Check if DID is in cache
	if document, ok := w.Documents[did]; ok {
		return document, nil
	}
	
	// Parse method from DID
	method, err := parseMethod(did)
	if err != nil {
		return nil, err
	}
	
	// Check if method is supported
	methodHandler, ok := w.SupportedMethods[method]
	if !ok {
		return nil, fmt.Errorf("unsupported DID method: %s", method)
	}
	
	// Resolve DID document using the method handler
	document, err := methodHandler.Resolve(did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve DID: %w", err)
	}
	
	// Store the document in cache
	w.Documents[did] = document
	
	return document, nil
}

// Update updates a W3C-compliant DID document
func (w *W3CDIDClient) Update(did string, document *W3CDIDDocument, privateKey *ecdsa.PrivateKey) error {
	// Parse method from DID
	method, err := parseMethod(did)
	if err != nil {
		return err
	}
	
	// Check if method is supported
	methodHandler, ok := w.SupportedMethods[method]
	if !ok {
		return fmt.Errorf("unsupported DID method: %s", method)
	}
	
	// Update DID document using the method handler
	err = methodHandler.Update(did, document, privateKey)
	if err != nil {
		return fmt.Errorf("failed to update DID: %w", err)
	}
	
	// Update the document in cache
	w.Documents[did] = document
	
	return nil
}

// Deactivate deactivates a W3C-compliant DID
func (w *W3CDIDClient) Deactivate(did string, privateKey *ecdsa.PrivateKey) error {
	// Parse method from DID
	method, err := parseMethod(did)
	if err != nil {
		return err
	}
	
	// Check if method is supported
	methodHandler, ok := w.SupportedMethods[method]
	if !ok {
		return fmt.Errorf("unsupported DID method: %s", method)
	}
	
	// Deactivate DID using the method handler
	err = methodHandler.Deactivate(did, privateKey)
	if err != nil {
		return fmt.Errorf("failed to deactivate DID: %w", err)
	}
	
	// Remove the document from cache
	delete(w.Documents, did)
	
	return nil
}

// parseMethod parses the method from a DID
func parseMethod(did string) (string, error) {
	// Parse DID to extract method
	// DID format: did:<method>:<method-specific-id>
	parts := strings.Split(did, ":")
	if len(parts) < 3 || parts[0] != "did" {
		return "", errors.New("invalid DID format")
	}
	
	return parts[1], nil
}

// Create implements the did:tracepost method Create operation
func (t *TracePostDIDMethod) Create(options map[string]interface{}) (*W3CDIDDocument, *ecdsa.PrivateKey, error) {
	// Generate a new key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
		// Convert public key to JWK
	b64url := base64url{}
	publicKeyJwk := map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   b64url.Encode(privateKey.PublicKey.X.Bytes()),
		"y":   b64url.Encode(privateKey.PublicKey.Y.Bytes()),
	}
	
	// Get controller from options
	controller := ""
	if controllerOption, ok := options["controller"].(string); ok {
		controller = controllerOption
	}
	
	// Create method-specific identifier
	idBytes := make([]byte, 16)
	rand.Read(idBytes)
	methodSpecificId := fmt.Sprintf("%x", idBytes)
	
	// Create DID
	did := fmt.Sprintf("did:tracepost:%s", methodSpecificId)
	
	// Create verification method ID
	verificationMethodId := fmt.Sprintf("%s#keys-1", did)
	
	// Create document
	now := time.Now()
	document := &W3CDIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/jws-2020/v1",
		},
		ID: did,
		Controller: []string{
			controller,
		},
		VerificationMethod: []W3CVerificationMethod{
			{
				ID:           verificationMethodId,
				Type:         "JsonWebKey2020",
				Controller:   did,
				PublicKeyJwk: publicKeyJwk,
			},
		},
		Authentication: []string{
			verificationMethodId,
		},
		AssertionMethod: []string{
			verificationMethodId,
		},
		Service: []W3CService{
			{
				ID:   fmt.Sprintf("%s#logistics-service", did),
				Type: "LogisticsService",
				ServiceEndpoint: map[string]interface{}{
					"url": "https://api.tracepost.vn/logistics",
				},
			},
		},
		Created: now,
		Updated: now,
	}
	
	// Create proof
	proof, err := createProof(document, privateKey, verificationMethodId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create proof: %w", err)
	}
	
	document.Proof = proof
	
	return document, privateKey, nil
}

// createProof creates a proof for a W3C DID document
func createProof(document *W3CDIDDocument, privateKey *ecdsa.PrivateKey, verificationMethodId string) (*W3CProof, error) {
	// Make a copy of the document without the proof
	documentCopy := *document
	documentCopy.Proof = nil
	
	// Canonicalize the document
	bytes, err := canonicalize(documentCopy)
	if err != nil {
		return nil, fmt.Errorf("failed to canonicalize document: %w", err)
	}
	
	// Hash the canonicalized document
	hash := sha256.Sum256(bytes)
	
	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}
	
	// Create JWS
	jws := createJWS(r, s, hash[:])
	
	// Create proof
	proof := &W3CProof{
		Type:               "JsonWebSignature2020",
		Created:            time.Now(),
		VerificationMethod: verificationMethodId,
		ProofPurpose:       "assertionMethod",
		ProofValue:         jws,
	}
	
	return proof, nil
}

// canonicalize canonicalizes a W3C DID document
func canonicalize(document W3CDIDDocument) ([]byte, error) {
	// In a real implementation, this would use a proper canonicalization algorithm
	// such as JSON-LD canonicalization
	// For simplicity, we'll just use JSON marshaling
	return json.Marshal(document)
}

// createJWS creates a JWS from an ECDSA signature
func createJWS(r, s *big.Int, hash []byte) string {
	// In a real implementation, this would create a proper JWS
	// For simplicity, we'll just concatenate r and s and base64url encode
	b64url := base64url{}
	signature := append(r.Bytes(), s.Bytes()...)
	return b64url.Encode(signature)
}

// Resolve implements the did:tracepost method Resolve operation
func (t *TracePostDIDMethod) Resolve(did string) (*W3CDIDDocument, error) {
	// Check if DID is in the client's document store
	if document, ok := t.Client.Documents[did]; ok {
		return document, nil
	}
	
	// In a real implementation, this would resolve the DID from a blockchain or registry
	// For this implementation, if it's not in our local store, we can't resolve it
	return nil, fmt.Errorf("DID not found: %s", did)
}

// Update implements the did:tracepost method Update operation
func (t *TracePostDIDMethod) Update(did string, document *W3CDIDDocument, privateKey *ecdsa.PrivateKey) error {
	// Check if DID exists
	_, err := t.Resolve(did)
	if err != nil {
		return fmt.Errorf("DID not found: %s", did)
	}
	
	// Verify that the private key corresponds to the verification method
	if !verifyKeyOwnership(document, privateKey) {
		return errors.New("private key does not match any verification method")
	}
	
	// Update the document's updated timestamp
	document.Updated = time.Now()
	
	// Create a new proof
	verificationMethodId := fmt.Sprintf("%s#keys-1", did)
	proof, err := createProof(document, privateKey, verificationMethodId)
	if err != nil {
		return fmt.Errorf("failed to create proof: %w", err)
	}
	
	document.Proof = proof
	
	// In a real implementation, this would update the DID on a blockchain or registry
	// For this implementation, we'll just update it in our local store
	t.Client.Documents[did] = document
	
	return nil
}

// verifyKeyOwnership verifies that a private key corresponds to a verification method in the document
func verifyKeyOwnership(document *W3CDIDDocument, privateKey *ecdsa.PrivateKey) bool {
	// Check each verification method
	for _, vm := range document.VerificationMethod {
		if vm.Type == "JsonWebKey2020" && vm.PublicKeyJwk != nil {
			// Extract x and y coordinates from JWK
			xEncoded, xOk := vm.PublicKeyJwk["x"].(string)
			yEncoded, yOk := vm.PublicKeyJwk["y"].(string)
			
			if !xOk || !yOk {
				continue
			}
			
			b64url := base64url{}
			xBytes, err := b64url.Decode(xEncoded)
			if err != nil {
				continue
			}
			
			yBytes, err := b64url.Decode(yEncoded)
			if err != nil {
				continue
			}
			
			x := new(big.Int).SetBytes(xBytes)
			y := new(big.Int).SetBytes(yBytes)
			
			// Check if the coordinates match the private key's public key
			if x.Cmp(privateKey.PublicKey.X) == 0 && y.Cmp(privateKey.PublicKey.Y) == 0 {
				return true
			}
		}
	}
	
	return false
}

// Deactivate implements the did:tracepost method Deactivate operation
func (t *TracePostDIDMethod) Deactivate(did string, privateKey *ecdsa.PrivateKey) error {
	// Check if DID exists
	document, err := t.Resolve(did)
	if err != nil {
		return fmt.Errorf("DID not found: %s", did)
	}
	
	// Verify that the private key corresponds to the verification method
	if !verifyKeyOwnership(document, privateKey) {
		return errors.New("private key does not match any verification method")
	}
	
	// In a real implementation, this would deactivate the DID on a blockchain or registry
	// For this implementation, we'll just remove it from our local store
	delete(t.Client.Documents, did)
	
	return nil
}

// base64url provides base64url encoding and decoding
type base64url struct{}

// Encode encodes data using base64url encoding
func (base64url) Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Decode decodes data using base64url encoding
func (base64url) Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
