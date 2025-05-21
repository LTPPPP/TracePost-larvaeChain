# API Note

This document lists all available API endpoints and sample request bodies.

## Authentication

- **POST** `/api/v1/auth/login`

  Sample JSON body:

  ```json
  {
    "username": "john_doe",
    "password": "securePassword123"
  }
  ```

- **PUT** `/api/v1/users/me/password`

  Sample JSON body:

  ```json
  {
    "oldPassword": "securePassword123",
    "newPassword": "newSecurePassword456"
  }
  ```

- **POST** `/api/v1/auth/forgot-password`

  Sample JSON body:

  ```json
  {
    "email": "john.doe@example.com"
  }
  ```

## User Management

- **PUT** `/api/v1/users/:userId/lock`

  Sample JSON body:

  ```json
  {
    "action": "lock"
  }
  ```

- **GET** `/api/v1/users`

  (Query parameter: `role`)

- **PUT** `/api/v1/hatcheries/:hatcheryId/approve`

  (No request body)

- **PUT** `/api/v1/certificates/:certificateId/revoke`

  (No request body)

## Decentralized Identity

- **POST** `/api/v1/did/issue`

  Sample JSON body:

  ```json
  {
    "entityId": "hatchery_123",
    "type": "hatchery"
  }
  ```

- **POST** `/api/v1/did/revoke`

  Sample JSON body:

  ```json
  {
    "did": "did:example:123456789abcdefghi"
  }
  ```

## Blockchain Integration

- **POST** `/api/v1/blockchain/configure`

  Sample JSON body:

  ```json
  {
    "nodeUrl": "http://blockchain-node.example.com",
    "network": "mainnet"
  }
  ```

- **GET** `/api/v1/blockchain/monitor`

  (No request body)

## Hatchery Management

- **GET** `/api/v1/hatcheries/:hatcheryId`

  (No request body)

- **PUT** `/api/v1/hatcheries/:hatcheryId`

  Sample JSON body:

  ```json
  {
    "name": "Green Hatchery",
    "address": "123 Hatchery Lane",
    "certificates": ["cert_001", "cert_002"]
  }
  ```

- **POST** `/api/v1/batches`

  Sample JSON body:

  ```json
  {
    "batchId": "batch_001",
    "geneticInfo": "GMO-Free",
    "hatcheryId": "hatchery_123"
  }
  ```

- **POST** `/api/v1/events`

  Sample JSON body:

  ```json
  {
    "batchId": "batch_001",
    "eventType": "feeding",
    "timestamp": "2025-05-20T10:00:00Z"
  }
  ```

- **GET** `/api/v1/batches/:batchId/qr`

  (No request body)

## Distributor Operations

- **GET** `/api/v1/qr/:code`

  (No request body)

- **POST** `/api/v1/events`

  Sample JSON body:

  ```json
  {
    "batchId": "batch_001",
    "eventType": "transport",
    "temperature": "5C",
    "timestamp": "2025-05-20T12:00:00Z"
  }
  ```

## Full API Endpoints (Chi tiết tất cả các API)

### 1. User Management

- GET /users
- GET /users/{userId}
- POST /users
- PUT /users/{userId}
- DELETE /users/{userId}

### 2. Health Check

- GET /health

### 3. Mobile Client

- GET /mobile/trace/{qrCode}
- GET /mobile/batch/{batchId}/summary

### 4. Interoperability

- GET /interop/chains
- GET /interop/txs/{txId}

### 5. Alliance

- POST /alliance/share
- GET /alliance/members
- POST /alliance/join

### 6. Shipments

- GET /shipments/transfers
- GET /shipments/transfers/{id}
- GET /shipments/transfers/batch/{batchId}
- POST /shipments/transfers
- PUT /shipments/transfers/{id}
- DELETE /shipments/transfers/{id}
- GET /shipments/transfers/{id}/qr

### 7. Scaling

- POST /scaling/sharding/configure

### 8. Admin Operations

- PUT /admin/users/{userId}/status
- GET /admin/users
- PUT /admin/hatcheries/{hatcheryId}/approve
- PUT /admin/certificates/{docId}/revoke
- POST /admin/compliance/check
- POST /admin/compliance/export
- POST /admin/identity/issue
- POST /admin/identity/revoke
- POST /admin/blockchain/nodes/configure
- GET /admin/blockchain/monitor

### 9. Admin Analytics

- GET /admin/analytics/dashboard
- GET /admin/analytics/system
- GET /admin/analytics/blockchain
- GET /admin/analytics/compliance
- GET /admin/analytics/users
- GET /admin/analytics/batches
- GET /admin/analytics/export
- POST /admin/analytics/refresh

### 10. NFT & Tokenization

- POST /nft/contracts
- POST /nft/batches/tokenize
- GET /nft/batches/{batchId}
- GET /nft/tokens/{tokenId}
- PUT /nft/tokens/{tokenId}/transfer
- POST /nft/transactions/tokenize
- GET /nft/transactions/{transferId}
- GET /nft/transactions/{transferId}/trace
- GET /nft/transactions/{transferId}/qr
