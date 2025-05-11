# 🌐 TracePost-larvaeChainChain

TracePost-larvaeChainChain is a comprehensive blockchain-based logistics traceability system designed specifically for the shrimp larvae supply chain. It ensures transparency, data integrity, and traceability across the entire supply chain, leveraging cutting-edge technologies to provide a secure and efficient solution for all stakeholders.

---

## 🚀 Features

- **Hatchery Management**: Register and manage hatcheries producing shrimp larvae, ensuring accurate tracking of origins.
- **Batch Creation**: Record detailed information about shrimp larvae batches, including origin, health, and environmental conditions.
- **Supply Chain Events**: Track events like feeding, processing, packaging, and transportation, ensuring end-to-end traceability.
- **Environment Monitoring**: Monitor environmental conditions (temperature, pH, salinity, etc.) in real-time to ensure optimal conditions for shrimp larvae.
- **Document Management**: Upload and verify certificates, licenses, and other critical documents for compliance and transparency.
- **QR Code Generation**: Generate QR codes for batch traceability, enabling quick access to batch history and details.
- **Blockchain Integration**: Immutable recording of critical events on a custom blockchain, ensuring data integrity and transparency.
- **Decentralized Identity (DID)**: Secure verification for supply chain actors using decentralized identity standards.
- **Interoperability**: Share data with external chains and export to GS1 EPCIS format for global compatibility.
- **Analytics and Insights**: Gain actionable insights into supply chain performance and detect anomalies using integrated analytics.

---

## 🛠️ Technology Stack

| Technology                                                                                                                 | Description                              |
| -------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| ![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)                                  | Backend programming language             |
| ![Fiber](https://img.shields.io/badge/Fiber-333333?style=for-the-badge&logo=fiber&logoColor=white)                         | High-performance web framework           |
| ![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-2E3148?style=for-the-badge&logo=cosmos&logoColor=white)            | Custom Layer 1 blockchain                |
| ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white)          | Database for metadata and off-chain data |
| ![IPFS](https://img.shields.io/badge/IPFS-65C2CB?style=for-the-badge&logo=ipfs&logoColor=white)                            | Decentralized storage for documents      |
| ![Swagger](https://img.shields.io/badge/Swagger-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)                   | API documentation                        |
| ![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)                      | Containerization                         |
| ![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-7C3AED?style=for-the-badge&logo=opentelemetry&logoColor=white) | Tracing and logging                      |
| ![Next.js](https://img.shields.io/badge/Next.js-000000?style=for-the-badge&logo=next.js&logoColor=white)                   | Frontend framework for SSR               |
| ![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white)     | Styling framework                        |

---

## 📂 Project Structure

```
blockchain-logistics-traceability/
├── back-end/          # Backend services
│   ├── api/           # API handlers and routes
│   ├── blockchain/    # Blockchain integration
│   ├── db/            # Database models and connections
│   ├── ipfs/          # IPFS integration
│   ├── main.go        # Application entry point
│   └── ...
├── front-end/         # Frontend application
│   ├── src/           # Source code
│   ├── public/        # Static assets
│   └── ...
├── docker-compose.yml # Docker Compose configuration
├── LICENSE            # License file
└── README.md          # Project documentation
```

---

## 🧑‍💻 Getting Started

### Prerequisites

- [Go](https://golang.org/) 1.21 or higher
- [Docker](https://www.docker.com/) and Docker Compose
- [PostgreSQL](https://www.postgresql.org/) (optional if using Docker)
- [IPFS](https://ipfs.io/) node (optional if using Docker)

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

3. Access the API at [http://localhost:8080](http://localhost:8080).
4. Access the Swagger UI at [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

### Running Locally

1. Clone the repository:

   ```bash
   git clone https://github.com/LTPPPP/TracePost-larvaeChain.git
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

5. Access the API at [http://localhost:8080](http://localhost:8080).
6. Access the Swagger UI at [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

---

## 📖 Use Cases

### Authentication

- Securely log in and register users to access the system.

### Hatcheries

- Manage hatcheries, including registration, updates, and tracking of shrimp larvae origins.

### Batches

- Create and manage shrimp larvae batches with detailed metadata, including health and environmental conditions.

### Events

- Record and track supply chain events such as feeding, processing, packaging, and transportation.

### Environment Monitoring

- Monitor and log environmental conditions like temperature, pH, and salinity to ensure optimal conditions.

### Document Management

- Upload and verify critical documents such as certificates and licenses for compliance.

### QR Code Tracing

- Generate and scan QR codes for batch traceability, enabling quick access to batch history and details.

---

## 📜 License

This project is licensed under the [MIT License](LICENSE).

---

## 📧 Contact

For questions or support, please contact [support@vietchain.com](mailto:support@vietchain.com).
