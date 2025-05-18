# IPFS Integration Guide for TracePost-larvaeChain

## Overview

This document provides a comprehensive guide to setting up, configuring, and using IPFS (InterPlanetary File System) within the TracePost-larvaeChain system. IPFS is used for distributed, content-addressed storage of documents, certificates, and other immutable data related to shrimp larvae traceability.

## Table of Contents

1. [Introduction to IPFS](#introduction-to-ipfs)
2. [System Architecture](#system-architecture)
3. [Installation and Setup](#installation-and-setup)
   - [Local Development Environment](#local-development-environment)
   - [Docker-based Deployment](#docker-based-deployment)
4. [Configuration](#configuration)
   - [Environment Variables](#environment-variables)
   - [IPFS Node Configuration](#ipfs-node-configuration)
5. [API Integration](#api-integration)
   - [Uploading Documents](#uploading-documents)
   - [Retrieving Documents](#retrieving-documents)
   - [Pinning and Persistence](#pinning-and-persistence)
6. [Security Considerations](#security-considerations)
7. [Performance Optimization](#performance-optimization)
8. [Troubleshooting](#troubleshooting)
9. [Best Practices](#best-practices)

## Introduction to IPFS

InterPlanetary File System (IPFS) is a protocol and peer-to-peer network for storing and sharing data in a distributed file system. IPFS uses content-addressing to uniquely identify each file in a global namespace connecting all computing devices.

Key benefits of using IPFS in TracePost-larvaeChain:

- **Immutability**: Once data is stored, it cannot be modified, ensuring data integrity
- **Decentralization**: No single point of failure
- **Content addressing**: Files are referenced by their content, not location
- **Deduplication**: Identical files are stored only once
- **Offline-first**: Data can be accessed without continuous internet connection

## System Architecture

In TracePost-larvaeChain, IPFS is used as a distributed storage layer for:

- Certificates and compliance documents
- Batch metadata and visual documentation
- Verification evidence and audit trails
- Supply chain event documentation

The system architecture follows this pattern:

```
+---------------------+        +-------------------+        +-------------------+
| TracePost-larvae API| <----> | IPFS Node/Gateway | <----> | IPFS Network      |
+---------------------+        +-------------------+        +-------------------+
         ^                                                          ^
         |                                                          |
         v                                                          v
+---------------------+                                    +-------------------+
| PostgreSQL Database |                                    | Pinning Services  |
+---------------------+                                    +-------------------+
(Stores CIDs & metadata)                                  (Ensures persistence)
```

## Installation and Setup

### Local Development Environment

1. **Install IPFS locally**:

   Download and install IPFS from [https://dist.ipfs.io/#go-ipfs](https://dist.ipfs.io/#go-ipfs)

   ```bash
   # For Windows, using PowerShell:
   Invoke-WebRequest -Uri https://dist.ipfs.io/go-ipfs/v0.13.0/go-ipfs_v0.13.0_windows-amd64.zip -OutFile ipfs.zip
   Expand-Archive -Path ipfs.zip -DestinationPath $HOME\ipfs
   cd $HOME\ipfs\go-ipfs
   .\ipfs.exe init
   ```

2. **Start the IPFS daemon**:

   ```bash
   .\ipfs.exe daemon
   ```

3. **Verify IPFS is running**:

   ```bash
   .\ipfs.exe id
   ```

   This should display your IPFS node's ID and other information.

### Docker-based Deployment

The TracePost-larvaeChain system includes an IPFS node in the docker-compose configuration. This is the recommended approach for production environments.

1. **Use the provided docker-compose.yml**:

   The IPFS service is already configured in the docker-compose.yml file:

   ```yaml
   ipfs:
     image: ipfs/kubo:latest
     container_name: tracepost-ipfs
     restart: always
     ports:
       - "5001:5001" # API port
       - "8081:8080" # Gateway port
     volumes:
       - ipfs-data:/data/ipfs
       - ./ipfs-config:/ipfs-config
     command: ["daemon", "--migrate=true", "--enable-gc"]
     networks:
       - tracepost-network
     healthcheck:
       test: ["CMD", "ipfs", "id"]
       interval: 30s
       timeout: 10s
       retries: 3
       start_period: 15s
   ```

2. **Start the containers**:

   ```bash
   docker-compose up -d
   ```

3. **Verify the IPFS container is running**:

   ```bash
   docker-compose ps ipfs
   ```

## Configuration

### Environment Variables

TracePost-larvaeChain uses the following environment variables for IPFS configuration:

| Variable           | Description                                     | Default Value        | Example                    |
| ------------------ | ----------------------------------------------- | -------------------- | -------------------------- |
| `IPFS_NODE_URL`    | URL of the IPFS API endpoint                    | http://ipfs:5001     | http://localhost:5001      |
| `IPFS_API_KEY`     | API key for remote IPFS service (if applicable) |                      | YourSecretAPIKey           |
| `IPFS_GATEWAY_URL` | Public gateway URL for retrieving content       | https://ipfs.io/ipfs | http://localhost:8080/ipfs |

These variables are defined in the `.env` file and loaded at startup.

### IPFS Node Configuration

Custom IPFS node configuration can be specified in the `ipfs-config` directory. This includes:

1. **Creating a custom IPFS config**:

   Create a file `ipfs-config/ipfs.json` with your custom configuration. For example:

   ```json
   {
     "Addresses": {
       "API": "/ip4/0.0.0.0/tcp/5001",
       "Gateway": "/ip4/0.0.0.0/tcp/8080",
       "Swarm": ["/ip4/0.0.0.0/tcp/4001", "/ip4/0.0.0.0/tcp/8081/ws"]
     },
     "API": {
       "HTTPHeaders": {
         "Access-Control-Allow-Origin": ["*"]
       }
     },
     "Bootstrap": [
       "/dnsaddr/bootstrap.libp2p.io/p2p/QmNnooDu7bfjPFoTZYxMNLWUQJyrVwtbZg5gBMjTezGAJN",
       "/dnsaddr/bootstrap.libp2p.io/p2p/QmQCU2EcMqAqQPR2i9bChDtGNJchTbq5TbXJJ16u19uLTa"
     ],
     "Datastore": {
       "StorageMax": "10GB",
       "StorageGCWatermark": 90,
       "GCPeriod": "1h"
     },
     "Swarm": {
       "ConnMgr": {
         "LowWater": 200,
         "HighWater": 500
       }
     }
   }
   ```

2. **Applying the configuration**:

   For Docker deployment, the configuration file is automatically applied when the container starts.

   For local development:

   ```bash
   ipfs config --json Addresses.API '"/ip4/0.0.0.0/tcp/5001"'
   ipfs config --json Addresses.Gateway '"/ip4/0.0.0.0/tcp/8080"'
   ```

## API Integration

The TracePost-larvaeChain system integrates with IPFS through the Go IPFS HTTP client library (`github.com/ipfs/go-ipfs-api`).

### Uploading Documents

The system provides the `SaveDocumentToIPFS` function in `models/models.go` for uploading documents to IPFS:

```go
// SaveDocumentToIPFS uploads a document to IPFS and returns the CID and URI
func SaveDocumentToIPFS(filePath string) (string, string, error) {
    // Connect to IPFS node
    ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
    if ipfsNodeURL == "" {
        ipfsNodeURL = "http://localhost:5001" // Default to local IPFS node
    }
    sh := shell.NewShell(ipfsNodeURL)

    // Open the file
    file, err := os.Open(filePath)
    if err != nil {
        return "", "", fmt.Errorf("error opening file: %v", err)
    }
    defer file.Close()

    // Add the file to IPFS
    cid, err := sh.Add(file)
    if err != nil {
        return "", "", fmt.Errorf("error adding file to IPFS: %v", err)
    }

    // Construct the IPFS URI using the gateway URL
    gatewayURL := os.Getenv("IPFS_GATEWAY_URL")
    if gatewayURL == "" {
        gatewayURL = "https://ipfs.io/ipfs" // Default to public gateway
    }
    uri := fmt.Sprintf("%s/%s", gatewayURL, cid)

    return cid, uri, nil
}
```

To upload a document via API:

1. **API Endpoint**: `POST /api/v1/documents`
2. **Request Body**:
   ```json
   {
     "batch_id": 123,
     "document_type": "certificate",
     "title": "Health Certificate",
     "description": "Health certificate for batch 123"
   }
   ```
3. **Form Data**: Include the file as `document_file` in multipart/form-data

Example usage:

```bash
curl -X POST http://localhost:8080/api/v1/documents \
  -H "Authorization: Bearer <your-jwt-token>" \
  -F "document_file=@/path/to/certificate.pdf" \
  -F "batch_id=123" \
  -F "document_type=certificate" \
  -F "title=Health Certificate" \
  -F "description=Health certificate for batch 123"
```

### Retrieving Documents

To retrieve a document:

1. **API Endpoint**: `GET /api/v1/documents/:documentId`
2. **Response**:
   ```json
   {
     "id": 1,
     "batch_id": 123,
     "document_type": "certificate",
     "title": "Health Certificate",
     "description": "Health certificate for batch 123",
     "ipfs_cid": "QmX5J3jFvgQKmTJjz2h6brY4zsXmiNMRa3678MXCTexy1B",
     "ipfs_uri": "https://ipfs.io/ipfs/QmX5J3jvgQKmTJjz2h6brY4zsXmiNMRa3678MXCTexy1B",
     "file_size": 12345,
     "file_type": "application/pdf",
     "uploaded_by": "user123",
     "created_at": "2023-05-17T10:30:00Z",
     "updated_at": "2023-05-17T10:30:00Z"
   }
   ```

Implementation of document retrieval function:

```go
func GetIPFSContent(cid string) ([]byte, error) {
    ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
    if ipfsNodeURL == "" {
        ipfsNodeURL = "http://localhost:5001"
    }
    sh := shell.NewShell(ipfsNodeURL)

    reader, err := sh.Cat(cid)
    if err != nil {
        return nil, fmt.Errorf("error retrieving content from IPFS: %v", err)
    }
    defer reader.Close()

    return io.ReadAll(reader)
}
```

### Pinning and Persistence

To ensure documents remain available, they should be pinned to the IPFS node:

```go
func PinIPFSContent(cid string) error {
    ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
    if ipfsNodeURL == "" {
        ipfsNodeURL = "http://localhost:5001"
    }
    sh := shell.NewShell(ipfsNodeURL)

    return sh.Pin(cid)
}
```

For production environments, consider using a pinning service like Pinata or Infura:

```go
func PinToRemoteService(cid string) error {
    apiKey := os.Getenv("PINATA_API_KEY")
    apiSecret := os.Getenv("PINATA_API_SECRET")

    client := &http.Client{}
    req, err := http.NewRequest("POST", "https://api.pinata.cloud/pinning/pinByHash", nil)
    if err != nil {
        return err
    }

    req.Header.Add("pinata_api_key", apiKey)
    req.Header.Add("pinata_secret_api_key", apiSecret)

    q := req.URL.Query()
    q.Add("hashToPin", cid)
    req.URL.RawQuery = q.Encode()

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("pinning failed: %s", body)
    }

    return nil
}
```

## Security Considerations

1. **Access Control**:

   - The IPFS node's API port (5001) should not be publicly accessible
   - Use authentication for the IPFS API if exposed beyond localhost
   - Consider using a reverse proxy with authentication

2. **Content Verification**:

   - Always verify CIDs when retrieving content
   - Store document hashes in the blockchain for additional verification

3. **Private Documents**:
   - For sensitive documents, consider encryption before uploading to IPFS
   - Implement access control at the application level

Example of encrypted upload:

```go
func EncryptAndSaveToIPFS(data []byte, publicKey []byte) (string, error) {
    // Encrypt data using publicKey
    encryptedData, err := encrypt(data, publicKey)
    if err != nil {
        return "", err
    }

    // Upload encrypted data to IPFS
    ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
    sh := shell.NewShell(ipfsNodeURL)

    cid, err := sh.Add(bytes.NewReader(encryptedData))
    if err != nil {
        return "", err
    }

    return cid, nil
}
```

## Performance Optimization

1. **Parallel Uploads**:
   For batch uploads, use goroutines to upload files in parallel:

   ```go
   func BatchUploadToIPFS(files []string) ([]string, error) {
       var wg sync.WaitGroup
       results := make([]string, len(files))
       errors := make([]error, len(files))

       for i, file := range files {
           wg.Add(1)
           go func(idx int, filePath string) {
               defer wg.Done()
               cid, _, err := SaveDocumentToIPFS(filePath)
               if err != nil {
                   errors[idx] = err
                   return
               }
               results[idx] = cid
           }(i, file)
       }

       wg.Wait()

       // Check for errors
       for _, err := range errors {
           if err != nil {
               return nil, err
           }
       }

       return results, nil
   }
   ```

2. **Connection Reuse**:
   Create a singleton IPFS shell instance:

   ```go
   var (
       ipfsShell *shell.Shell
       once      sync.Once
   )

   func GetIPFSShell() *shell.Shell {
       once.Do(func() {
           ipfsNodeURL := os.Getenv("IPFS_NODE_URL")
           if ipfsNodeURL == "" {
               ipfsNodeURL = "http://localhost:5001"
           }
           ipfsShell = shell.NewShell(ipfsNodeURL)
       })
       return ipfsShell
   }
   ```

3. **Content Deduplication**:
   Calculate file hashes before uploading to check if already in IPFS:

   ```go
   func CheckFileExists(filePath string) (bool, string, error) {
       // Calculate file hash
       f, err := os.Open(filePath)
       if err != nil {
           return false, "", err
       }
       defer f.Close()

       h := sha256.New()
       if _, err := io.Copy(h, f); err != nil {
           return false, "", err
       }

       hash := fmt.Sprintf("%x", h.Sum(nil))

       // Check database for existing file with this hash
       var cid string
       err = db.QueryRow("SELECT ipfs_cid FROM documents WHERE file_hash = $1", hash).Scan(&cid)
       if err == nil {
           return true, cid, nil
       }

       return false, "", nil
   }
   ```

## Troubleshooting

### Common IPFS Issues

1. **Connection Refused**:

   - Check if IPFS daemon is running
   - Verify the IPFS_NODE_URL environment variable is correct
   - Ensure network connectivity between API and IPFS node

2. **Slow Uploads/Downloads**:

   - Check IPFS node connectivity to the network
   - Adjust bootstrap peers
   - Consider using a dedicated IPFS cluster for production

3. **Content Not Found**:

   - Verify CID is correct
   - Check if content is pinned to the node
   - Try multiple gateways

4. **Docker Volume Issues**:
   - Ensure persistent volumes are properly configured
   - Check Docker container logs for IPFS errors

### Diagnostic Commands

```bash
# Check IPFS node status
docker-compose exec ipfs ipfs id

# List pinned items
docker-compose exec ipfs ipfs pin ls

# Check IPFS peers
docker-compose exec ipfs ipfs swarm peers

# Check IPFS config
docker-compose exec ipfs ipfs config show
```

## Best Practices

1. **Content Management**:

   - Implement garbage collection policies
   - Schedule regular backups of pinned CIDs
   - Monitor storage usage

2. **Gateway Access**:

   - Use dedicated gateway for high-traffic applications
   - Consider CDN integration for frequently accessed content

3. **Metadata Structure**:

   - Store structured metadata alongside documents
   - Use standardized schemas for document types

4. **Integration with Blockchain**:
   - Store IPFS CIDs on the blockchain for immutable proof
   - Create verifiable links between supply chain events and documentation

Example blockchain integration:

```go
func RegisterDocumentOnBlockchain(documentID int, cid string) error {
    // Get blockchain client
    client := blockchain.GetClient()

    // Register document CID on blockchain
    txID, err := client.RegisterDocument(documentID, cid)
    if err != nil {
        return err
    }

    // Update document record with blockchain transaction ID
    _, err = db.Exec(
        "UPDATE documents SET blockchain_tx_id = $1 WHERE id = $2",
        txID, documentID,
    )

    return err
}
```

---

This guide provides a comprehensive overview of IPFS integration in TracePost-larvaeChain. For specific implementation details, refer to the source code in the `ipfs` directory and the `models` package.
