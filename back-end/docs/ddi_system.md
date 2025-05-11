# Decentralized Digital Identity (DDI) System

This document provides information on how to use the DDI verification system for identity and access control in the TracePost-larvaeChain application.

## Overview

The DDI system implements W3C Decentralized Identifiers (DIDs) and Verifiable Credentials standards to verify the identity of users and control access to sensitive operations in the supply chain. This ensures that only authorized parties can write or retrieve critical data.

## Architecture

The DDI system consists of the following components:

1. **W3C DID Implementation**

   - Supports the `did:tracepost` method
   - Conforms to W3C DID Core specification
   - Provides cryptographic proof of identity

2. **Verifiable Credentials**

   - Issues and verifies credentials for supply chain participants
   - Supports credential revocation and expiration
   - Enables attribute-based access control

3. **Smart Contract Integration**

   - On-chain verification of permissions
   - Transaction validation based on identity and permissions
   - Credential registry

4. **API Layer**
   - Middleware for DID authentication
   - Permission verification for protected operations
   - RESTful endpoints for DID operations

## DID Method: did:tracepost

### Format

```
did:tracepost:<entity-type>:<identifier>
```

- `<entity-type>`: Type of entity (hatchery, farm, processor, etc.)
- `<identifier>`: Unique identifier derived from the public key

### DID Document Structure

```json
{
  "@context": ["https://www.w3.org/ns/did/v1"],
  "id": "did:tracepost:hatchery:123456789abcdef",
  "controller": ["did:tracepost:hatchery:123456789abcdef"],
  "verificationMethod": [
    {
      "id": "did:tracepost:hatchery:123456789abcdef#keys-1",
      "type": "EcdsaSecp256k1VerificationKey2019",
      "controller": "did:tracepost:hatchery:123456789abcdef",
      "publicKeyJwk": {
        "kty": "EC",
        "crv": "secp256k1",
        "x": "...",
        "y": "..."
      }
    }
  ],
  "authentication": ["did:tracepost:hatchery:123456789abcdef#keys-1"],
  "assertionMethod": ["did:tracepost:hatchery:123456789abcdef#keys-1"],
  "service": [
    {
      "id": "did:tracepost:hatchery:123456789abcdef#profile",
      "type": "TracePostProfile",
      "serviceEndpoint": "https://tracepost.example/api/profile/123456789abcdef"
    }
  ],
  "created": "2023-05-12T12:00:00Z",
  "updated": "2023-05-12T12:00:00Z"
}
```

## Permissions System

The permissions system uses a combination of on-chain verification and smart contracts to manage access control.

### Permission Structure

```json
{
  "action": "create_batch",
  "resource": "batch",
  "conditions": ["time < 2024-12-31T23:59:59Z"],
  "expiry": 1735689599
}
```

### Available Actions

- `create_batch` - Create a new batch
- `update_batch_status` - Update batch status
- `transfer_batch` - Transfer a batch to another entity
- `view_batch` - View batch information
- `create_shipment` - Create a new shipment
- `update_shipment` - Update shipment information
- `record_event` - Record a supply chain event
- `analyze_data` - Analyze supply chain data
- `issue_credential` - Issue a verifiable credential

## How to Use

### Creating a DDI

To create a new decentralized digital identity, use the command-line tool:

```bash
go run cmd/ddi-tool/main.go generate --type hatchery --name "My Hatchery"
```

This will generate a new DID and save the private key to a file. Keep this file secure and never share it.

### Generating a Proof for Authentication

To generate a proof for API authentication:

```bash
go run cmd/ddi-tool/main.go proof --did did:tracepost:hatchery:1234567890abcdef --key your_did_key_file.key
```

This will output a proof that can be used for API authentication.

### Using DDI Authentication in API Requests

To authenticate API requests using DDI, include the following HTTP headers:

```
X-DID: your-did
X-DID-Proof: your-generated-proof
X-DID-Timestamp: 2023-05-12T12:00:00Z
```

Example using curl:

```bash
curl -X POST https://api.example.com/api/v1/batches \
  -H "X-DID: did:tracepost:hatchery:1234567890abcdef" \
  -H "X-DID-Proof: abcdefghijklmnopqrstuvwxyz1234567890" \
  -H "X-DID-Timestamp: 2023-05-12T12:00:00Z" \
  -H "Content-Type: application/json" \
  -d '{"batch_name": "Batch 001", "species": "Litopenaeus vannamei", "quantity": 100000}'
```

### Verifying a DDI Proof

To verify a DID proof:

```bash
go run cmd/ddi-tool/main.go verify --did did:tracepost:hatchery:1234567890abcdef --proof your-proof
```

This will check if the proof is valid and display the permissions associated with the DID.

