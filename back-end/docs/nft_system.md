# NFT Transaction System Documentation

## Database Schema

### transaction_nft Table

The `transaction_nft` table stores information about NFTs created for tracking shipments in the logistics chain.

| Column Name          | Data Type   | Description                                                              |
| -------------------- | ----------- | ------------------------------------------------------------------------ |
| id                   | SERIAL      | Primary key for the NFT record                                           |
| tx_id                | TEXT        | Unique transaction ID associated with this NFT                           |
| shipment_transfer_id | TEXT        | Foreign key to shipment_transfer table                                   |
| token_id             | TEXT        | The NFT token ID on the blockchain                                       |
| contract_address     | TEXT        | The smart contract address that minted this NFT                          |
| token_uri            | TEXT        | URI pointing to the NFT metadata (usually IPFS)                          |
| qr_code_url          | TEXT        | URL to the QR code for this NFT                                          |
| owner_address        | TEXT        | Blockchain address of the current owner                                  |
| status               | VARCHAR(30) | Current status of the NFT (active, burned, transferred, locked, expired) |
| blockchain_record_id | INT         | Reference to blockchain_record table                                     |
| batch_id             | INT         | Reference to batch table                                                 |
| metadata             | JSONB       | JSON metadata of the NFT                                                 |
| metadata_schema      | TEXT        | Schema version or type for the metadata                                  |
| digest_hash          | TEXT        | Hash for data integrity verification                                     |
| created_at           | TIMESTAMP   | Creation timestamp                                                       |
| updated_at           | TIMESTAMP   | Last update timestamp                                                    |
| is_active            | BOOLEAN     | Flag indicating if the NFT is active                                     |

### transaction_nft_history Table

The `transaction_nft_history` table tracks all changes to NFTs over time.

| Column Name      | Data Type   | Description                                                                 |
| ---------------- | ----------- | --------------------------------------------------------------------------- |
| id               | SERIAL      | Primary key for the history record                                          |
| nft_id           | INT         | Foreign key to the transaction_nft table                                    |
| previous_status  | VARCHAR(30) | Previous NFT status                                                         |
| new_status       | VARCHAR(30) | New NFT status                                                              |
| previous_owner   | TEXT        | Previous owner address                                                      |
| new_owner        | TEXT        | New owner address                                                           |
| action_type      | VARCHAR(50) | Type of action performed (status_change, ownership_change, metadata_update) |
| action_timestamp | TIMESTAMP   | When the action occurred                                                    |
| action_by        | INT         | User who performed the action                                               |
| tx_id            | TEXT        | Transaction ID for this history record                                      |
| reason           | TEXT        | Reason for the change                                                       |
| metadata_change  | JSONB       | JSON representation of metadata changes                                     |
| created_at       | TIMESTAMP   | Creation timestamp                                                          |

## Data Integrity Mechanisms

### 1. Referential Integrity

The NFT table maintains referential integrity through foreign keys to:

- shipment_transfer table (shipment_transfer_id)
- blockchain_record table (blockchain_record_id)
- batch table (batch_id)

### 2. Data Validation

- Metadata validation through JSON schema checking
- Status validation through CHECK constraints
- Uniqueness constraints to prevent duplicates
- Digest hash for data integrity verification

### 3. History Tracking

- Automatic tracking of all NFT changes via database triggers
- Complete history of status and ownership changes
- Audit trail for compliance and dispute resolution

### 4. Security

- Sensitive data can be encrypted
- Soft delete mechanism instead of hard deletes
- Permission-based access control

## Monitoring and Alerting

### NFT Monitoring System

A monitoring system runs regularly to:

1. Check for data integrity issues
2. Detect duplicate NFTs
3. Validate cross-table references
4. Generate alerts when issues are found

### Logging

Comprehensive logging captures:

1. All NFT operations (create, update, transfer)
2. Errors and exceptions
3. Data validation failures
4. System events

## Best Practices

### 1. Data Insertion

When inserting a new NFT:

- Generate a digest hash for data integrity
- Validate metadata format
- Check for duplicates
- Ensure related records exist

```go
// Example: Creating a new NFT
func CreateNewNFT(shipmentID string, tokenID string, contractAddress string, owner string, metadata []byte) (int, error) {
    // 1. Validate metadata
    if err := ValidateNFTMetadata(metadata); err != nil {
        return 0, fmt.Errorf("invalid metadata: %w", err)
    }

    // 2. Check for duplicates
    var exists bool
    err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM transaction_nft WHERE shipment_transfer_id = $1 AND token_id = $2)",
        shipmentID, tokenID).Scan(&exists)
    if err != nil {
        return 0, fmt.Errorf("failed to check for duplicates: %w", err)
    }
    if exists {
        return 0, fmt.Errorf("NFT with shipment ID %s and token ID %s already exists", shipmentID, tokenID)
    }

    // 3. Insert the NFT
    // ...
}
```

### 2. Data Updates

When updating an NFT:

- The trigger will automatically create history records
- Update the digest hash
- Validate the new status

### 3. Data Queries

When querying NFT data:

- Always include is_active=true in WHERE clause
- Verify data integrity for sensitive operations
- Use proper indexing for performance

### 4. Compliance and Auditing

For compliance purposes:

- Use the history table for audit trails
- Never hard delete records
- Maintain complete transaction history
