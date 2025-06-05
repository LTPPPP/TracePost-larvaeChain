# 🌐 TracePost-LarvaeChain API Documentation

[![API Version](https://img.shields.io/badge/API-v1.0-blue?style=flat&logo=swagger)](https://api.tracepost.com/docs)
[![OpenAPI](https://img.shields.io/badge/OpenAPI-3.0-green?style=flat&logo=openapi-initiative)](https://spec.openapis.org/oas/v3.0.3)
[![REST](https://img.shields.io/badge/REST-API-orange?style=flat&logo=rest)](https://restfulapi.net/)
[![GraphQL](https://img.shields.io/badge/GraphQL-E10098?style=flat&logo=graphql)](https://graphql.org/)

## 📋 Table of Contents

- [Overview](#-overview)
- [System Architecture](#-system-architecture)
- [API Workflow](#-api-workflow)
- [Authentication](#-authentication)
- [Core API Services](#-core-api-services)
- [API Reference](#-api-reference)
- [Error Handling](#-error-handling)
- [Rate Limiting](#-rate-limiting)
- [Webhooks](#-webhooks)
- [SDK & Libraries](#-sdk--libraries)
- [Testing](#-testing)
- [Best Practices](#-best-practices)

## 🎯 Overview

The **TracePost-LarvaeChain API** is a comprehensive RESTful and GraphQL API suite designed for enterprise-grade blockchain-based supply chain traceability. This documentation provides detailed specifications for integrating with our aquaculture traceability platform, enabling seamless tracking from hatchery to consumer.

### Mission Statement

To provide a robust, scalable, and standards-compliant API ecosystem that enables complete transparency and traceability in the global shrimp larvae supply chain through blockchain technology.

### Key Features

- **🔗 Blockchain Integration**: Direct interaction with custom Cosmos SDK blockchain
- **🌐 Multi-Chain Support**: Cross-chain interoperability with Ethereum, Polkadot
- **🔐 Advanced Security**: Zero-knowledge proofs and decentralized identity
- **📊 Real-time Analytics**: Live supply chain monitoring and insights
- **🏆 Compliance Ready**: Built-in support for international standards
- **⚡ High Performance**: 10K+ TPS with sub-second response times

### API Principles

- **RESTful Design**: Follows REST architectural principles
- **OpenAPI 3.0 Compliant**: Complete specification available
- **Idempotent Operations**: Safe retry mechanisms
- **Stateless Architecture**: Scalable and reliable
- **Version Controlled**: Backward compatibility guaranteed

## 🏗️ System Architecture

### API Gateway Architecture

```mermaid
graph TB
    subgraph "Client Applications"
        WEB[Web Frontend]
        MOBILE[Mobile App]
        IOT[IoT Devices]
        EXTERNAL[External Systems]
    end
    
    subgraph "API Gateway Layer"
        LB[Load Balancer]
        AUTH[Authentication Service]
        RATE[Rate Limiter]
        VALID[Request Validator]
    end
    
    subgraph "Core API Services"
        BAAS[BaaS API]
        IDENTITY[Identity API]
        BATCH[Batch API]
        NFT[NFT API]
        INTEROP[Interoperability API]
        ANALYTICS[Analytics API]
    end
    
    subgraph "Data & Blockchain Layer"
        POSTGRES[(PostgreSQL)]
        REDIS[(Redis Cache)]
        IPFS[(IPFS Storage)]
        COSMOS[Cosmos Blockchain]
        ETH[Ethereum Bridge]
        POLKADOT[Polkadot Bridge]
    end
    
    WEB --> LB
    MOBILE --> LB
    IOT --> LB
    EXTERNAL --> LB
    
    LB --> AUTH
    AUTH --> RATE
    RATE --> VALID
    
    VALID --> BAAS
    VALID --> IDENTITY
    VALID --> BATCH
    VALID --> NFT
    VALID --> INTEROP
    VALID --> ANALYTICS
    
    BAAS --> COSMOS
    IDENTITY --> POSTGRES
    BATCH --> POSTGRES
    BATCH --> IPFS
    NFT --> COSMOS
    INTEROP --> ETH
    INTEROP --> POLKADOT
    ANALYTICS --> REDIS
```

### Microservices Overview

| Service | Responsibility | Dependencies | Performance |
|---------|----------------|--------------|-------------|
| **BaaS API** | Blockchain infrastructure management | Cosmos SDK, Tendermint | 5K+ TPS |
| **Identity API** | Decentralized identity & credentials | PostgreSQL, W3C DID | 1K+ ops/sec |
| **Batch API** | Supply chain batch management | PostgreSQL, IPFS | 10K+ ops/sec |
| **NFT API** | Digital asset tokenization | Blockchain, IPFS | 1K+ mint/sec |
| **Interoperability API** | Cross-chain communication | IBC, Bridges | 500+ tx/sec |
| **Analytics API** | Data processing & insights | Redis, PostgreSQL | Real-time |

## 🔄 API Workflow

### Sequential System Flow

The TracePost-LarvaeChain API follows a carefully orchestrated workflow that ensures data integrity, security, and traceability throughout the supply chain process.

#### 1. **Infrastructure Setup (BaaS API)**

**Purpose**: Establishes the foundational blockchain infrastructure

```mermaid
sequenceDiagram
    participant Client
    participant BaaS
    participant Blockchain
    participant Consensus
    
    Client->>BaaS: POST /baas/networks
    BaaS->>Blockchain: Initialize Network
    Blockchain->>Consensus: Setup Validators
    Consensus-->>Blockchain: Network Ready
    Blockchain-->>BaaS: Network ID
    BaaS-->>Client: Network Configuration
```

**Key Operations:**
- Network creation and configuration
- Validator setup and consensus mechanism
- Smart contract deployment
- Cross-chain bridge initialization

**Endpoints:**
```http
POST   /baas/networks              # Create blockchain network
GET    /baas/networks              # List all networks
GET    /baas/networks/{networkId}  # Get network details
PUT    /baas/networks/{networkId}  # Update network config
DELETE /baas/networks/{networkId}  # Decommission network
```

#### 2. **Identity Management (Identity API)**

**Purpose**: Establishes trusted identities for all supply chain participants

```mermaid
sequenceDiagram
    participant Entity
    participant Identity
    participant DIDRegistry
    participant Blockchain
    
    Entity->>Identity: POST /identity/did
    Identity->>DIDRegistry: Create DID Document
    DIDRegistry->>Blockchain: Register DID
    Blockchain-->>DIDRegistry: DID Confirmed
    DIDRegistry-->>Identity: DID Created
    Identity-->>Entity: DID & Credentials
```

**Key Operations:**
- DID creation and resolution
- Verifiable credential issuance
- Identity verification and validation
- Access control management

**Sample DID Creation:**
```json
POST /identity/did
{
  "entity_type": "hatchery",
  "entity_name": "Premium Aquaculture Hatchery",
  "metadata": {
    "location": {
      "country": "Vietnam",
      "region": "Mekong Delta",
      "coordinates": {
        "latitude": 10.0452,
        "longitude": 105.7469
      }
    },
    "certifications": [
      {
        "type": "ASC",
        "id": "ASC-VN-2024-001",
        "issuer": "Aquaculture Stewardship Council",
        "issued_date": "2024-01-15T00:00:00Z",
        "expires_date": "2027-01-15T00:00:00Z"
      },
      {
        "type": "BAP",
        "id": "BAP-VN-2024-002",
        "issuer": "Best Aquaculture Practices",
        "issued_date": "2024-02-01T00:00:00Z",
        "expires_date": "2027-02-01T00:00:00Z"
      }
    ],
    "capacity": {
      "annual_production": 50000000,
      "species_specialization": ["Litopenaeus vannamei", "Penaeus monodon"]
    }
  }
}
```

#### 3. **Batch Lifecycle Management (Batch API)**

**Purpose**: Tracks complete lifecycle of shrimp larvae batches

```mermaid
sequenceDiagram
    participant Hatchery
    participant Batch
    participant IPFS
    participant Blockchain
    participant QR
    
    Hatchery->>Batch: POST /batches
    Batch->>IPFS: Store Metadata
    IPFS-->>Batch: IPFS Hash
    Batch->>Blockchain: Record Batch
    Blockchain-->>Batch: TX Hash
    Batch->>QR: Generate QR Code
    QR-->>Batch: QR Code Data
    Batch-->>Hatchery: Batch Created
```

**Key Operations:**
- Batch creation and registration
- Environmental monitoring integration
- Quality scoring and assessment
- Supply chain event tracking
- QR code generation and management

**Sample Batch Creation:**
```json
POST /batches
{
  "hatchery_id": "did:tracepost:hatchery:premium-aqua-001",
  "species": "Litopenaeus vannamei",
  "strain": "Pacific White Shrimp - SPF",
  "quantity": 2000000,
  "production_date": "2024-03-15T06:00:00Z",
  "genetic_lineage": {
    "broodstock_id": "BS-2024-001",
    "generation": "F3",
    "genetic_markers": ["WSSV_resistant", "high_growth"]
  },
  "environmental_conditions": {
    "temperature": {
      "value": 28.5,
      "unit": "celsius",
      "tolerance": "±1°C"
    },
    "ph": {
      "value": 8.1,
      "tolerance": "±0.2"
    },
    "salinity": {
      "value": 35.0,
      "unit": "ppt",
      "tolerance": "±2 ppt"
    },
    "dissolved_oxygen": {
      "value": 6.8,
      "unit": "mg/L",
      "tolerance": "≥6.0 mg/L"
    }
  },
  "quality_metrics": {
    "survival_rate": 95.5,
    "growth_rate": "high",
    "disease_resistance": "excellent",
    "stress_tolerance": "high"
  },
  "certifications": [
    "ASC-VN-2024-001",
    "BAP-VN-2024-002",
    "Organic-VN-2024-003"
  ]
}
```

#### 4. **Cross-Chain Interoperability (Interoperability API)**

**Purpose**: Enables data sharing across multiple blockchain networks

```mermaid
sequenceDiagram
    participant Source
    participant Interop
    participant IBC
    participant Target
    participant Validator
    
    Source->>Interop: POST /interoperability/batches/share
    Interop->>IBC: Create Packet
    IBC->>Target: Send Packet
    Target->>Validator: Validate Data
    Validator-->>Target: Validation Result
    Target-->>IBC: Acknowledge
    IBC-->>Interop: Success
    Interop-->>Source: Shared Successfully
```

**Key Operations:**
- Cross-chain batch data sharing
- Bridge management and monitoring
- Protocol compliance verification
- Data integrity validation

**Sample Cross-Chain Sharing:**
```json
POST /interoperability/batches/share
{
  "batch_id": "BATCH-2024-03-001",
  "target_chains": [
    {
      "chain_id": "ethereum-mainnet",
      "bridge_type": "IBC",
      "target_contract": "0x1234567890AbCdEf1234567890AbCdEf12345678"
    },
    {
      "chain_id": "polkadot-kusama",
      "bridge_type": "XCM",
      "target_parachain": 2000
    }
  ],
  "data_fields": [
    "batch_id",
    "hatchery_id",
    "species",
    "quality_score",
    "certifications"
  ],
  "privacy_level": "public",
  "expiry_duration": "30d"
}
```

#### 5. **NFT Tokenization (NFT API)**

**Purpose**: Converts physical assets into tradeable digital tokens

```mermaid
sequenceDiagram
    participant Owner
    participant NFT
    participant Contract
    participant Blockchain
    participant Metadata
    
    Owner->>NFT: POST /nft/batches/tokenize
    NFT->>Contract: Deploy Contract
    Contract-->>NFT: Contract Address
    NFT->>Metadata: Store Metadata
    Metadata-->>NFT: Metadata URI
    NFT->>Blockchain: Mint NFT
    Blockchain-->>NFT: Token ID
    NFT-->>Owner: NFT Created
```

**Key Operations:**
- Smart contract deployment
- Batch tokenization and minting
- Ownership transfer management
- Metadata storage and retrieval

**Sample NFT Creation:**
```json
POST /nft/batches/tokenize
{
  "batch_id": "BATCH-2024-03-001",
  "network_id": "cosmos-tracepost-1",
  "recipient_address": "cosmos1abc...def",
  "token_standard": "ERC-721",
  "metadata": {
    "name": "Premium Shrimp Larvae Batch #001",
    "description": "High-quality SPF Pacific White Shrimp larvae from certified hatchery",
    "image": "ipfs://QmX...abc",
    "attributes": [
      {
        "trait_type": "Species",
        "value": "Litopenaeus vannamei"
      },
      {
        "trait_type": "Quality Score",
        "value": 98.5
      },
      {
        "trait_type": "Certification",
        "value": "ASC Certified"
      },
      {
        "trait_type": "Origin",
        "value": "Mekong Delta, Vietnam"
      }
    ],
    "properties": {
      "quantity": 2000000,
      "production_date": "2024-03-15T06:00:00Z",
      "hatchery": "Premium Aquaculture Hatchery",
      "blockchain_verified": true
    }
  }
}
```

## 🔐 Authentication

### Authentication Methods

#### 1. **JWT Bearer Token Authentication**

**Primary method for API access**

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Token Structure:**
```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user_12345",
    "iss": "tracepost-api",
    "aud": "tracepost-clients",
    "exp": 1640995200,
    "iat": 1640908800,
    "role": "hatchery_operator",
    "permissions": [
      "batches:read",
      "batches:write",
      "identity:read"
    ],
    "entity_id": "did:tracepost:hatchery:premium-aqua-001"
  }
}
```

#### 2. **API Key Authentication**

**For system-to-system integration**

```http
X-API-Key: tp_live_1234567890abcdef
X-API-Secret: sk_1234567890abcdef1234567890abcdef
```

#### 3. **OAuth 2.0 / OpenID Connect**

**For third-party integrations**

```http
Authorization: Bearer oauth_access_token_here
```

### Authentication Endpoints

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "operator@premium-aqua.com",
  "password": "SecurePassword123!",
  "mfa_code": "123456"
}
```

**Response:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "rt_1234567890abcdef...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "scope": "read write admin",
  "user": {
    "id": "user_12345",
    "email": "operator@premium-aqua.com",
    "role": "hatchery_operator",
    "entity_id": "did:tracepost:hatchery:premium-aqua-001",
    "permissions": ["batches:read", "batches:write"]
  }
}
```

#### Token Refresh
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "rt_1234567890abcdef..."
}
```

### Role-Based Access Control (RBAC)

| Role | Permissions | Description |
|------|-------------|-------------|
| **System Admin** | `*:*` | Full system access |
| **Hatchery Operator** | `batches:*, events:*, identity:read` | Manage batches and events |
| **Distributor** | `batches:read, shipments:*, events:read` | Track and manage shipments |
| **Auditor** | `*:read, compliance:*` | Read-only access with compliance tools |
| **Consumer** | `batches:read, trace:read` | View traceability information |

## 🛠️ Core API Services

### 1. **Blockchain-as-a-Service (BaaS) API**

**Base URL:** `/baas`

**Purpose:** Provides foundational blockchain infrastructure management capabilities.

#### Network Management

##### Create Blockchain Network
```http
POST /baas/networks
Content-Type: application/json

{
  "name": "TracePost Production Network",
  "type": "cosmos-sdk",
  "consensus": {
    "type": "tendermint",
    "block_time": "6s",
    "validators": 21
  },
  "features": [
    "smart_contracts",
    "ibc_enabled",
    "nft_support"
  ],
  "scaling": {
    "type": "horizontal",
    "max_validators": 100,
    "sharding_enabled": true
  }
}
```

##### Get Network Status
```http
GET /baas/networks/{networkId}/status
```

**Response:**
```json
{
  "network_id": "net-prod-20240315",
  "status": "active",
  "block_height": 2845672,
  "total_validators": 21,
  "active_validators": 21,
  "transactions_per_second": 8947,
  "average_block_time": "5.8s",
  "consensus_health": "excellent",
  "network_metrics": {
    "total_transactions": 15847623,
    "total_blocks": 2845672,
    "chain_size_gb": 156.7,
    "peer_count": 89
  }
}
```

### 2. **Identity Management API**

**Base URL:** `/identity`

**Purpose:** Manages decentralized identities and verifiable credentials.

#### DID Operations

##### Create DID
```http
POST /identity/did
Content-Type: application/json

{
  "entity_type": "hatchery",
  "entity_name": "EcoShrimp Hatchery Ltd",
  "metadata": {
    "registration": {
      "business_license": "BL-VN-2024-001",
      "tax_id": "VN123456789",
      "incorporation_date": "2010-05-15"
    },
    "location": {
      "address": "123 Aquaculture Road, Can Tho, Vietnam",
      "coordinates": {
        "latitude": 10.0452,
        "longitude": 105.7469
      }
    },
    "certifications": [
      {
        "type": "ASC",
        "number": "ASC-VN-2024-001",
        "issuer": "Aquaculture Stewardship Council",
        "valid_from": "2024-01-01",
        "valid_until": "2027-01-01"
      }
    ]
  }
}
```

##### Issue Verifiable Credential
```http
POST /identity/credentials/issue
Content-Type: application/json

{
  "issuer_did": "did:tracepost:authority:certification-body",
  "subject_did": "did:tracepost:hatchery:ecoshrimp-001",
  "credential_type": "QualityCertification",
  "claims": {
    "certification_name": "Premium Quality Assurance",
    "certification_level": "Grade A",
    "quality_score": 98.5,
    "audit_date": "2024-03-01",
    "valid_until": "2025-03-01",
    "standards_compliance": [
      "ISO 22000",
      "HACCP",
      "ASC"
    ]
  },
  "evidence": [
    {
      "type": "DocumentEvidence",
      "document_type": "audit_report",
      "ipfs_hash": "QmX1Y2Z3...",
      "verification_method": "digital_signature"
    }
  ]
}
```

### 3. **Batch Management API**

**Base URL:** `/batches`

**Purpose:** Comprehensive batch lifecycle management and tracking.

#### Batch Lifecycle

##### Create Batch
```http
POST /batches
Content-Type: application/json

{
  "hatchery_id": "did:tracepost:hatchery:ecoshrimp-001",
  "species": "Litopenaeus vannamei",
  "strain": "SPF Pacific White",
  "quantity": 5000000,
  "production_date": "2024-03-15T06:00:00Z",
  "expected_harvest": "2024-09-15T06:00:00Z",
  "genetic_info": {
    "broodstock_source": "Hawaii SPF Line",
    "generation": "F4",
    "genetic_markers": [
      "WSSV_resistance",
      "high_growth_rate",
      "disease_tolerance"
    ]
  },
  "environmental_setup": {
    "tank_system": "RAS",
    "water_source": "filtered_seawater",
    "initial_conditions": {
      "temperature": 28.0,
      "salinity": 35.0,
      "ph": 8.0,
      "dissolved_oxygen": 7.2
    }
  },
  "quality_targets": {
    "survival_rate_target": 95.0,
    "growth_rate_target": "high",
    "uniformity_target": 90.0
  }
}
```

##### Add Environmental Data
```http
POST /batches/{batchId}/environmental-data
Content-Type: application/json

{
  "timestamp": "2024-03-16T10:30:00Z",
  "measurements": {
    "temperature": {
      "value": 28.2,
      "unit": "celsius",
      "sensor_id": "temp_001",
      "status": "normal"
    },
    "ph": {
      "value": 8.1,
      "sensor_id": "ph_001",
      "status": "optimal"
    },
    "dissolved_oxygen": {
      "value": 7.0,
      "unit": "mg/L",
      "sensor_id": "do_001",
      "status": "good"
    },
    "ammonia": {
      "value": 0.02,
      "unit": "mg/L",
      "sensor_id": "nh3_001",
      "status": "safe"
    }
  },
  "automated": true,
  "iot_device_id": "iot_monitor_001"
}
```

### 4. **NFT Management API**

**Base URL:** `/nft`

**Purpose:** Digital asset tokenization and NFT lifecycle management.

#### Contract Management

##### Deploy NFT Contract
```http
POST /nft/contracts
Content-Type: application/json

{
  "contract_name": "TracePostBatchNFT",
  "contract_symbol": "TPBN",
  "network_id": "cosmos-tracepost-1",
  "contract_type": "ERC-721",
  "features": [
    "enumerable",
    "metadata",
    "royalties",
    "batch_linked"
  ],
  "royalty_info": {
    "recipient": "cosmos1royalty...",
    "percentage": 2.5
  },
  "metadata_base_uri": "https://metadata.tracepost.com/nft/"
}
```

##### Tokenize Batch
```http
POST /nft/batches/tokenize
Content-Type: application/json

{
  "batch_id": "BATCH-2024-03-001",
  "contract_address": "cosmos1contract...",
  "recipient_address": "cosmos1recipient...",
  "metadata": {
    "name": "Premium Shrimp Larvae Batch #001",
    "description": "Certified organic shrimp larvae from EcoShrimp Hatchery",
    "image": "ipfs://QmImageHash...",
    "external_url": "https://tracepost.com/batch/BATCH-2024-03-001",
    "attributes": [
      {
        "trait_type": "Species",
        "value": "Litopenaeus vannamei"
      },
      {
        "trait_type": "Quality Grade",
        "value": "Premium A+"
      },
      {
        "trait_type": "Quantity",
        "value": 5000000,
        "display_type": "number"
      },
      {
        "trait_type": "Certification",
        "value": "ASC Certified"
      }
    ]
  }
}
```

### 5. **Analytics API**

**Base URL:** `/analytics`

**Purpose:** Advanced data analytics and business intelligence.

#### Supply Chain Analytics

##### Get Batch Analytics
```http
GET /analytics/batches/{batchId}/insights
```

**Response:**
```json
{
  "batch_id": "BATCH-2024-03-001",
  "analytics_summary": {
    "quality_score": 98.5,
    "performance_grade": "A+",
    "risk_assessment": "low",
    "compliance_status": "fully_compliant"
  },
  "environmental_analysis": {
    "stability_score": 96.2,
    "optimal_conditions_percentage": 94.8,
    "critical_events": 0,
    "trend_analysis": {
      "temperature": "stable",
      "ph": "improving",
      "oxygen": "optimal"
    }
  },
  "predictive_insights": {
    "estimated_survival_rate": 96.2,
    "harvest_quality_prediction": "excellent",
    "market_readiness_date": "2024-09-10T00:00:00Z",
    "yield_prediction": {
      "quantity": 4810000,
      "confidence": 92.3
    }
  },
  "benchmark_comparison": {
    "industry_average_quality": 85.2,
    "hatchery_average_quality": 94.1,
    "performance_ranking": "top_5_percent"
  }
}
```

## 📖 API Reference

### Standard Response Format

All API responses follow a consistent structure:

```json
{
  "success": true,
  "data": {
    // Response payload
  },
  "metadata": {
    "timestamp": "2024-03-15T10:30:00Z",
    "version": "v1.0",
    "request_id": "req_1234567890"
  },
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 156,
    "total_pages": 8
  }
}
```

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "BATCH_NOT_FOUND",
    "message": "The specified batch could not be found",
    "details": {
      "batch_id": "BATCH-2024-03-999",
      "suggestion": "Verify the batch ID and try again"
    }
  },
  "metadata": {
    "timestamp": "2024-03-15T10:30:00Z",
    "request_id": "req_1234567890"
  }
}
```

### Common HTTP Status Codes

| Status Code | Meaning | Usage |
|-------------|---------|-------|
| `200 OK` | Success | Successful GET, PUT requests |
| `201 Created` | Resource Created | Successful POST requests |
| `202 Accepted` | Request Accepted | Async operations started |
| `204 No Content` | Success, No Body | Successful DELETE requests |
| `400 Bad Request` | Invalid Request | Malformed request data |
| `401 Unauthorized` | Authentication Required | Missing or invalid auth |
| `403 Forbidden` | Access Denied | Insufficient permissions |
| `404 Not Found` | Resource Not Found | Resource doesn't exist |
| `409 Conflict` | Resource Conflict | Duplicate or conflicting data |
| `422 Unprocessable Entity` | Validation Error | Invalid data format |
| `429 Too Many Requests` | Rate Limit Exceeded | Too many requests |
| `500 Internal Server Error` | Server Error | Unexpected server error |

## ⚠️ Error Handling

### Error Categories

#### 1. **Validation Errors (4xx)**

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "field_errors": [
        {
          "field": "quantity",
          "error": "must be greater than 0",
          "provided_value": -1000
        },
        {
          "field": "species",
          "error": "must be a valid species name",
          "provided_value": "invalid_species"
        }
      ]
    }
  }
}
```

#### 2. **Business Logic Errors (4xx)**

```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS",
    "message": "User does not have permission to perform this action",
    "details": {
      "required_permission": "batches:write",
      "user_permissions": ["batches:read"]
    }
  }
}
```

#### 3. **System Errors (5xx)**

```json
{
  "success": false,
  "error": {
    "code": "BLOCKCHAIN_UNAVAILABLE",
    "message": "Blockchain network is temporarily unavailable",
    "details": {
      "network_id": "cosmos-tracepost-1",
      "retry_after": 30
    }
  }
}
```

### Error Code Reference

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request data validation failed |
| `AUTHENTICATION_REQUIRED` | 401 | Valid authentication required |
| `INSUFFICIENT_PERMISSIONS` | 403 | User lacks required permissions |
| `RESOURCE_NOT_FOUND` | 404 | Requested resource doesn't exist |
| `DUPLICATE_RESOURCE` | 409 | Resource already exists |
| `RATE_LIMIT_EXCEEDED` | 429 | API rate limit exceeded |
| `BLOCKCHAIN_UNAVAILABLE` | 503 | Blockchain network unavailable |

## 🔒 Rate Limiting

### Rate Limit Tiers

| Tier | Requests/Minute | Burst Limit | Use Case |
|------|-----------------|-------------|----------|
| **Free** | 100 | 10 | Development & testing |
| **Professional** | 1,000 | 50 | Small-medium operations |
| **Enterprise** | 10,000 | 500 | Large-scale operations |
| **Custom** | Negotiable | Custom | Enterprise partnerships |

### Rate Limit Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
X-RateLimit-Retry-After: 60
```

### Rate Limit Response

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "API rate limit exceeded",
    "details": {
      "limit": 1000,
      "remaining": 0,
      "reset_time": "2024-03-15T11:00:00Z",
      "retry_after": 60
    }
  }
}
```

## 🔗 Webhooks

### Webhook Events

#### Batch Events
```json
{
  "event": "batch.created",
  "timestamp": "2024-03-15T10:30:00Z",
  "data": {
    "batch_id": "BATCH-2024-03-001",
    "hatchery_id": "did:tracepost:hatchery:ecoshrimp-001",
    "species": "Litopenaeus vannamei",
    "quantity": 5000000
  }
}
```

#### NFT Events
```json
{
  "event": "nft.minted",
  "timestamp": "2024-03-15T10:30:00Z",
  "data": {
    "token_id": "12345",
    "contract_address": "cosmos1contract...",
    "owner_address": "cosmos1owner...",
    "batch_id": "BATCH-2024-03-001"
  }
}
```

### Webhook Configuration

```http
POST /webhooks/endpoints
Content-Type: application/json

{
  "url": "https://your-app.com/webhooks/tracepost",
  "events": [
    "batch.created",
    "batch.updated",
    "nft.minted",
    "nft.transferred"
  ],
  "secret": "webhook_secret_key",
  "active": true
}
```

## 📦 SDK & Libraries

### Official SDKs

#### JavaScript/TypeScript
```bash
npm install @tracepost/api-client
```

```typescript
import { TracePostClient } from '@tracepost/api-client';

const client = new TracePostClient({
  apiKey: 'your-api-key',
  environment: 'production'
});

// Create a batch
const batch = await client.batches.create({
  hatcheryId: 'did:tracepost:hatchery:001',
  species: 'Litopenaeus vannamei',
  quantity: 1000000
});
```

#### Python
```bash
pip install tracepost-api
```

```python
from tracepost import TracePostClient

client = TracePostClient(
    api_key='your-api-key',
    environment='production'
)

# Create a batch
batch = client.batches.create(
    hatchery_id='did:tracepost:hatchery:001',
    species='Litopenaeus vannamei',
    quantity=1000000
)
```

#### Go
```bash
go get github.com/tracepost/go-client
```

```go
package main

import (
    "github.com/tracepost/go-client"
)

func main() {
    client := tracepost.NewClient("your-api-key")
    
    batch, err := client.Batches.Create(&tracepost.CreateBatchRequest{
        HatcheryID: "did:tracepost:hatchery:001",
        Species:    "Litopenaeus vannamei",
        Quantity:   1000000,
    })
}
```

## 🧪 Testing

### API Testing Tools

#### Postman Collection
- Download our comprehensive Postman collection
- Pre-configured environments for development and production
- Automated test scripts included

#### OpenAPI Testing
```bash
# Generate test clients
openapi-generator generate -i swagger.yaml -g go -o ./test-client
```

### Test Environments

| Environment | Base URL | Purpose |
|-------------|----------|---------|
| **Development** | `https://api-dev.tracepost.com` | Internal development |
| **Staging** | `https://api-staging.tracepost.com` | Pre-production testing |
| **Production** | `https://api.tracepost.com` | Live production API |

## 💡 Best Practices

### Request Guidelines

1. **Use HTTPS**: Always use HTTPS for API requests
2. **Set User-Agent**: Include a descriptive User-Agent header
3. **Handle Retries**: Implement exponential backoff for retries
4. **Cache Responses**: Cache responses when appropriate
5. **Validate Data**: Always validate data before sending requests

### Error Handling

1. **Check HTTP Status**: Always check HTTP status codes
2. **Parse Error Details**: Extract detailed error information
3. **Implement Fallbacks**: Have fallback mechanisms for critical operations
4. **Log Errors**: Log errors for debugging and monitoring

### Security Best Practices

1. **Secure API Keys**: Never expose API keys in client-side code
2. **Use JWT Tokens**: Prefer JWT tokens for user authentication
3. **Validate Webhooks**: Always validate webhook signatures
4. **Rate Limit Awareness**: Monitor and respect rate limits
5. **Data Encryption**: Encrypt sensitive data in transit and at rest

---

<div align="center">

### 🚀 Ready to Integrate with TracePost-LarvaeChain?

[![API Documentation](https://img.shields.io/badge/API-Documentation-blue?style=for-the-badge&logo=swagger)](https://api.tracepost.com/docs)
[![Postman Collection](https://img.shields.io/badge/Postman-Collection-orange?style=for-the-badge&logo=postman)](https://www.postman.com/tracepost/workspace/tracepost-api)
[![SDK Downloads](https://img.shields.io/badge/SDK-Downloads-green?style=for-the-badge&logo=github)](https://github.com/tracepost/sdks)

**Built for Enterprise • Powered by Blockchain • Secured by Design**

</div>
