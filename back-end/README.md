# TracePost-larvaeChainChain

A high-performance backend system for shrimp larvae traceability using blockchain technology.

## Overview

TracePost-larvaeChainChain is a complete traceability solution for the shrimp larvae supply chain that leverages blockchain technology to ensure data integrity and transparency. The system records each step of the supply chain, from hatchery to final distribution, and makes this information verifiable and accessible to all stakeholders.

## Technology Stack

- **Programming Language**: Golang
- **Framework**: Fiber (for high performance and concurrency)
- **Blockchain**: Custom Layer 1 based on Cosmos SDK
- **Smart Contracts**: Simple contracts for key events (batch creation, environment updates, processing, packaging, transportation)
- **Consensus Mechanism**: Proof of Authority (PoA) or Byzantine Fault Tolerance (BFT)
- **API Documentation**: Swagger UI (via swaggo/fiber-swagger)
- **Database**: PostgreSQL (for metadata and off-chain data)
- **Metadata Storage**: IPFS (for images, certificates, and other documents)
- **Tracing & Logging**: OpenTelemetry (ready for integration)
- **Containerization**: Docker and Docker Compose

## Core Features

1. **Hatchery Management**: Register and manage hatcheries that produce shrimp larvae
2. **Batch Creation**: Register new batches of shrimp larvae with detailed information
3. **Supply Chain Events**: Record events throughout the supply chain (feeding, processing, packaging, transportation)
4. **Environment Monitoring**: Track environmental conditions such as temperature, pH, salinity, etc.
5. **Document Management**: Upload and verify certificates and other documents
6. **QR Code Generation**: Generate QR codes for batch traceability
7. **Traceability API**: Public API for end-user verification
8. **Blockchain Integration**: All critical events are recorded on the blockchain for immutability and transparency
9. **Decentralized Identity**: DID support for hatcheries and other supply chain actors for secure verification

## Architecture

The system follows a clean, modular architecture:

- **API Layer**: RESTful API built with Fiber
- **Service Layer**: Business logic for batch management, events, documents, etc.
- **Data Layer**: PostgreSQL for off-chain data and indexing
- **Blockchain Layer**: Custom blockchain based on Cosmos SDK
- **IPFS Layer**: Storage for metadata and documents

## Project Structure

```
TracePost-larvaeChain/
├── api/              # API handlers and routes
├── blockchain/       # Blockchain integration
├── config/           # Application configuration
├── db/               # Database connection and models
├── ipfs/             # IPFS integration
├── middleware/       # Middleware functions
├── models/           # Data models
├── .env              # Environment variables
├── Dockerfile        # Docker configuration
├── docker-compose.yml # Docker Compose configuration
├── go.mod            # Go module definition
├── main.go           # Application entry point
└── README.md         # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL (optional if using Docker)
- IPFS node (optional if using Docker)

### Running with Docker

1. Clone the repository:

   ```bash
   git clone https://github.com/LTPPPP/TracePost-larvaeChain.git
   cd TracePost-larvaeChain
   ```

2. Start the application and all required services:

   ```bash
   docker-compose up -d
   ```

3. Access the API at http://localhost:8080
4. Access the Swagger UI at http://localhost:8080/swagger/index.html

### Running Locally

1. Clone the repository:

   ```bash
   git clone https://github.com/vietchain/TracePost-larvaeChain.git
   cd TracePost-larvaeChain
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Set up the database:

   ```bash
   # Create a PostgreSQL database named 'tracepost'
   # Update .env file with your database credentials
   ```

4. Run the application:

   ```bash
   go run main.go
   ```

5. Access the API at http://localhost:8080
6. Access the Swagger UI at http://localhost:8080/swagger/index.html

## API Endpoints

The following main API endpoints are available:

### Authentication

- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration

### Hatcheries

