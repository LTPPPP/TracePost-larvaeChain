# TracePost-larvaeChain

A high-performance backend system for shrimp larvae traceability using blockchain technology.

## Overview

TracePost-larvaeChain is a complete traceability solution for the shrimp larvae supply chain that leverages blockchain technology to ensure data integrity and transparency. The system records each step of the supply chain, from hatchery to final distribution, and makes this information verifiable and accessible to all stakeholders, while maintaining compliance with industry standards.

## Technology Stack

- 🐹 **Programming Language**: [Go 1.22+](https://go.dev/) - High-performance, statically typed language with built-in concurrency
- 🚀 **Framework**: [Fiber v2](https://gofiber.io/) - Express-inspired web framework built on top of Fasthttp for high performance (up to 10x faster than net/http)
- ⛓️ **Blockchain**: Custom Layer 1 based on [Cosmos SDK](https://cosmos.network/) v0.47 - Modular blockchain framework supporting IBC protocol
- 📜 **Smart Contracts**: [Solidity](https://soliditylang.org/) v0.8.20 contracts for key events - Support for EVM-compatible blockchain networks
- 🔒 **Consensus Mechanism**: [Tendermint](https://tendermint.com/) v0.35 providing Proof of Authority (PoA) with Byzantine Fault Tolerance (BFT)
- 📚 **API Documentation**: [Swagger UI](https://swagger.io/) (via [gofiber/swagger](https://github.com/gofiber/swagger)) - Tài liệu API tương tác với ví dụ thực tế
- 🗃️ **Database**: [PostgreSQL 16](https://www.postgresql.org/) - Advanced open source relational database with JSONB support
- 🗂️ **Metadata Storage**: [IPFS](https://ipfs.tech/) v0.20 - Distributed, content-addressed storage for immutable data
- ⚡ **Caching**: [Redis](https://redis.io/) v7.2 - In-memory data structure store for high-performance caching (>100K ops/sec)
- 📊 **Tracing & Logging**: Structured JSON logging with [zerolog](https://github.com/rs/zerolog) - High-performance logging with configurable output formats
- 🐳 **Containerization**: [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) - Container orchestration for consistent deployment
- 🪪 **Identity Management**: [Decentralized Identifiers (DIDs)](https://www.w3.org/TR/did-core/) - W3C-compliant DIDs with [Verifiable Credentials](https://www.w3.org/TR/vc-data-model/)
- 🔐 **Authentication**: [JWT](https://jwt.io/) tokens with HMAC SHA-256 signing - Secure, stateless authentication mechanism
- 🌉 **Interoperability**: [IBC Protocol](https://ibcprotocol.org/) - Cross-chain communication between heterogeneous blockchains
- 🧩 **Zero-Knowledge Proofs**: [gnark](https://github.com/ConsenSys/gnark) - Go-based ZKP library for privacy-preserving verification
- 📱 **QR Code Generation**: [go-qrcode](https://github.com/skip2/go-qrcode) - High-quality QR code generation for traceability

## Core Features

1. **Hatchery Management**: Register and manage hatcheries that produce shrimp larvae

   - Detailed hatchery profiles with certification status and history
   - Real-time performance monitoring and reporting
   - Multi-tier access control for hatchery operators and regulators

2. **Batch Creation**: Register new batches of shrimp larvae with detailed information

   - Unique blockchain-verified batch identifiers
   - Species and strain tracking with genetic information
   - Automated batch quality scoring based on environmental data

3. **Supply Chain Events**: Record events throughout the supply chain (feeding, processing, packaging, transportation)

   - Timestamped, immutable event records with blockchain verification
   - Actor authentication using decentralized identifiers
   - Smart contract triggers for critical events (e.g., temperature deviations)

4. **Environment Monitoring**: Track environmental conditions such as temperature, pH, salinity, etc.

   - IoT integration for continuous monitoring
   - Anomaly detection with automated alerts
   - Historical trend analysis and visualization

5. **Document Management**: Upload and verify certificates and other documents

   - Content-addressed storage on IPFS for immutability
   - Document verification using digital signatures
   - Tamper-evident document history

6. **QR Code Generation**: Generate QR codes for batch traceability

   - Cryptographically signed QR codes with anti-counterfeiting features
   - Dynamic QR codes for real-time information
   - Offline verification capabilities

7. **Traceability API**: Public API for end-user verification

   - GraphQL and REST endpoints
   - Rate-limited public access
   - Consumer-facing verification portal

8. **Blockchain Integration**: All critical events are recorded on the blockchain for immutability and transparency

   - Multi-chain support for redundancy
   - Decentralized consensus with BFT

9. **Decentralized Identity**: DID support for hatcheries and other supply chain actors for secure verification

   - W3C-compliant DID implementation
   - Verifiable credentials for certifications
   - Revocation capabilities for compromised identities

10. **NFT Certification**: Digital certificates as NFTs for premium batches

    - ERC-721 and ERC-1155 support
    - Transferable ownership records
    - Royalty mechanisms for certification authorities

11. **Cross-Chain Interoperability**: Share traceability data across multiple blockchain networks

    - IBC protocol for Cosmos ecosystem
    - Cross-chain messaging bridges for Ethereum, Polkadot, and BSC
    - Standards-compliant data formats (GS1 EPCIS)

12. **Compliance Reporting**: Generate reports for regulatory compliance

    - Automated compliance checking
    - Regulatory templates for multiple jurisdictions
    - Audit trail for inspection history

13. **Real-time Analytics Dashboard**: Comprehensive analytics for administrators
    - System performance monitoring and metrics
    - Blockchain health and transaction monitoring
    - Compliance analytics and certification status tracking
    - User activity patterns and engagement metrics
    - Batch production analytics and supply chain insights
    - Exportable reports in multiple formats

## Architecture

The system follows a clean, modular architecture:

- **API Layer**: RESTful API built with Fiber

  - Rate-limited endpoints (100 req/min)
  - JWT authentication with 24hr expiration
  - Swagger documentation
  - GraphQL support for complex queries

- **Service Layer**: Business logic for batch management, events, documents, etc.

  - Domain-driven design principles
  - Event-sourcing architecture for audit trails
  - Command Query Responsibility Segregation (CQRS)
  - Circuit breakers for fault tolerance

- **Data Layer**: PostgreSQL for off-chain data and indexing

  - Connection pooling (max 20 connections)
  - JSONB for flexible schemas
  - Advanced indexing for high-performance queries
  - Transactional integrity with ACID compliance

- **Blockchain Layer**: Custom blockchain based on Cosmos SDK with bridges to other networks

  - Tendermint consensus (10K+ TPS)
  - IBC protocol for cross-chain communication
  - Custom modules for specialized business logic

- **IPFS Layer**: Distributed storage for metadata and documents

  - Content addressing for immutability
  - Pinning service for data availability
  - Gateway access for public verification
  - Encrypted storage for sensitive documents

- **Identity Layer**: W3C-compliant DID implementation for secure identity management

  - Self-sovereign identity principles
  - Verifiable credential issuance and verification
  - Decentralized key management
  - Revocation mechanisms

- **Caching Layer**: Redis for high-speed data access and session management
  - In-memory caching for frequently accessed data
  - Pub/sub for real-time notifications
  - Distributed locking for concurrency control
  - Data persistence for failover recovery

### System Diagram

```
┌────────────────┐      ┌────────────────┐      ┌────────────────┐
│   Client Apps  │◄────►│  API Gateway   │◄────►│  Service Layer │
└────────────────┘      └────────────────┘      └───────┬────────┘
                                                         │
         ┌──────────────────────────────────────────────┼──────────────────────────────────┐
         │                                               │                                  │
         ▼                                               ▼                                  ▼
┌────────────────┐                             ┌────────────────┐                 ┌────────────────┐
│   PostgreSQL   │◄───────────────────────────►│  Blockchain    │◄───────────────►│  IPFS Storage  │
└────────────────┘                             └────────────────┘                 └────────────────┘
         ▲                                               ▲                                  ▲
         │                                               │                                  │
         └──────────────────────┐           ┌────────────┘                  ┌──────────────┘
                                │           │                               │
                                ▼           ▼                               ▼
                         ┌────────────────────────┐             ┌────────────────────────┐
                         │    Redis Cache Layer   │             │  Identity Registry     │
                         └────────────────────────┘             └────────────────────────┘
```

## Project Structure

```
TracePost-larvaeChain/
├── api/                          # API handlers and routes
│   ├── alliance.go               # Alliance chain integration
│   ├── analytics.go              # Batch analytics endpoints
│   ├── api.go                    # Core API setup
│   ├── auth.go                   # Authentication handlers
│   ├── baas.go                   # Blockchain-as-a-Service endpoints
│   ├── batch.go                  # Batch management endpoints
│   ├── company.go                # Company management endpoints
│   ├── compliance.go             # Regulatory compliance endpoints
│   ├── exporter.go               # Export data formatters
│   ├── geo.go                    # Geolocation endpoints
│   ├── handlers.go               # Common handler utilities
│   ├── hatch.go                  # Hatchery management endpoints
│   ├── identity.go               # DID management endpoints
│   ├── interoperability.go       # Cross-chain interop endpoints
│   ├── nft.go                    # NFT certification endpoints
│   ├── processor.go              # Data processor endpoints
│   ├── scaling.go                # Sharding configuration
│   ├── shipment.go               # Shipment management endpoints
│   ├── supplychain.go            # Supply chain event endpoints
│   └── zkp.go                    # Zero-knowledge proof endpoints
├── blockchain/                   # Blockchain integration
│   ├── analytics.go              # On-chain analytics
│   ├── baas.go                   # Blockchain-as-a-Service implementation
│   ├── blockchain.go             # Core blockchain interface
│   ├── ddi_client.go             # Decentralized identity client
│   ├── identity.go               # Identity management
│   ├── interoperability.go       # Cross-chain communication
│   ├── zkp.go                    # Zero-knowledge proof integration
│   └── bridges/                  # Cross-chain interoperability
│       ├── cosmos.go             # Cosmos ecosystem bridge
│       └── polkadot.go           # Polkadot ecosystem bridge
├── cmd/                          # Command-line tools
│   └── ddi-tool/                 # DID management tools
│       └── main.go               # DID CLI tool entry point
├── config/                       # Application configuration
│   ├── baas.go                   # BaaS configuration
│   └── config.go                 # App configuration loader
├── contracts/                    # Smart contracts
│   ├── LogisticsTraceability.sol # Main traceability contract
│   └── LogisticsTraceabilityNFT.sol # NFT certification contract
├── db/                           # Database connection and models
│   ├── db.go                     # Database connection manager
│   ├── nft_monitor.go            # NFT transaction monitor
│   └── transaction_nft.go        # NFT transaction handlers
├── docs/                         # Documentation and Swagger specs
│   ├── ddi_system.md             # DID system documentation
│   ├── docs.go                   # Swagger documentation
│   ├── nft_system.md             # NFT system documentation
│   ├── swagger.json              # OpenAPI specification (JSON)
│   └── swagger.yaml              # OpenAPI specification (YAML)
├── init-scripts/                 # Database initialization scripts
├── ipfs/                         # IPFS integration
│   └── ipfs.go                   # IPFS client interface
├── ipfs-config/                  # IPFS configuration
├── logs/                         # Application logs
├── middleware/                   # Middleware functions
│   ├── ddi_middleware.go         # DID authentication middleware
│   └── middleware.go             # Core middleware components
├── models/                       # Data models
│   └── models.go                 # Core data structures
├── .env                          # Environment variables
├── .env.example                  # Cấu hình môi trường thực tế
├── Dockerfile                    # Docker configuration
├── docker-compose.yml            # Docker Compose configuration
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── main.go                       # Application entry point
└── README.md                     # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- PostgreSQL 16 (optional if using Docker)
- IPFS node (optional if using Docker)
- Redis (optional if using Docker)

### Running with Docker

1. Clone the repository:

   ```bash
   git clone https://github.com/LTPPPP/TracePost-larvaeChain.git
   cd TracePost-larvaeChain
   ```

2. Copy tệp cấu hình môi trường thực tế và chỉnh sửa nếu cần:

   ```bash
   cp .env.example .env
   # Edit .env with your preferred settings
   ```

3. Start the application and all required services:

   ```bash
   docker-compose up -d
   ```

4. Access the API at http://localhost:8080
5. Access the Swagger UI at http://localhost:8080/swagger/index.html
6. Access the database admin panel at http://localhost:8082 (Adminer) or http://localhost:8083 (pgAdmin)

### Running Locally

1. Clone the repository:

   ```bash
   git clone https://github.com/LTPPPP/TracePost-larvaeChain.git
   cd TracePost-larvaeChain
   ```

2. Copy tệp cấu hình môi trường thực tế và chỉnh sửa nếu cần:

   ```bash
   cp .env.example .env
   # Edit .env with your preferred settings
   ```

3. Install dependencies:

   ```bash
   go mod download
   go mod tidy
   ```

4. Set up the required services (PostgreSQL, Redis, IPFS) or use Docker for those components:

   ```bash
   # Start only the dependency services
   docker-compose up -d postgres redis ipfs
   ```

5. Run the application:

   ```bash
   go run main.go
   ```

6. Access the API at http://localhost:8080
7. Access the Swagger UI at http://localhost:8080/swagger/index.html

## API Endpoints

### Blockchain APIs

- `POST /api/v1/blockchain/search` - Search blockchain records.
- `GET /api/v1/blockchain/verify` - Verify blockchain data.
- `POST /api/v1/blockchain/audit` - Perform a blockchain audit.
- `POST /api/v1/blockchain/deploy` - Deploy the LogisticsTraceability contract.

### NFT APIs

- `POST /api/v1/nft/deploy` - Deploy an NFT contract.
- `POST /api/v1/nft/tokenize` - Tokenize a batch.
- `GET /api/v1/nft/batch-details` - Get details of a batch NFT.
- `GET /api/v1/nft/details` - Get details of an NFT.
- `POST /api/v1/nft/transfer` - Transfer an NFT.
- `POST /api/v1/nft/tokenize-transaction` - Tokenize a transaction.
- `GET /api/v1/nft/transaction-details` - Get details of a transaction NFT.
- `GET /api/v1/nft/trace` - Trace a transaction.

### Exporter APIs

- `POST /api/v1/exporter/create` - Create a new exporter.
- `GET /api/v1/exporter/all` - Get all exporters.

### Geo APIs

- `POST /api/v1/geo/record` - Record geolocation data.
- `GET /api/v1/geo/journey` - Get the journey of a batch.
- `GET /api/v1/geo/current-location` - Get the current location of a batch.

### Hatchery APIs

- `GET /api/v1/hatchery/all` - Get all hatcheries.
- `GET /api/v1/hatchery/:id` - Get details of a specific hatchery.

## Deployment Instructions

### Prerequisites

- **Go 1.22+**: Install from [Go's official website](https://go.dev/).
- **Docker**: Ensure Docker and Docker Compose are installed.
- **PostgreSQL**: Set up a PostgreSQL database.
- **Redis**: Install Redis for caching.
- **IPFS**: Install and configure IPFS.

### Steps

1. Clone the repository:

   ```bash
   git clone https://github.com/your-repo/TracePost-larvaeChain.git
   cd TracePost-larvaeChain/back-end
   ```

2. Set up environment variables:

   ```bash
   cp .env.example .env
   # Update .env with your configuration
   ```

3. Build and run the application:

   ```bash
   docker-compose up --build
   ```

4. Access the API documentation:

   - Swagger UI: `http://localhost:8080/swagger`

5. Run tests:
   ```bash
   go test ./...
   ```

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.

## License

This project is licensed under the MIT License. See the `LICENSE` file for details.
