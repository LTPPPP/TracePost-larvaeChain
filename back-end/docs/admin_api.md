# Admin API Documentation

The Admin API provides powerful management features for system administrators. These endpoints are restricted to users with the `admin` role and enable comprehensive control over users, hatcheries, compliance, and blockchain infrastructure.

## User Management

### Lock/Unlock User Account

`PUT /api/v1/admin/users/{userId}/status`

Enables admins to activate or deactivate user accounts.

**Request Body:**

```json
{
  "is_active": true,
  "reason": "User account restored after identity verification"
}
```

**Response:**

```json
{
  "success": true,
  "message": "User account successfully unlocked",
  "data": {
    "userId": 123,
    "status": true,
    "reason": "User account restored after identity verification",
    "updated": "2025-05-19T10:30:45Z"
  }
}
```

### Get Users By Role

`GET /api/v1/admin/users?role=hatchery_manager`

Retrieves users filtered by their assigned role.

**Parameters:**

- `role` (query, optional): Filter users by role

**Response:**

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 123,
      "username": "farm_manager1",
      "email": "manager@example.com",
      "full_name": "Farm Manager",
      "role": "hatchery_manager",
      "company_id": 45,
      "is_active": true,
      "last_login": "2025-05-18T14:22:30Z"
    },
    {
      "id": 124,
      "username": "farm_manager2",
      "email": "manager2@example.com",
      "full_name": "Farm Manager 2",
      "role": "hatchery_manager",
      "company_id": 46,
      "is_active": true,
      "last_login": "2025-05-17T09:45:12Z"
    }
  ]
}
```

## Hatchery Management

### Approve Hatchery Registration

`PUT /api/v1/admin/hatcheries/{hatcheryId}/approve`

Approves or rejects a hatchery account registration.

**Request Body:**

```json
{
  "is_approved": true,
  "comment": "All verification requirements met"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Hatchery account successfully approved",
  "data": {
    "hatcheryId": 45,
    "status": true,
    "comment": "All verification requirements met",
    "updated": "2025-05-19T11:20:35Z"
  }
}
```

## Compliance Management

### Revoke Certificate

`PUT /api/v1/admin/certificates/{docId}/revoke`

Revokes a compliance certificate when violations are found.

**Request Body:**

```json
{
  "reason": "Environmental standards violation detected during spot inspection"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Certificate successfully revoked",
  "data": {
    "documentId": 789,
    "reason": "Environmental standards violation detected during spot inspection",
    "revokedAt": "2025-05-19T13:45:22Z",
    "transaction": "0x123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
  }
}
```

### Check Standard Compliance

`POST /api/v1/admin/compliance/check`

Automatically checks batch compliance against FDA/ASC standards.

**Request Body:**

```json
{
  "batch_id": 123,
  "standards": ["FDA", "ASC"]
}
```

**Response:**

```json
{
  "success": true,
  "message": "Compliance check completed",
  "data": {
    "batch_id": 123,
    "standards": ["FDA", "ASC"],
    "compliance": {
      "FDA": true,
      "ASC": false
    },
    "details": {
      "FDA": [],
      "ASC": ["Density exceeds ASC recommended level of 300 PL/m³"]
    },
    "parameters": {
      "temperature": 28.5,
      "pH": 7.8,
      "salinity": 12.5,
      "density": 350
    },
    "generated_at": "2025-05-19T14:30:22Z"
  }
}
```

### Export Compliance Report

`POST /api/v1/admin/compliance/export`

Generates and exports compliance reports in multiple formats.

**Request Body:**

```json
{
  "batch_id": 123,
  "format": "gs1_epcis"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Compliance report generated in gs1_epcis format",
  "data": {
    "batch_id": 123,
    "format": "gs1_epcis",
    "report_url": "/api/v1/admin/reports/gs1/123",
    "generated_at": "2025-05-19T15:10:45Z",
    "details": {
      "format": "GS1 EPCIS XML",
      "version": "1.2",
      "standard": "GS1",
      "epcisEvents": 12
    }
  }
}
```

## Decentralized Identity

### Issue DID

`POST /api/v1/admin/identity/issue`

Issues a decentralized identifier (DID) for an entity.

**Request Body:**

```json
{
  "entity_type": "hatchery",
  "entity_id": 45,
  "claims": {
    "certification": "organic",
    "verification_level": "premium",
    "region": "central_vietnam"
  }
}
```

**Response:**

```json
{
  "success": true,
  "message": "DID issued successfully",
  "data": {
    "did": "did:tracepost:hatchery:45",
    "entity_type": "hatchery",
    "entity_id": 45,
    "issued_at": "2025-05-19T16:20:15Z",
    "claims": {
      "certification": "organic",
      "verification_level": "premium",
      "region": "central_vietnam"
    }
  }
}
```

### Revoke DID

`POST /api/v1/admin/identity/revoke`

Revokes a compromised decentralized identifier.

**Request Body:**

```json
{
  "did": "did:tracepost:hatchery:45",
  "reason": "Security breach detected on entity's systems"
}
```

**Response:**

```json
{
  "success": true,
  "message": "DID revoked successfully",
  "data": {
    "did": "did:tracepost:hatchery:45",
    "reason": "Security breach detected on entity's systems",
    "revoked_at": "2025-05-19T17:05:30Z",
    "status": "revoked"
  }
}
```

## Blockchain Integration

### Configure Blockchain Node

`POST /api/v1/admin/blockchain/nodes/configure`

Configures a blockchain node in the network.

**Request Body:**

```json
{
  "network_id": "tracepost-main",
  "node_name": "validator-central-1",
  "node_type": "validator",
  "endpoint": "https://node1.blockchain.tracepost.com",
  "parameters": {
    "max_peers": "50",
    "consensus_timeout": "5000ms",
    "batch_sync": "true"
  },
  "is_validator": true,
  "is_active": true
}
```

**Response:**

```json
{
  "success": true,
  "message": "Blockchain node configured successfully",
  "data": {
    "network_id": "tracepost-main",
    "node_name": "validator-central-1",
    "node_type": "validator",
    "endpoint": "https://node1.blockchain.tracepost.com",
    "is_validator": true,
    "is_active": true,
    "configured_at": "2025-05-19T18:15:40Z"
  }
}
```

### Monitor Blockchain Transactions

`GET /api/v1/admin/blockchain/monitor`

Monitors transactions across multiple blockchain networks.

**Response:**

```json
{
  "success": true,
  "message": "Cross-chain transactions retrieved successfully",
  "data": {
    "transactions": [
      {
        "chain_id": "tracepost-main",
        "tx_hash": "0x123456789abcdef",
        "status": "confirmed",
        "block_number": 12345,
        "timestamp": "2025-05-19T17:15:22Z",
        "sender": "0xabcdef123456789",
        "receiver": "0x987654321fedcba",
        "value": "0.05 ETH",
        "gas_used": 21000
      },
      {
        "chain_id": "cosmos-ibc",
        "tx_hash": "ABCDEF1234567890",
        "status": "pending",
        "timestamp": "2025-05-19T17:45:10Z",
        "sender": "cosmos1abcdefg",
        "receiver": "cosmos1hijklmn",
        "value": "100 ATOM"
      }
    ],
    "fetched_at": "2025-05-19T18:30:00Z",
    "chain_count": 2
  }
}
```