- `GET /api/v1/hatcheries` - Get all hatcheries
- `GET /api/v1/hatcheries/:hatcheryId` - Get hatchery by ID
- `POST /api/v1/hatcheries` - Create a new hatchery (requires admin/manager role)
- `PUT /api/v1/hatcheries/:hatcheryId` - Update an existing hatchery (requires admin/manager role)
- `DELETE /api/v1/hatcheries/:hatcheryId` - Delete a hatchery (requires admin role)
- `GET /api/v1/hatcheries/:hatcheryId/batches` - Get all batches for a specific hatchery
- `GET /api/v1/hatcheries/:hatcheryId/stats` - Get statistics for a specific hatchery

### Batches

- `GET /api/v1/batches` - Get all batches
- `GET /api/v1/batches/:batchId` - Get batch by ID
- `POST /api/v1/batches` - Create a new batch
- `PUT /api/v1/batches/:batchId/status` - Update batch status
- `GET /api/v1/batches/:batchId/events` - Get batch events
- `GET /api/v1/batches/:batchId/documents` - Get batch documents
- `GET /api/v1/batches/:batchId/environment` - Get batch environment data
- `GET /api/v1/batches/:batchId/qr` - Generate batch QR code
- `GET /api/v1/batches/:batchId/history` - Get batch blockchain history

### Events

- `POST /api/v1/events` - Create a new event

### Environment

- `POST /api/v1/environment` - Record environment data

### Documents

- `POST /api/v1/documents` - Upload a document
- `GET /api/v1/documents/:documentId` - Get document by ID

### QR Code Tracing

- `GET /api/v1/qr/:code` - Trace by QR code

### Users

- `GET /api/v1/users/me` - Get current user
- `PUT /api/v1/users/me` - Update current user
- `PUT /api/v1/users/me/password` - Change password

### Interoperability (New for 2025)

- `POST /api/v1/interop/chains` - Register external chain for interoperability
- `POST /api/v1/interop/share-batch` - Share batch with external chain
- `GET /api/v1/interop/export/:batchId` - Export batch data to GS1 EPCIS format

### Identity (New for 2025)

- `POST /api/v1/identity/create` - Create a new decentralized identity
- `GET /api/v1/identity/resolve/:did` - Resolve a decentralized identifier
- `POST /api/v1/identity/claims` - Create a verifiable claim
- `GET /api/v1/identity/claims/verify/:claimId` - Verify a claim
- `POST /api/v1/identity/claims/revoke/:claimId` - Revoke a claim

## Data Model

### Hatchery

The hatchery is the origin point in the supply chain where shrimp larvae are produced:

```json
{
  "id": 1,
  "name": "Ocean Blue Hatchery",
  "location": "Da Nang, Vietnam",
  "contact": "contact@oceanbluehatchery.com",
  "created_at": "2025-01-15T08:00:00Z",
  "updated_at": "2025-04-20T10:15:00Z",
  "batches": [...] // Related batches
}
```

### Batch

A batch represents a group of shrimp larvae produced by a hatchery:

```json
{
  "id": 1,
  "batch_id": "BATCH-12345-1620000000",
  "hatchery_id": 1,
  "creation_date": "2025-05-03T08:00:00Z",
  "species": "Litopenaeus vannamei",
  "quantity": 50000,
  "status": "created",
  "blockchain_tx_id": "0x123456789abcdef",
  "metadata_hash": "QmZ9...a1b2c3"
}
```

## Future Enhancements

- **Advanced Hatchery Analytics**: Predictive analytics for hatchery performance and disease risk assessment
- **GS1 EPCIS Integration**: Bridge module for mapping data to GS1 EPCIS standard
- **Multi-Blockchain Support**: Bridge to Ethereum, Polygon, or Hyperledger
- **Advanced Analytics**: Machine learning for environmental data analysis
- **Mobile Application**: Companion mobile app for scanning QR codes and viewing traceability data
- **Geospatial Tracking**: Real-time location tracking for transportation of larvae batches

## License

[MIT License](LICENSE)

## Contact

For questions or support, please contact support@vietchain.com.
