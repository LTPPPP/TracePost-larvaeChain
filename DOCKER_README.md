# TracePost - Docker Compose Setup

This docker-compose.yml file orchestrates the entire TracePost blockchain logistics traceability application.

## Project Components

- **Backend API**: Go-based API server with PostgreSQL, Redis, IPFS, and blockchain integration
- **Web Frontend**: Next.js web application
- **Mobile Frontend**: Expo React Native app (web version)
- **Database**: PostgreSQL 16 with initialization scripts
- **Cache**: Redis for caching and sessions
- **Storage**: IPFS for decentralized file storage
- **Blockchain**: Ganache CLI for blockchain simulation
- **Admin Tools**: Adminer and PgAdmin for database management

## Prerequisites

- Docker and Docker Compose installed
- At least 4GB RAM available for containers
- Ports 3000, 8080, 8081, 8082, 8083, 5001, 5432, 6379, 8545, 19006 available

## Environment Setup

1. Copy environment variables from the backend:
```bash
cp back-end/.env.example back-end/.env
```

2. Configure the environment variables in `back-end/.env`:
```env
DB_USER=postgres
DB_PASSWORD=your_secure_password
DB_NAME=tracepost
BLOCKCHAIN_CHAIN_ID=1337
PGADMIN_EMAIL=admin@tracepost.com
PGADMIN_PASSWORD=your_admin_password
```

## Running the Application

### Start All Services
```bash
docker-compose up -d
```

### Start Specific Services
```bash
# Backend only with dependencies
docker-compose up -d postgres redis ipfs blockchain-mock api

# Frontend only (requires backend)
docker-compose up -d web-frontend mobile-frontend
```

### With Production Nginx (Optional)
```bash
docker-compose --profile production up -d
```

## Service URLs

- **Web Application**: http://localhost:3000
- **Mobile Web App**: http://localhost:19006  
- **API Server**: http://localhost:8080
- **API Metrics**: http://localhost:9090
- **IPFS API**: http://localhost:5001
- **IPFS Gateway**: http://localhost:8081
- **Blockchain RPC**: http://localhost:8545
- **Adminer**: http://localhost:8082
- **PgAdmin**: http://localhost:8083
- **Nginx (Production)**: http://localhost:80

## Health Checks

All services include health checks. Monitor status with:
```bash
docker-compose ps
```

## Logs

View logs for specific services:
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f web-frontend
docker-compose logs -f mobile-frontend
```

## Development Mode

For development with hot reload:
```bash
# Backend development
cd back-end
go run main.go

# Web frontend development  
cd front-end/web/tracepost
npm run dev

# Mobile frontend development
cd front-end/app/tracepost
npm run web
```

## Data Persistence

The following volumes persist data:
- `postgres-data`: Database data
- `redis-data`: Redis cache data  
- `ipfs-data`: IPFS node data
- `pgadmin-data`: PgAdmin configuration

## Troubleshooting

### Container Won't Start
```bash
# Check logs
docker-compose logs [service-name]

# Rebuild containers
docker-compose build --no-cache [service-name]
```

### Database Connection Issues
```bash
# Reset database
docker-compose down
docker volume rm tracepost_postgres-data
docker-compose up -d postgres
```

### Port Conflicts
Edit the docker-compose.yml port mappings if you have conflicts:
```yaml
ports:
  - "3001:3000"  # Change external port
```

## Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes
docker-compose down -v

# Stop and remove images
docker-compose down --rmi all
```

## Production Deployment

1. Set production environment variables
2. Use the nginx profile for reverse proxy
3. Configure SSL certificates
4. Set up proper firewall rules
5. Configure backup for persistent volumes

```bash
docker-compose --profile production up -d
```

## Network Architecture

All services run on the `tracepost-network` bridge network (172.20.0.0/16) for inter-service communication.

## Security Notes

- Change default passwords in production
- Use proper SSL certificates  
- Configure firewall rules
- Regularly update container images
- Backup persistent data volumes 