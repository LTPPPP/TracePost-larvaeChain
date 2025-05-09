# Decentralized Digital Identity (DDI) System

This document provides information on how to use the DDI verification system for identity and access control in the TracePost-larvaeChain application.

## Overview

The DDI system uses blockchain-based decentralized identifiers to verify the identity of users and control access to sensitive operations in the supply chain. This ensures that only authorized parties can write or retrieve critical data.

## Features

- Decentralized identity creation and management
- Cryptographic proof-based authentication
- Fine-grained permission control
- DDI middleware for protecting API endpoints
- Command-line tools for DDI management

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
```

Example using curl:

```bash
curl -X POST https://api.example.com/api/v1/batches \
  -H "X-DID: did:tracepost:hatchery:1234567890abcdef" \
  -H "X-DID-Proof: abcdefghijklmnopqrstuvwxyz1234567890" \
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

## Permissions

DDI permissions control what operations an entity can perform:

- `create_batch`: Allows creating new batches
- `update_batch_status`: Allows updating batch status
- `record_event`: Allows recording supply chain events
- `record_environment`: Allows recording environmental data
- `upload_document`: Allows uploading documents
- `create_shipment`: Allows creating shipment transfers
- `update_shipment`: Allows updating shipment transfers
- `delete_shipment`: Allows deleting shipment transfers

## For Developers

To implement DDI authentication in your client application:

1. Store the DID and private key securely
2. Generate a new proof for each API request
3. Include the DID and proof in the HTTP headers
4. Handle authentication errors appropriately

See the `blockchain/ddi_client.go` file for a reference implementation of a DDI client.
