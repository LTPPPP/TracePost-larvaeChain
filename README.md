# 🌐 TracePost-larvaeChain

TracePost-larvaeChain is a state-of-the-art blockchain-based traceability platform specifically designed for the shrimp larvae supply chain. This innovative solution addresses transparency, security, and data integrity challenges in the aquaculture industry, ensuring trust and efficiency across all stakeholders.

---

## 🚀 Features

### User Management & Role-Based Access Control

- Secure user registration, login, and session management.
- Role-based access control for administrators, hatcheries, distributors, and other stakeholders.
- System administration capabilities for managing users, roles, and participating entities (companies/alliances).

### Hatchery & Batch Lifecycle Management

- Register and manage hatcheries with detailed metadata.
- Create and track shrimp larvae batches, including origin, health, and environmental conditions.
- Generate QR codes for easy traceability and batch identification.

### End-to-End Supply Chain Event Tracking

- Record and immutably store all critical events (e.g., feeding, processing, packaging, transportation) on the blockchain.
- Link related documents (via IPFS) to supply chain events for enhanced traceability.

### Real-Time Environmental Monitoring

- Monitor and log environmental conditions such as temperature, pH, and salinity.
- Associate real-time data with specific batches or locations to ensure optimal conditions.

### Document & Compliance Management

- Upload and verify critical documents such as certificates and licenses.
- Utilize IPFS for decentralized storage, ensuring data integrity and regulatory compliance.

### Advanced Traceability with QR Codes & Geo-Tracking

- Generate QR codes for quick access to batch history and detailed product information.
- Track the geographical location of shipments or assets using integrated geo-tracking capabilities.

### Robust Blockchain Integration & Data Integrity

- Leverage a custom blockchain (potentially Cosmos SDK-based) for immutable data recording.
- Utilize NFTs to represent unique batches or certificates, ensuring authenticity and transparency.

### Decentralized Identity (DID) & Self-Sovereign Identity (SSI)

- Securely verify supply chain actors using Decentralized Identity standards (W3C DID).
- Manage permissions via smart contracts for enhanced security and trust.

### Interoperability & Data Exchange

- Facilitate data sharing with external blockchain systems (e.g., Cosmos, Polkadot).
- Support data export in standard formats like GS1 EPCIS for global compatibility.

### Powerful Analytics & Reporting

- Gain actionable insights into supply chain performance.
- Detect anomalies and generate comprehensive reports for informed decision-making.

### Enhanced Security with Zero-Knowledge Proofs (ZKP)

- Implement ZKPs for private data verification, enhancing privacy without compromising verifiability.

### Internationalization Support

- Designed for global use with multi-language support for API responses and user interfaces.

### Blockchain-as-a-Service (BaaS) Integration

- Flexible deployment with support for Blockchain-as-a-Service platforms (e.g., Azure, IBM).
- Alternative to self-managed blockchain infrastructure for scalability and ease of use.

---

## 🛠️ Technology Stack

| Component                                                                                                                  | Technology                                   |
| -------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------- |
| ![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)                                  | Backend programming language                 |
| ![Fiber](https://img.shields.io/badge/Fiber-333333?style=for-the-badge&logo=fiber&logoColor=white)                         | High-performance web framework               |
| ![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-2E3148?style=for-the-badge&logo=cosmos&logoColor=white)            | Custom Layer 1 blockchain                    |
| ![PostgreSQL](https://img.shields.io/badge/PostgreSQL-336791?style=for-the-badge&logo=postgresql&logoColor=white)          | Database for metadata and off-chain data     |
| ![IPFS](https://img.shields.io/badge/IPFS-65C2CB?style=for-the-badge&logo=ipfs&logoColor=white)                            | Decentralized storage for documents          |
| ![Swagger](https://img.shields.io/badge/Swagger-85EA2D?style=for-the-badge&logo=swagger&logoColor=black)                   | API documentation                            |
| ![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)                      | Containerization                             |
| ![OpenTelemetry](https://img.shields.io/badge/OpenTelemetry-7C3AED?style=for-the-badge&logo=opentelemetry&logoColor=white) | Tracing and logging                          |
| ![Next.js](https://img.shields.io/badge/Next.js-000000?style=for-the-badge&logo=next.js&logoColor=white)                   | Frontend framework for server-side rendering |
| ![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white)     | Styling framework                            |

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

### Running the Backend

1. Clone the repository:

   ```bash
   git clone https://github.com/LTPPPP/TracePost-larvaeChain.git
   cd TracePost-larvaeChain/back-end
   ```

2. Install dependencies:

   ```bash
   go mod download
   ```

3. Set up the database:

   ```bash
   # Create a PostgreSQL database named 'tracepost'
   # Update the .env file with your database credentials
   ```

4. Run the backend application:

   ```bash
   go run main.go
   ```

5. Access the API at [http://localhost:8080](http://localhost:8080).
6. Access the Swagger UI for API documentation at [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html).

### Running the Frontend

#### Web Application

1. Navigate to the frontend directory:

   ```bash
   cd TracePost-larvaeChain/front-end/web/tracepost
   ```

2. Install dependencies:

   ```bash
   # Using npm:
   npm install

   # Using yarn:
   yarn install

   # Using pnpm:
   pnpm install

   # Using bun:
   bun install
   ```

3. Start the development server:

   ```bash
   # Using npm:
   npm run dev

   # Using yarn:
   yarn dev

   # Using pnpm:
   pnpm dev

   # Using bun:
   bun run dev
   ```

4. Access the web application at [http://localhost:3000](http://localhost:3000).

#### Mobile Application

1. Navigate to the mobile app directory:

   ```bash
   cd TracePost-larvaeChain/front-end/app/tracepost
   ```

2. Install dependencies:

   ```bash
   # Using npm:
   npm install

   # Using yarn:
   yarn install

   # Using pnpm:
   pnpm install

   # Using bun:
   bun install
   ```

3. Start the Metro bundler:

   ```bash
   # Using npm:
   npm start

   # Using yarn:
   yarn start

   # Using pnpm:
   pnpm start

   # Using bun:
   bun run start
   ```

4. Run the mobile application on an emulator or physical device:

   ```bash
   # For Android:
   npm run android
   yarn android
   pnpm android
   bun run android

   # For iOS:
   npm run ios
   yarn ios
   pnpm ios
   bun run ios
   ```

### Additional Notes

#### Backend Configuration

- Ensure the `.env` file is properly configured with the following variables:

  - `DB_HOST`: Hostname of the PostgreSQL server.
  - `DB_PORT`: Port number for the database.
  - `DB_USER`: Username for the database.
  - `DB_PASSWORD`: Password for the database.
  - `DB_NAME`: Name of the database.

- IPFS integration requires an active IPFS node. Update the `.env` file with:
  - `IPFS_API_URL`: URL of the IPFS API endpoint.
  - `IPFS_API_KEY`: API key for accessing IPFS (if applicable).

#### Frontend Configuration

- For the web application, ensure the `next.config.ts` file is properly set up for API endpoints:

  - `API_BASE_URL`: Base URL for the backend API.

- For the mobile application, update the `app.json` file with:
  - `expo.android.package`: Android package name.
  - `expo.ios.bundleIdentifier`: iOS bundle identifier.

#### Docker Setup

- Use the `docker-compose.yml` file in the root directory to start both backend and frontend services:

  ```bash
  docker-compose up -d
  ```

- Verify the services are running:

  ```bash
  docker ps
  ```

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
