# 🔗 Blockchain Logistics Traceability System

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Python](https://img.shields.io/badge/Python-3.9%2B-blue)
![FastAPI](https://img.shields.io/badge/FastAPI-0.100.0%2B-green)
![License](https://img.shields.io/badge/license-MIT-brightgreen)

A robust and secure blockchain-based system for tracking and verifying logistics operations across multiple supply chains. This system leverages blockchain technology to ensure data integrity, traceability, and transparency throughout the logistics process.

## 🌟 Features

- **Multi-blockchain Support**: Ethereum, Substrate, and VietnamChain integration
- **Secure Authentication**: JWT-based authentication with role-based access control
- **Comprehensive API**: RESTful API with Swagger UI documentation
- **Real-time Tracking**: Monitor shipments and events in real-time
- **Document Verification**: Verify document authenticity using blockchain
- **Event Logging**: Record and verify logistics events on the blockchain
- **Alert System**: Automated alerts for logistics anomalies
- **Audit Trail**: Complete audit trail for all system operations

## 🚀 Getting Started

### Prerequisites

- Python 3.9 or higher
- PostgreSQL or SQLite (for development)
- Ethereum node (or Ganache for development)
- Node.js and npm (for front-end)

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/yourusername/blockchain-logistics-traceability.git
   cd blockchain-logistics-traceability/back-end
   ```

2. Install the required packages:

   ```bash
   pip install -r requirements.txt
   ```

3. Set up your environment variables by creating a `.env` file:

   ```
   # Basic configuration
   SECRET_KEY=your_secret_key_at_least_32_characters_long
   DATABASE_URL=postgresql://username:password@localhost:5432/blockchain_logistics
   ASYNC_DATABASE_URL=postgresql+asyncpg://username:password@localhost:5432/blockchain_logistics
   DEBUG=true

   # Blockchain settings - Ethereum
   BLOCKCHAIN_ETHEREUM_ENABLED=true
   BLOCKCHAIN_ETHEREUM_NODE_URL=http://localhost:8545
   BLOCKCHAIN_ETHEREUM_PRIVATE_KEY=your_private_key
   BLOCKCHAIN_ETHEREUM_CHAIN_ID=1337
   ```

4. Initialize the database:

   ```bash
   alembic upgrade head
   ```

5. Run the application:

   ```bash
   uvicorn app.main:app --reload
   ```

6. Access the API documentation at `http://localhost:7070/docs`

### Running with Docker

1. Make sure Docker and Docker Compose are installed on your system.

2. Build and start the containers:

   ```bash
   docker-compose up -d
   ```

   This will start both the PostgreSQL database and the application.

3. Access the API documentation at `http://localhost:7070/docs`

### Running with Virtual Environment (venv)

1. Create and activate a virtual environment:

   ```bash
   # Create virtual environment
   python -m venv venv

   # Activate on Windows
   venv\Scripts\activate

   # Activate on Linux/Mac
   source venv/bin/activate
   ```

2. Install dependencies:

   ```bash
   pip install -r requirements.txt
   ```

3. Set up PostgreSQL database:

   - Install PostgreSQL on your system
   - Create a database named 'logistics_traceability'
   - Update the `.env` file with your database credentials

4. Create the `.env` file with local database configuration:

   ```
   SECRET_KEY=your_secret_key_at_least_32_characters_long
   DATABASE_URL=postgresql://postgres:postgres@localhost:5432/logistics_traceability
   ASYNC_DATABASE_URL=postgresql+asyncpg://postgres:postgres@localhost:5432/logistics_traceability
   DEBUG=true
   BLOCKCHAIN_ETHEREUM_ENABLED=false
   CORS_ORIGINS=http://localhost:3000,http://localhost:7070
   ```

5. Initialize the database schema:

   ```bash
   alembic upgrade head
   ```

6. Run the application:

   ```bash
   uvicorn app.main:app --reload
   ```

7. Access the API documentation at `http://localhost:7070/docs`

## 📚 Documentation

### API Endpoints

The system provides comprehensive RESTful API endpoints for:

- User authentication and management
- Shipment tracking and management
- Event recording and verification
- Blockchain verification of logistics data
- Document storage and verification
- Alerts and notifications

### Blockchain Integration

The system supports multiple blockchain networks:

- **Ethereum**: For primary data verification and smart contract execution
- **Substrate**: For scalable and flexible custom chain operations
- **VietnamChain**: For region-specific blockchain operations

Smart contracts are used for:

- Shipment registry
- Event logging
- Access control
- Sensor data verification

## 🏗️ Architecture

The system follows a clean architecture pattern with:

- **API Layer**: FastAPI routes and endpoints
- **Service Layer**: Business logic implementation
- **Repository Layer**: Data access and persistence
- **Blockchain Layer**: Blockchain communication and verification
- **Model Layer**: Database schema and data models
- **Schema Layer**: Data validation and transformation

## 🧪 Testing

Run tests with pytest:

```bash
pytest
```

## 🛠️ Project Structure

```
back-end/
├── alembic/               # Database migration scripts
├── app/                   # Main application code
│   ├── api/               # API routes and middleware
│   ├── blockchain/        # Blockchain integration
│   ├── core/              # Core functionality
│   ├── db/                # Database configuration and repositories
│   ├── models/            # Database models
│   ├── oracle/            # External data source integration
│   ├── schemas/           # Pydantic schemas for validation
│   ├── services/          # Business logic services
│   └── utils/             # Utility functions
├── tests/                 # Test suite
└── .env                   # Environment variables
```

## 🔐 Security

This system implements several security measures:

- JWT-based authentication
- Role-based access control
- Blockchain verification of data integrity
- HTTPS encryption
- Input validation
- Audit logging

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see below for details:

## 📞 Contact

For any inquiries, please contact:

- **Email**: [support@blockchain-logistics.com](mailto:lamphat279@gmail.com)
- **GitHub Issues**: [GitHub Repository](https://github.com/LTPPPP/blockchain-logistics-traceability/issues)
