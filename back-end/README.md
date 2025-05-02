# VieChains: Blockchain Logistics Traceability

A blockchain-based solution for supply chain tracking and logistics traceability using Ethereum smart contracts and distributed storage systems.

## 📜 Overview

VieChains is an end-to-end logistics traceability platform that combines blockchain technology with traditional logistics management to provide immutable, transparent tracking of shipments throughout the supply chain. The system enables real-time tracking, verification of authenticity, and secure sharing of logistics information among stakeholders.

## 🚢 Key Features

- **Blockchain-Powered Traceability**: Record shipment events on Ethereum or other EVM-compatible blockchains for immutable verification
- **Cross-Chain Compatibility**: Optional bridge service for transferring shipment data between different blockchains
- **Sharding Support**: Horizontal scaling for high-volume logistics operations
- **Real-Time Event Recording**: Track shipment status changes with timestamped blockchain verification
- **Role-Based Access Control**: Separate permissions for shippers, warehouses, customs officials, and other stakeholders
- **RESTful API**: Comprehensive API for integration with existing logistics systems
- **Swagger Documentation**: Interactive API documentation for developers
- **Optimized Smart Contract**: Gas-efficient Solidity contract for tracking logistics events

## 🏗️ System Architecture

The system consists of the following components:

1. **Backend API**: Node.js Express server that handles business logic and blockchain interactions
2. **Smart Contracts**: Solidity contracts deployed on Ethereum for verifiable record-keeping
3. **Frontend**: Web interface for tracking shipments and recording events (not included in this repository)
4. **Storage Layer**: Hybrid storage system using local files and optional IPFS integration
5. **Blockchain Bridge**: Optional service for cross-chain compatibility
6. **Sharding Service**: Optional horizontal scaling for high-volume operations

## 🔧 Technology Stack

- **Backend**: Node.js, Express
- **Blockchain**: Ethereum (Solidity), Ethers.js
- **Storage**: File system, MongoDB (optional), IPFS (optional)
- **API Documentation**: Swagger/OpenAPI
- **Authentication**: JWT (JSON Web Tokens)
- **Deployment**: Docker, Docker Compose

## 🚀 Getting Started

### Prerequisites

- [Docker](https://www.docker.com/get-started) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Node.js](https://nodejs.org/) v16 or higher (for local development)
- [Git](https://git-scm.com/)

### Quick Start

1. Clone the repository:

   ```bash
   git clone https://github.com/LTPPPP/blockchain-logistics-traceability.git
   cd blockchain-logistics-traceability
   cd back-end
   ```

2. Set up environment variables:

   ```bash
   cp back-end/.env.example back-end/.env
   # Edit the .env file with your configuration
   ```

3. Start the application using Docker:

   ```bash
   docker-compose up -d
   ```

4. The application will be available at:
   - API: http://localhost:7070
   - API Documentation: http://localhost:7070/api-docs
   - Frontend (if enabled): http://localhost:80

### Configuration Options

#### Enabling Blockchain Features

By default, blockchain features are disabled for easier testing. To enable them:

1. Uncomment the Ganache service in `docker-compose.yml` for local blockchain development
2. Set the following environment variables in `docker-compose.yml`:
   ```
   - BLOCKCHAIN_ENABLED=true
   - BLOCKCHAIN_ETHEREUM_ENABLED=true
   - BLOCKCHAIN_ETHEREUM_NODE_URL=http://ganache:8545
   ```
3. Deploy the smart contract and update `BLOCKCHAIN_ETHEREUM_CONTRACT_ADDRESS`

#### Using IPFS for Decentralized Storage

To enable IPFS integration:

1. Uncomment the IPFS service in `docker-compose.yml`
2. Set the following environment variables:
   ```
   - IPFS_ENABLED=true
   - IPFS_API_URL=http://ipfs:5001
   - IPFS_GATEWAY_URL=http://ipfs:8080
   ```

#### Advanced Features

For production deployments, additional features can be enabled:

- **Sharding**: Set `SHARDING_ENABLED=true` to enable horizontal scaling
- **Cross-Chain Bridge**: Set `BRIDGE_ENABLED=true` to enable transfers between blockchains
- **MongoDB**: Configure connection string in environment variables to use MongoDB instead of file storage

## 📝 API Documentation

The API documentation is available at `http://localhost:7070/api-docs` when the application is running. The Swagger UI provides a detailed overview of all available endpoints and allows for interactive testing.

Key API endpoints:

- `/api/v1/auth`: Authentication endpoints (register, login)
- `/api/v1/shipments`: Shipment management endpoints
- `/api/v1/events`: Event recording endpoints
- `/api/v1/tracing`: Verification and tracing endpoints

## 🧪 Development

### Local Development Setup

1. Install dependencies:

   ```bash
   cd back-end
   npm install
   ```

2. Start services in development mode:

   ```bash
   npm run dev
   ```

3. Run tests:
   ```bash
   npm test
   ```

### Smart Contract Development

The Solidity smart contract (`LogisticsTraceability.sol`) can be deployed using tools like Truffle or Hardhat. For local development:

1. Start Ganache for a local Ethereum network
2. Deploy the contract using your preferred tool
3. Update `BLOCKCHAIN_ETHEREUM_CONTRACT_ADDRESS` in your `.env` file

## 🔐 Security Considerations

For production deployments, ensure you:

1. Use a secure `JWT_SECRET` environment variable
2. Properly secure private keys for blockchain interactions
3. Set up proper firewalls and access controls
4. Implement rate limiting for API endpoints
5. Configure HTTPS for all traffic

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
