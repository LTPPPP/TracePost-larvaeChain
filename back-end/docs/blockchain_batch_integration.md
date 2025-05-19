# Blockchain Integration for Batch Traceability

This document outlines the blockchain integration features implemented for batch tracking in the VieChains blockchain logistics traceability system.

## Features Added

### 1. Enhanced Batch Blockchain Models

- Added `BatchBlockchainData` model to represent blockchain batch data
- Added `BlockchainTx` model for blockchain transaction representation

### 2. Enhanced Blockchain Client Functions

- `GetBatchBlockchainData`: Retrieves comprehensive batch data from blockchain
- `VerifyBatchIntegrity`: Verifies batch data integrity against blockchain records
- `VerifyBatchDataOnChain`: Performs comprehensive batch data verification on blockchain
- Enhanced transaction submission with extended metadata

### 3. Enhanced Batch API Endpoints

- `CreateBatch`: Now records extended metadata on blockchain with transaction tracking
- `UpdateBatchStatus`: Records status changes with comprehensive blockchain data
- `GetBatchBlockchainData`: Retrieves batch data directly from blockchain
- `VerifyBatchIntegrity`: Verifies batch integrity against blockchain records
- `GetBatchHistory`: Retrieves complete batch history with blockchain transactions

### 4. New Blockchain API Endpoints

- `SearchBlockchainRecords`: Search for blockchain records with various filters
- `GetBlockchainVerification`: Performs comprehensive batch verification
- `BatchBlockchainAudit`: Generates a complete audit trail from blockchain data
- `GetBatchFromBlockchain`: Retrieves batch data directly from blockchain

## How It Works

### Batch Creation

When a new batch is created:

1. Batch is saved to the database
2. Basic batch data is recorded on blockchain
3. Extended metadata is recorded in a separate blockchain transaction
4. Blockchain transaction IDs and metadata hashes are stored in database
5. A batch creation event is recorded

### Batch Status Updates

When a batch status is updated:

1. Status is updated in database
2. Basic status change is recorded on blockchain
3. Extended status change data is recorded on blockchain
4. A status change event is created and linked to blockchain records

### Blockchain Verification

For batch verification:

1. Database records are compared against blockchain data
2. Any discrepancies are detected and reported
3. Transaction continuity and integrity are verified
4. A comprehensive verification report is generated

### Blockchain Audit

The audit trail includes:

1. All blockchain transactions for the batch
2. All database events with their blockchain records
3. Complete batch history with timestamps and metadata
4. Verification of blockchain data integrity

## Usage

### Retrieving Batch Blockchain Data

```
GET /api/v1/batches/{batchId}/blockchain
```

### Verifying Batch Integrity

```
GET /api/v1/batches/{batchId}/verify
```

### Getting Batch History with Blockchain Data

```
GET /api/v1/batches/{batchId}/history
```

### Searching Blockchain Records

```
POST /api/v1/blockchain/search
{
  "related_table": "batch",
  "related_id": 123,
  "from_timestamp": "2025-01-01T00:00:00Z",
  "to_timestamp": "2025-05-01T00:00:00Z",
  "limit": 100
}
```

### Getting Complete Blockchain Audit

```
GET /api/v1/blockchain/audit/{batchId}
```

## Benefits

- **Enhanced Traceability**: Every batch operation is recorded on blockchain
- **Data Integrity**: Database records are verified against blockchain
- **Comprehensive Audit Trail**: Complete history available for compliance and auditing
- **Tamper Evidence**: Any data tampering can be detected through verification
- **Compliance Support**: Detailed records support regulatory compliance