## Protected Operations

The following operations are protected by DDI authentication:

- Creating and updating batches
- Recording supply chain events
- Uploading documents
- Recording environmental data
- Managing shipment transfers
- Analyzing supply chain data

## Integration with Blockchain

The DDI system integrates with blockchain through:

1. **Identity Registry Contract**

   - Stores DIDs and their current status
   - Maps DIDs to public keys
   - Provides resolution functionality

2. **Permission Registry Contract**

   - Stores permissions for DIDs
   - Validates permission claims
   - Handles permission delegation

3. **Access Validator Contract**
   - Validates transaction requests
   - Checks permissions before allowing state changes
   - Prevents unauthorized operations

## API Endpoints

### DID Management

- `POST /api/v1/identity/did` - Create a new DID
- `GET /api/v1/identity/did/:did` - Resolve a DID
- `PUT /api/v1/identity/did/:did` - Update a DID
- `DELETE /api/v1/identity/did/:did` - Deactivate a DID

### Credential Management

- `POST /api/v1/identity/credentials` - Issue a credential
- `GET /api/v1/identity/credentials/:id` - Get a credential
- `POST /api/v1/identity/credentials/verify` - Verify a credential
- `PUT /api/v1/identity/credentials/:id/revoke` - Revoke a credential

### Permission Management

- `POST /api/v1/identity/permissions` - Grant a permission
- `DELETE /api/v1/identity/permissions/:id` - Revoke a permission
- `GET /api/v1/identity/permissions/:did` - Get permissions for a DID

## Using the DDI Client

```go
// Initialize DDI client
config := blockchain.DDIClientConfig{
    PrivateKeyPEM: privateKey,
    DID:           "did:tracepost:hatchery:123456789abcdef",
    ContractAddress: "0x1234567890abcdef1234567890abcdef12345678",
}

blockchainClient := blockchain.NewBlockchainClient(
    "https://blockchain.example",
    privateKey,
    "0xMyAccount",
    "1",
    "poa",
)

ddiClient, err := blockchain.NewDDIClient(config, blockchainClient)
if err != nil {
    log.Fatalf("Failed to create DDI client: %v", err)
}

// Generate proof for authentication
proof, err := ddiClient.GenerateProof()
if err != nil {
    log.Fatalf("Failed to generate proof: %v", err)
}

// Check permission
hasPermission, err := ddiClient.CheckPermission("create_batch", "batch")
if err != nil {
    log.Fatalf("Failed to check permission: %v", err)
}

if !hasPermission {
    log.Fatal("Permission denied")
}

// Verify transaction
isValid, err := ddiClient.VerifyTransaction("create_batch", "batch", batchData)
if err != nil {
    log.Fatalf("Failed to verify transaction: %v", err)
}

if !isValid {
    log.Fatal("Transaction verification failed")
}

// Create a verifiable credential
credential, err := ddiClient.CreateVerifiableCredential(
    "did:tracepost:farm:987654321fedcba",
    map[string]interface{}{
        "role": "Certified Farm",
        "certificationLevel": "Gold",
    },
    365, // Valid for 1 year
)
if err != nil {
    log.Fatalf("Failed to create credential: %v", err)
}
```

## Security Considerations

1. **Key Management**

   - Private keys must be securely stored
   - Consider using HSMs for key protection
   - Implement key rotation procedures

2. **Authentication**

   - Use timestamped messages to prevent replay attacks
   - Verify DIDs before accepting authentication
   - Implement rate limiting to prevent brute force attacks

3. **Authorization**

   - Always verify permissions for protected operations
   - Implement least privilege principle
   - Audit all permission changes

4. **Smart Contract Security**
   - Ensure contract functions are properly protected
   - Implement access control using roles
   - Consider formal verification of critical contracts

## For Developers

To implement DDI authentication in your client application:

1. Store the DID and private key securely
2. Generate a new proof for each API request
3. Include the DID and proof in the HTTP headers
4. Handle authentication errors appropriately

See the `blockchain/ddi_client.go` file for a reference implementation of a DDI client.

## Best Practices

1. **DID Management**

   - Create different DIDs for different contexts
   - Use separate key pairs for different operations
   - Keep DID documents up to date

2. **Credential Usage**

   - Only issue credentials with appropriate verification
   - Set reasonable expiration times
   - Include minimal necessary information in credentials

3. **Integration**
   - Use middleware for consistent authentication
   - Verify permissions before performing actions
   - Log all authentication and authorization decisions

## Conclusion

The DDI system provides a secure, decentralized identity framework for supply chain participants. By leveraging W3C standards and blockchain technology, it ensures that identity and access management is both secure and privacy-preserving, while enabling fine-grained control over supply chain operations.
