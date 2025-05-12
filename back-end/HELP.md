# HELP - Core API Flows

This document outlines the main sequences of API calls (core flows) for common use cases.

---

## 1. Authentication Flow

1. **Register**
   - POST `/api/v1/auth/register`
2. **Login**
   - POST `/api/v1/auth/login`
   - Retrieve JWT access token and refresh token
3. **Refresh Token** (optional)
   - POST `/api/v1/auth/refresh`
4. **Logout**
   - POST `/api/v1/auth/logout`

---

## 2. Company & User Management

1. **Create Company** (admin)
   - POST `/api/v1/companies`
2. **List Companies**
   - GET `/api/v1/companies`
3. **Get Company Details**
   - GET `/api/v1/companies/:companyId`
4. **Update Company** (admin)
   - PUT `/api/v1/companies/:companyId`
5. **Delete Company** (admin)
   - DELETE `/api/v1/companies/:companyId`

**User Lifecycle** (admin)

1. POST `/api/v1/users` (create)
2. GET `/api/v1/users` (list)
3. GET `/api/v1/users/:userId` (detail)
4. PUT `/api/v1/users/:userId` (update)
5. DELETE `/api/v1/users/:userId` (remove)
6. **Profile**
   - GET `/api/v1/users/me`
   - PUT `/api/v1/users/me`
   - PUT `/api/v1/users/me/password`

---

## 3. Hatchery & Batch Management

1. **Create Hatchery** (admin or hatchery_manager)
   - POST `/api/v1/hatcheries`
2. **List Hatcheries**
   - GET `/api/v1/hatcheries`
3. **Get Hatchery**
   - GET `/api/v1/hatcheries/:hatcheryId`
4. **Update Hatchery**
   - PUT `/api/v1/hatcheries/:hatcheryId`
5. **Delete Hatchery** (admin)
   - DELETE `/api/v1/hatcheries/:hatcheryId`
6. **Batch Operations**
   a. **Create Batch** (DDI-protected)

   - POST `/api/v1/batches`  
     b. **List Batches**
   - GET `/api/v1/batches`  
     c. **Get Batch**
   - GET `/api/v1/batches/:batchId`  
     d. **Update Batch Status** (DDI)
   - PUT `/api/v1/batches/:batchId/status`

7. **Batch Data**
   - GET `/api/v1/batches/:batchId/events`
   - GET `/api/v1/batches/:batchId/documents`
   - GET `/api/v1/batches/:batchId/environment`
   - GET `/api/v1/batches/:batchId/history`

---

## 4. Shipment Transfers

1. **List Transfers**
   - GET `/api/v1/shipments/transfers`
2. **Get Transfer**
   - GET `/api/v1/shipments/transfers/:id`
3. **Get Transfers by Batch**
   - GET `/api/v1/shipments/transfers/batch/:batchId`
4. **Create Transfer** (DDI)
   - POST `/api/v1/shipments/transfers`
5. **Update Transfer** (DDI)
   - PUT `/api/v1/shipments/transfers/:id`
6. **Delete Transfer** (DDI)
   - DELETE `/api/v1/shipments/transfers/:id`

---

## 5. Supply Chain & Traceability

1. **Supply Chain Details**
   - GET `/api/v1/supplychain/:batchId`
2. **Supply Chain QR Code**
   - GET `/api/v1/supplychain/:batchId/qr`
3. **Trace via QR** (public)
   - GET `/api/v1/qr/:batchId`
4. **Gateway QR**
   - GET `/api/v1/qr/gateway/:batchId`

---

## 6. Events, Documents & Environment

1. **Record Event** (DDI)
   - POST `/api/v1/events`
2. **Upload Document** (DDI)
   - POST `/api/v1/documents`
3. **Get Document**
   - GET `/api/v1/documents/:documentId`
4. **Record Environment Data** (DDI)
   - POST `/api/v1/environment`

---

## 7. Mobile Endpoints

1. **Mobile Trace by QR**
   - GET `/api/v1/mobile/trace/:qrCode`
2. **Mobile Batch Summary**
   - GET `/api/v1/mobile/batch/:batchId/summary`

---

## 8. Blockchain & Interoperability

1. **Query On-Chain Batch**
   - GET `/api/v1/blockchain/batch/:batchId`
2. **Query On-Chain Event**
   - GET `/api/v1/blockchain/event/:eventId`
3. **Interoperability Operations** (admin)
   - Register Chains: POST `/api/v1/interop/chains`
   - Share Batch: POST `/api/v1/interop/share-batch`
   - Export Batch: GET `/api/v1/interop/export/:batchId`
   - List Chains: GET `/api/v1/interop/chains`
   - Get TX: GET `/api/v1/interop/txs/:txId`

**Cosmos & Polkadot**

- Create Cosmos Bridge: POST `/api/v1/interop/bridges/cosmos`
- Send IBC Packet: POST `/api/v1/interop/ibc/send`
- Create Polkadot Bridge: POST `/api/v1/interop/bridges/polkadot`
- Send XCM Message: POST `/api/v1/interop/xcm/send`

---

## 9. Blockchain-as-a-Service (BaaS)

1. Create Network: POST `/api/v1/baas/networks`
2. List Networks: GET `/api/v1/baas/networks`
3. Get Network: GET `/api/v1/baas/networks/:networkId`
4. Update Network: PUT `/api/v1/baas/networks/:networkId`
5. Delete Network: DELETE `/api/v1/baas/networks/:networkId`
6. Deploy Contract: POST `/api/v1/baas/deployments`
7. List Deployments: GET `/api/v1/baas/deployments`
8. Get Deployment: GET `/api/v1/baas/deployments/:deploymentId`

---

## 10. Decentralized Identity (DDI)

**Public DID**

1. Create DID: POST `/api/v1/identity/did`
2. Resolve DID: GET `/api/v1/identity/did/:did`
3. Verify DID Proof: POST `/api/v1/identity/verify`

**Protected DID** 4. Claim Operations (JWT)

- Create Claim: POST `/api/v1/identity/claim`
- Get Claim: GET `/api/v1/identity/claim/:claimId`
- Verify Claim: POST `/api/v1/identity/claim/verify`
- Revoke Claim: PUT `/api/v1/identity/claim/:claimId/revoke`

5. Permission Operations
   - Update: PUT `/api/v1/identity/permissions`
   - Verify: POST `/api/v1/identity/permissions/verify`

---

## 11. Compliance & Geospatial Tracking

1. Check Compliance: GET `/api/v1/compliance/check/:batchId`
2. Generate Report: GET `/api/v1/compliance/report/:batchId`
3. List Standards: GET `/api/v1/compliance/standards`
4. Validate Standard: POST `/api/v1/compliance/validate`

**Geo Location** 5. Record Location: POST `/api/v1/geo/location` 6. Get Journey: GET `/api/v1/geo/batch/:batchId/journey` 7. Current Location: GET `/api/v1/geo/batch/:batchId/current-location`

---

For detailed request/response schemas, refer to the Swagger UI at `/swagger`.
