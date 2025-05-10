// ssi.go
package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// SSIClient provides Self-Sovereign Identity capabilities
type SSIClient struct {
	// Base identity client
	BaseClient *IdentityClient
	
	// DID method name
	DIDMethod string
	
	// DID document registry
	DIDRegistry map[string]*DIDDocument
	
	// Verifiable credential registry
	CredentialRegistry map[string]*VerifiableCredential
	
	// Verifiable presentation registry
	PresentationRegistry map[string]*VerifiablePresentation
	
	// Trusted issuers
	TrustedIssuers map[string]bool
}

// DIDDocument represents a decentralized identifier document
type DIDDocument struct {
	ID                 string                 `json:"id"`
	Context            []string               `json:"@context"`
	Controller         string                 `json:"controller,omitempty"`
	VerificationMethod []VerificationMethod   `json:"verificationMethod"`
	Authentication     []string               `json:"authentication"`
	AssertionMethod    []string               `json:"assertionMethod,omitempty"`
	KeyAgreement       []string               `json:"keyAgreement,omitempty"`
	CapabilityInvocation []string             `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []string             `json:"capabilityDelegation,omitempty"`
	Service            []ServiceEndpoint      `json:"service,omitempty"`
	Created            time.Time              `json:"created"`
	Updated            time.Time              `json:"updated"`
	Proof              *DIDDocumentProof      `json:"proof,omitempty"`
}

// VerificationMethod represents a verification method in a DID document
type VerificationMethod struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	Controller      string   `json:"controller"`
	PublicKeyJwk    map[string]interface{} `json:"publicKeyJwk,omitempty"`
	PublicKeyBase58 string   `json:"publicKeyBase58,omitempty"`
	PublicKeyHex    string   `json:"publicKeyHex,omitempty"`
}

// ServiceEndpoint represents a service endpoint in a DID document
type ServiceEndpoint struct {
	ID              string   `json:"id"`
	Type            string   `json:"type"`
	ServiceEndpoint string   `json:"serviceEndpoint"`
	Description     string   `json:"description,omitempty"`
}

// DIDDocumentProof represents a proof in a DID document
type DIDDocumentProof struct {
	Type                 string    `json:"type"`
	Created              time.Time `json:"created"`
	VerificationMethod   string    `json:"verificationMethod"`
	ProofPurpose         string    `json:"proofPurpose"`
	ProofValue           string    `json:"proofValue"`
}

// VerifiableCredential represents a verifiable credential
type VerifiableCredential struct {
	Context           []string               `json:"@context"`
	ID                string                 `json:"id"`
	Type              []string               `json:"type"`
	Issuer            string                 `json:"issuer"`
	IssuanceDate      time.Time              `json:"issuanceDate"`
	ExpirationDate    time.Time              `json:"expirationDate,omitempty"`
	CredentialSubject map[string]interface{} `json:"credentialSubject"`
	Proof             *CredentialProof       `json:"proof"`
	Status            CredentialStatus       `json:"credentialStatus,omitempty"`
}

// CredentialProof represents a proof in a verifiable credential
type CredentialProof struct {
	Type                 string    `json:"type"`
	Created              time.Time `json:"created"`
	VerificationMethod   string    `json:"verificationMethod"`
	ProofPurpose         string    `json:"proofPurpose"`
	ProofValue           string    `json:"proofValue"`
	Challenge            string    `json:"challenge,omitempty"`
	Domain               string    `json:"domain,omitempty"`
}

// CredentialStatus represents the status of a verifiable credential
type CredentialStatus struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// VerifiablePresentation represents a verifiable presentation
type VerifiablePresentation struct {
	Context            []string               `json:"@context"`
	ID                 string                 `json:"id,omitempty"`
	Type               []string               `json:"type"`
	Holder             string                 `json:"holder"`
	VerifiableCredential []*VerifiableCredential `json:"verifiableCredential"`
	Proof              *PresentationProof     `json:"proof"`
}

// PresentationProof represents a proof in a verifiable presentation
type PresentationProof struct {
	Type                 string    `json:"type"`
	Created              time.Time `json:"created"`
	VerificationMethod   string    `json:"verificationMethod"`
	ProofPurpose         string    `json:"proofPurpose"`
	ProofValue           string    `json:"proofValue"`
	Challenge            string    `json:"challenge"`
	Domain               string    `json:"domain"`
}

// NewSSIClient creates a new SSI client
func NewSSIClient(baseClient *IdentityClient, didMethod string) *SSIClient {
	return &SSIClient{
		BaseClient:           baseClient,
		DIDMethod:            didMethod,
		DIDRegistry:          make(map[string]*DIDDocument),
		CredentialRegistry:   make(map[string]*VerifiableCredential),
		PresentationRegistry: make(map[string]*VerifiablePresentation),
		TrustedIssuers:       make(map[string]bool),
	}
}

// CreateDID creates a new decentralized identifier
func (sc *SSIClient) CreateDID(controller string) (*DIDDocument, *ecdsa.PrivateKey, error) {	// Generate a new key pair
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	
	// Convert public key to hex
	pubKeyBytes := elliptic.Marshal(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)
	pubKeyHex := hex.EncodeToString(pubKeyBytes)
	
	// Create DID
	digest := sha256.Sum256(pubKeyBytes)
	keyID := fmt.Sprintf("%x", digest[:8])
	did := fmt.Sprintf("did:%s:%s", sc.DIDMethod, keyID)
	
	// Create verification method ID
	verificationMethodID := fmt.Sprintf("%s#keys-1", did)
	
	// Create DID Document
	now := time.Now()
	didDocument := &DIDDocument{
		ID: did,
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
		},
		Controller: controller,
		VerificationMethod: []VerificationMethod{
			{
				ID:           verificationMethodID,
				Type:         "EcdsaSecp256r1VerificationKey2019",
				Controller:   did,
				PublicKeyHex: pubKeyHex,
			},
		},
		Authentication: []string{verificationMethodID},
		AssertionMethod: []string{verificationMethodID},
		Service: []ServiceEndpoint{
			{
				ID:              fmt.Sprintf("%s#service-1", did),
				Type:            "TracePostServiceEndpoint",
				ServiceEndpoint: "https://api.tracepost.vn/verifiable-credentials",
				Description:     "TracePost verification service",
			},
		},
		Created: now,
		Updated: now,
	}
	
	// Create proof
	proof, err := sc.createDIDDocumentProof(didDocument, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create DID document proof: %w", err)
	}
	
	didDocument.Proof = proof
	
	// Add to registry
	sc.DIDRegistry[did] = didDocument
	
	return didDocument, privateKey, nil
}

// ResolveDID resolves a DID to a DID document
func (sc *SSIClient) ResolveDID(did string) (*DIDDocument, error) {
	// Check if DID is in registry
	if document, exists := sc.DIDRegistry[did]; exists {
		return document, nil
	}
	
	// Check if DID is in format "did:{method}:{id}"
	parts := strings.SplitN(did, ":", 3)
	if len(parts) != 3 || parts[0] != "did" {
		return nil, fmt.Errorf("invalid DID format: %s", did)
	}
	
	// Check if method is supported
	if parts[1] != sc.DIDMethod {
		return nil, fmt.Errorf("unsupported DID method: %s", parts[1])
	}
	
	// In a real implementation, this would query a blockchain or decentralized registry
	// For this implementation, if it's not in our local registry, we can't resolve it
	return nil, fmt.Errorf("DID not found: %s", did)
}

// createDIDDocumentProof creates a proof for a DID document
func (sc *SSIClient) createDIDDocumentProof(document *DIDDocument, privateKey *ecdsa.PrivateKey) (*DIDDocumentProof, error) {
	// Create a canonical representation of the document without the proof
	documentWithoutProof := *document
	documentWithoutProof.Proof = nil
	
	documentBytes, err := json.Marshal(documentWithoutProof)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DID document: %w", err)
	}
	
	// Hash the canonical representation
	hash := sha256.Sum256(documentBytes)
	
	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}
	
	// Combine r and s to create the signature
	signature := append(r.Bytes(), s.Bytes()...)
	
	// Encode the signature as base64
	proofValue := base64.StdEncoding.EncodeToString(signature)
	
	// Create the proof
	verificationMethodID := fmt.Sprintf("%s#keys-1", document.ID)
	proof := &DIDDocumentProof{
		Type:               "EcdsaSecp256r1Signature2019",
		Created:            time.Now(),
		VerificationMethod: verificationMethodID,
		ProofPurpose:       "assertionMethod",
		ProofValue:         proofValue,
	}
	
	return proof, nil
}

// VerifyDIDDocument verifies a DID document
func (sc *SSIClient) VerifyDIDDocument(document *DIDDocument) (bool, error) {
	if document.Proof == nil {
		return false, errors.New("DID document has no proof")
	}
	
	// Find the verification method
	var verificationMethod *VerificationMethod
	for _, vm := range document.VerificationMethod {
		if vm.ID == document.Proof.VerificationMethod {
			verificationMethod = &vm
			break
		}
	}
	
	if verificationMethod == nil {
		return false, fmt.Errorf("verification method not found: %s", document.Proof.VerificationMethod)
	}
	
	// Get the public key
	if verificationMethod.PublicKeyHex == "" {
		return false, errors.New("verification method has no public key")
	}
	
	// Decode the public key
	pubKeyBytes, err := hex.DecodeString(verificationMethod.PublicKeyHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode public key: %w", err)
	}
	
	// Parse the public key
	x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
	if x == nil {
		return false, errors.New("failed to unmarshal public key")
	}
	
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	
	// Create a canonical representation of the document without the proof
	documentWithoutProof := *document
	documentWithoutProof.Proof = nil
	
	documentBytes, err := json.Marshal(documentWithoutProof)
	if err != nil {
		return false, fmt.Errorf("failed to marshal DID document: %w", err)
	}
	
	// Hash the canonical representation
	hash := sha256.Sum256(documentBytes)
	
	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(document.Proof.ProofValue)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}
	
	// Split the signature into r and s
	if len(signature) != 64 {
		return false, errors.New("invalid signature length")
	}
	
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])
	
	// Verify the signature
	return ecdsa.Verify(publicKey, hash[:], r, s), nil
}

// IssueVerifiableCredential issues a verifiable credential
func (sc *SSIClient) IssueVerifiableCredential(
	issuerDID string,
	issuerPrivateKey *ecdsa.PrivateKey,
	subjectDID string,
	credentialType string,
	claims map[string]interface{},
	expirationDays int,
) (*VerifiableCredential, error) {
	// Check if issuer DID exists
	_, err := sc.ResolveDID(issuerDID)
	if err != nil {
		return nil, fmt.Errorf("issuer DID not found: %w", err)
	}
	
	// Check if subject DID exists
	_, err = sc.ResolveDID(subjectDID)
	if err != nil {
		return nil, fmt.Errorf("subject DID not found: %w", err)
	}
	
	// Create credential ID
	credentialID := fmt.Sprintf("urn:uuid:%s", generateRandomID())
	
	// Set issuance and expiration dates
	now := time.Now()
	expirationDate := now.AddDate(0, 0, expirationDays)
	
	// Create credential subject
	credentialSubject := map[string]interface{}{
		"id": subjectDID,
	}
	for k, v := range claims {
		credentialSubject[k] = v
	}
	
	// Create credential
	credential := &VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1",
		},
		ID:     credentialID,
		Type:   []string{"VerifiableCredential", credentialType},
		Issuer: issuerDID,
		IssuanceDate: now,
		ExpirationDate: expirationDate,
		CredentialSubject: credentialSubject,
		Status: CredentialStatus{
			ID:   fmt.Sprintf("https://api.tracepost.vn/credentials/status/%s", credentialID),
			Type: "CredentialStatusList2017",
		},
	}
	
	// Create proof
	verificationMethodID := fmt.Sprintf("%s#keys-1", issuerDID)
	proof, err := sc.createCredentialProof(credential, issuerPrivateKey, verificationMethodID)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential proof: %w", err)
	}
	
	credential.Proof = proof
	
	// Add to registry
	sc.CredentialRegistry[credentialID] = credential
	
	return credential, nil
}

// createCredentialProof creates a proof for a verifiable credential
func (sc *SSIClient) createCredentialProof(
	credential *VerifiableCredential,
	privateKey *ecdsa.PrivateKey,
	verificationMethodID string,
) (*CredentialProof, error) {
	// Create a canonical representation of the credential without the proof
	credentialWithoutProof := *credential
	credentialWithoutProof.Proof = nil
	
	credentialBytes, err := json.Marshal(credentialWithoutProof)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal credential: %w", err)
	}
	
	// Hash the canonical representation
	hash := sha256.Sum256(credentialBytes)
	
	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}
	
	// Combine r and s to create the signature
	signature := append(r.Bytes(), s.Bytes()...)
	
	// Encode the signature as base64
	proofValue := base64.StdEncoding.EncodeToString(signature)
	
	// Create challenge
	challenge := generateRandomID()
	
	// Create the proof
	proof := &CredentialProof{
		Type:               "EcdsaSecp256r1Signature2019",
		Created:            time.Now(),
		VerificationMethod: verificationMethodID,
		ProofPurpose:       "assertionMethod",
		ProofValue:         proofValue,
		Challenge:          challenge,
		Domain:             "tracepost.vn",
	}
	
	return proof, nil
}

// VerifyCredential verifies a verifiable credential
func (sc *SSIClient) VerifyCredential(credential *VerifiableCredential) (bool, error) {
	if credential.Proof == nil {
		return false, errors.New("credential has no proof")
	}
	
	// Check if issuer DID exists
	issuerDoc, err := sc.ResolveDID(credential.Issuer)
	if err != nil {
		return false, fmt.Errorf("issuer DID not found: %w", err)
	}
	
	// Find the verification method
	var verificationMethod *VerificationMethod
	for _, vm := range issuerDoc.VerificationMethod {
		if vm.ID == credential.Proof.VerificationMethod {
			verificationMethod = &vm
			break
		}
	}
	
	if verificationMethod == nil {
		return false, fmt.Errorf("verification method not found: %s", credential.Proof.VerificationMethod)
	}
	
	// Get the public key
	if verificationMethod.PublicKeyHex == "" {
		return false, errors.New("verification method has no public key")
	}
	
	// Decode the public key
	pubKeyBytes, err := hex.DecodeString(verificationMethod.PublicKeyHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode public key: %w", err)
	}
	
	// Parse the public key
	x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
	if x == nil {
		return false, errors.New("failed to unmarshal public key")
	}
	
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	
	// Create a canonical representation of the credential without the proof
	credentialWithoutProof := *credential
	credentialWithoutProof.Proof = nil
	
	credentialBytes, err := json.Marshal(credentialWithoutProof)
	if err != nil {
		return false, fmt.Errorf("failed to marshal credential: %w", err)
	}
	
	// Hash the canonical representation
	hash := sha256.Sum256(credentialBytes)
	
	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(credential.Proof.ProofValue)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}
	
	// Split the signature into r and s
	if len(signature) != 64 {
		return false, errors.New("invalid signature length")
	}
	
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])
	
	// Check expiration
	if !credential.ExpirationDate.IsZero() && time.Now().After(credential.ExpirationDate) {
		return false, errors.New("credential has expired")
	}
	
	// Verify the signature
	return ecdsa.Verify(publicKey, hash[:], r, s), nil
}

// CreateVerifiablePresentation creates a verifiable presentation
func (sc *SSIClient) CreateVerifiablePresentation(
	holderDID string,
	holderPrivateKey *ecdsa.PrivateKey,
	credentials []*VerifiableCredential,
	challenge string,
	domain string,
) (*VerifiablePresentation, error) {
	// Check if holder DID exists
	_, err := sc.ResolveDID(holderDID)
	if err != nil {
		return nil, fmt.Errorf("holder DID not found: %w", err)
	}
	
	// Verify each credential
	for _, credential := range credentials {
		valid, err := sc.VerifyCredential(credential)
		if err != nil {
			return nil, fmt.Errorf("failed to verify credential: %w", err)
		}
		if !valid {
			return nil, errors.New("credential verification failed")
		}
		
		// Check if the credential belongs to the holder
		if subjectID, ok := credential.CredentialSubject["id"].(string); ok {
			if subjectID != holderDID {
				return nil, errors.New("credential does not belong to holder")
			}
		}
	}
	
	// Create presentation ID
	presentationID := fmt.Sprintf("urn:uuid:%s", generateRandomID())
	
	// Create presentation
	presentation := &VerifiablePresentation{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1",
		},
		ID:     presentationID,
		Type:   []string{"VerifiablePresentation"},
		Holder: holderDID,
		VerifiableCredential: credentials,
	}
	
	// Create proof
	verificationMethodID := fmt.Sprintf("%s#keys-1", holderDID)
	proof, err := sc.createPresentationProof(presentation, holderPrivateKey, verificationMethodID, challenge, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to create presentation proof: %w", err)
	}
	
	presentation.Proof = proof
	
	// Add to registry
	sc.PresentationRegistry[presentationID] = presentation
	
	return presentation, nil
}

// createPresentationProof creates a proof for a verifiable presentation
func (sc *SSIClient) createPresentationProof(
	presentation *VerifiablePresentation,
	privateKey *ecdsa.PrivateKey,
	verificationMethodID string,
	challenge string,
	domain string,
) (*PresentationProof, error) {
	// Create a canonical representation of the presentation without the proof
	presentationWithoutProof := *presentation
	presentationWithoutProof.Proof = nil
	
	presentationBytes, err := json.Marshal(presentationWithoutProof)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal presentation: %w", err)
	}
	
	// Hash the canonical representation
	hash := sha256.Sum256(presentationBytes)
	
	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign hash: %w", err)
	}
	
	// Combine r and s to create the signature
	signature := append(r.Bytes(), s.Bytes()...)
	
	// Encode the signature as base64
	proofValue := base64.StdEncoding.EncodeToString(signature)
	
	// Create the proof
	proof := &PresentationProof{
		Type:               "EcdsaSecp256r1Signature2019",
		Created:            time.Now(),
		VerificationMethod: verificationMethodID,
		ProofPurpose:       "authentication",
		ProofValue:         proofValue,
		Challenge:          challenge,
		Domain:             domain,
	}
	
	return proof, nil
}

// VerifyPresentation verifies a verifiable presentation
func (sc *SSIClient) VerifyPresentation(presentation *VerifiablePresentation, challenge string, domain string) (bool, error) {
	if presentation.Proof == nil {
		return false, errors.New("presentation has no proof")
	}
	
	// Check if challenge matches
	if presentation.Proof.Challenge != challenge {
		return false, errors.New("challenge does not match")
	}
	
	// Check if domain matches
	if presentation.Proof.Domain != domain {
		return false, errors.New("domain does not match")
	}
	
	// Check if holder DID exists
	holderDocument, err := sc.ResolveDID(presentation.Holder)
	if err != nil {
		return false, fmt.Errorf("holder DID not found: %w", err)
	}
	
	// Find the verification method
	var verificationMethod *VerificationMethod
	for _, vm := range holderDocument.VerificationMethod {
		if vm.ID == presentation.Proof.VerificationMethod {
			verificationMethod = &vm
			break
		}
	}
	
	if verificationMethod == nil {
		return false, fmt.Errorf("verification method not found: %s", presentation.Proof.VerificationMethod)
	}
	
	// Get the public key
	if verificationMethod.PublicKeyHex == "" {
		return false, errors.New("verification method has no public key")
	}
	
	// Decode the public key
	pubKeyBytes, err := hex.DecodeString(verificationMethod.PublicKeyHex)
	if err != nil {
		return false, fmt.Errorf("failed to decode public key: %w", err)
	}
	
	// Parse the public key
	x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
	if x == nil {
		return false, errors.New("failed to unmarshal public key")
	}
	
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	
	// Create a canonical representation of the presentation without the proof
	presentationWithoutProof := *presentation
	presentationWithoutProof.Proof = nil
	
	presentationBytes, err := json.Marshal(presentationWithoutProof)
	if err != nil {
		return false, fmt.Errorf("failed to marshal presentation: %w", err)
	}
	
	// Hash the canonical representation
	hash := sha256.Sum256(presentationBytes)
	
	// Decode the signature
	signature, err := base64.StdEncoding.DecodeString(presentation.Proof.ProofValue)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}
	
	// Split the signature into r and s
	if len(signature) != 64 {
		return false, errors.New("invalid signature length")
	}	
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])
	
	// Verify each credential
	for _, credential := range presentation.VerifiableCredential {
		valid, err := sc.VerifyCredential(credential)
		if err != nil {
			return false, fmt.Errorf("failed to verify credential: %w", err)
		}
		if !valid {
			return false, errors.New("credential verification failed")
		}
		
		// Check if the credential belongs to the holder
		if subjectID, ok := credential.CredentialSubject["id"].(string); ok {
			if subjectID != presentation.Holder {
				return false, errors.New("credential does not belong to holder")
			}
		}
	}
	
	// Verify the signature
	return ecdsa.Verify(publicKey, hash[:], r, s), nil
}

// RegisterTrustedIssuer registers a trusted issuer
func (sc *SSIClient) RegisterTrustedIssuer(issuerDID string) {
	sc.TrustedIssuers[issuerDID] = true
}

// IsTrustedIssuer checks if an issuer is trusted
func (sc *SSIClient) IsTrustedIssuer(issuerDID string) bool {
	return sc.TrustedIssuers[issuerDID]
}

// CreateGDPRCompliantCredential creates a GDPR-compliant verifiable credential
func (sc *SSIClient) CreateGDPRCompliantCredential(
	issuerDID string,
	issuerPrivateKey *ecdsa.PrivateKey,
	subjectDID string,
	credentialType string,
	claims map[string]interface{},
	expirationDays int,
	dataProcessingConsent string,
) (*VerifiableCredential, error) {
	// Add GDPR-specific claims
	gdprClaims := map[string]interface{}{
		"dataProcessingConsent": dataProcessingConsent,
		"dataController":        issuerDID,
		"dataSubject":           subjectDID,
		"purposeOfProcessing":   "Supply chain traceability",
		"dataRetentionPeriod":   fmt.Sprintf("%d days", expirationDays),
		"rightToBeRemoved":      "https://api.tracepost.vn/data-removal",
		"gdprCompliant":         true,
	}
	
	// Merge GDPR claims with provided claims
	for k, v := range gdprClaims {
		claims[k] = v
	}
	
	return sc.IssueVerifiableCredential(issuerDID, issuerPrivateKey, subjectDID, credentialType, claims, expirationDays)
}

// RevokeCredential revokes a credential
func (sc *SSIClient) RevokeCredential(credentialID string, issuerDID string, issuerPrivateKey *ecdsa.PrivateKey) error {
	// Check if credential exists
	credential, exists := sc.CredentialRegistry[credentialID]
	if !exists {
		return fmt.Errorf("credential not found: %s", credentialID)
	}
	
	// Check if issuer matches
	if credential.Issuer != issuerDID {
		return errors.New("only the issuer can revoke a credential")
	}
	
	// In a real implementation, this would update a revocation registry on a blockchain
	// For this implementation, we'll just remove it from our registry
	delete(sc.CredentialRegistry, credentialID)
	
	return nil
}

// Helper function to generate a random ID
func generateRandomID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
