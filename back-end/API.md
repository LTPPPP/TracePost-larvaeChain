# API Documentation

## Overview

This document provides a comprehensive overview of the APIs available in the blockchain logistics traceability system. It includes the main features, API endpoints, and the sequential flow of API usage for different roles. This will help in designing use cases and workflows for each role.

---

## Features

1. **Authentication and Authorization**
   - User login, password management, and role-based access control.
2. **User Management**
   - Manage user accounts, roles, and compliance.
3. **Decentralized Identity**
   - Issue and revoke decentralized identities (DIDs).
4. **Blockchain Integration**
   - Configure and monitor blockchain nodes and transactions.
5. **Hatchery Management**
   - Manage hatchery information, batches, and environmental monitoring.
6. **Distributor Operations**
   - Real-time batch tracking and supply chain event logging.
7. **Traceability and Certification**
   - QR code generation, NFT certification, and traceability.

---

## API Endpoints

### Authentication

1. **Login**

   - Endpoint: `/api/v1/auth/login`
   - Method: POST
   - Description: Authenticate a user and retrieve a token.
   - JSON Body Example:
     ```json
     {
       "username": "john_doe",
       "password": "securePassword123"
     }
     ```

2. **Change Password**

   - Endpoint: `/api/v1/users/me/password`
   - Method: PUT
   - Description: Change the current user's password.
   - JSON Body Example:
     ```json
     {
       "oldPassword": "securePassword123",
       "newPassword": "newSecurePassword456"
     }
     ```

3. **Forgot Password**
   - Endpoint: `/api/v1/auth/forgot-password`
   - Method: POST
   - Description: Request a password reset link.
   - JSON Body Example:
     ```json
     {
       "email": "john.doe@example.com"
     }
     ```

### User Management

1. **Lock/Unlock User Account**

   - Endpoint: `/api/v1/users/:userId/lock`
   - Method: PUT
   - Description: Lock or unlock a user account.
   - JSON Body Example:
     ```json
     {
       "action": "lock"
     }
     ```

2. **List Users by Role**

   - Endpoint: `/api/v1/users`
   - Method: GET
   - Description: Retrieve a list of users filtered by role.
   - Query Parameters:
     - `role`: The role to filter users by (e.g., `hatchery`, `distributor`).

3. **Approve Hatchery Account**

   - Endpoint: `/api/v1/hatcheries/:hatcheryId/approve`
   - Method: PUT
   - Description: Approve a hatchery account.

4. **Revoke Certificate**
   - Endpoint: `/api/v1/certificates/:certificateId/revoke`
   - Method: PUT
   - Description: Revoke a certificate for violations.

### Decentralized Identity

1. **Issue DID**

   - Endpoint: `/api/v1/did/issue`
   - Method: POST
   - Description: Issue a decentralized identity (DID) to an entity.
   - JSON Body Example:
     ```json
     {
       "entityId": "hatchery_123",
       "type": "hatchery"
     }
     ```

2. **Revoke DID**
   - Endpoint: `/api/v1/did/revoke`
   - Method: POST
   - Description: Revoke a compromised DID.
   - JSON Body Example:
     ```json
     {
       "did": "did:example:123456789abcdefghi"
     }
     ```

### Blockchain Integration

1. **Configure Blockchain Node**

   - Endpoint: `/api/v1/blockchain/configure`
   - Method: POST
   - Description: Configure a blockchain node.
   - JSON Body Example:
     ```json
     {
       "nodeUrl": "http://blockchain-node.example.com",
       "network": "mainnet"
     }
     ```

2. **Monitor Transactions**
   - Endpoint: `/api/v1/blockchain/monitor`
   - Method: GET
   - Description: Monitor multi-chain transactions.

### Hatchery Management

1. **View Hatchery Information**

   - Endpoint: `/api/v1/hatcheries/:hatcheryId`
   - Method: GET
   - Description: Retrieve information about a hatchery.

2. **Update Hatchery Information**

   - Endpoint: `/api/v1/hatcheries/:hatcheryId`
   - Method: PUT
   - Description: Update hatchery information.
   - JSON Body Example:
     ```json
     {
       "name": "Green Hatchery",
       "address": "123 Hatchery Lane",
       "certificates": ["cert_001", "cert_002"]
     }
     ```

3. **Create Batch**

   - Endpoint: `/api/v1/batches`
   - Method: POST
   - Description: Create a new shrimp batch.
   - JSON Body Example:
     ```json
     {
       "batchId": "batch_001",
       "geneticInfo": "GMO-Free",
       "hatcheryId": "hatchery_123"
     }
     ```

4. **Log Event**

   - Endpoint: `/api/v1/events`
   - Method: POST
   - Description: Log a hatchery event with DID authentication.
   - JSON Body Example:
     ```json
     {
       "batchId": "batch_001",
       "eventType": "feeding",
       "timestamp": "2025-05-20T10:00:00Z"
     }
     ```

5. **Generate QR Code**
   - Endpoint: `/api/v1/batches/:batchId/qr`
   - Method: GET
   - Description: Generate a QR code for a batch.

### Distributor Operations

1. **Scan QR Code**

   - Endpoint: `/api/v1/qr/:code`
   - Method: GET
   - Description: Scan a QR code to retrieve batch information.

2. **Log Supply Chain Event**
   - Endpoint: `/api/v1/events`
   - Method: POST
   - Description: Log a supply chain event.
   - JSON Body Example:
     ```json
     {
       "batchId": "batch_001",
       "eventType": "transport",
       "temperature": "5C",
       "timestamp": "2025-05-20T12:00:00Z"
     }
     ```

---

## API Flow by User Role

### Admin

1. **Login** (`/api/v1/auth/login`)
2. **List Users by Role** (`/api/v1/users`)
3. **Approve Hatchery Account** (`/api/v1/hatcheries/:hatcheryId/approve`)
4. **Revoke Certificate** (`/api/v1/certificates/:certificateId/revoke`)
5. **Issue DID** (`/api/v1/did/issue`)
6. **Logout** (`/api/v1/auth/logout`)

### Hatchery

1. **Login** (`/api/v1/auth/login`)
2. **View Hatchery Information** (`/api/v1/hatcheries/:hatcheryId`)
3. **Update Hatchery Information** (`/api/v1/hatcheries/:hatcheryId`)
4. **Create Batch** (`/api/v1/batches`)
5. **Log Event** (`/api/v1/events`)
6. **Generate QR Code** (`/api/v1/batches/:batchId/qr`)
7. **Logout** (`/api/v1/auth/logout`)

### Distributor

1. **Login** (`/api/v1/auth/login`)
2. **Scan QR Code** (`/api/v1/qr/:code`)
3. **Log Supply Chain Event** (`/api/v1/events`)
4. **Logout** (`/api/v1/auth/logout`)

### User

1. **Login** (`/api/v1/auth/login`)
2. **Log Event** (`/api/v1/events`)
3. **Trace Batch History** (`/api/v1/qr/:code`)
4. **Logout** (`/api/v1/auth/logout`)

---

## Notes

- Ensure proper authentication before accessing any API.
- Follow the sequential flow for each role to maintain data integrity.
- Use appropriate error handling for API responses.
