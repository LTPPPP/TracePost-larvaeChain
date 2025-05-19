# Admin Analytics API Documentation

The Admin Analytics API provides real-time data and insights for system administrators. These endpoints enable monitoring of system performance, user activity, blockchain operations, and compliance metrics.

## Dashboard Analytics

`GET /api/v1/admin/analytics/dashboard`

Retrieves comprehensive analytics for the admin dashboard, combining all analytics categories into a single response.

**Response:**

```json
{
  "success": true,
  "message": "Analytics data retrieved successfully",
  "data": {
    "system": {
      "active_users": 128,
      "total_batches": 1052,
      "blockchain_tx_count": 5283,
      "api_requests_per_hour": 287,
      "avg_response_time_ms": 156.8,
      "system_health": "healthy",
      "server_cpu_usage": 35.5,
      "server_memory_usage": 45.2,
      "db_connections": 8,
      "last_updated": "2023-05-21T15:20:30Z"
    },
    "compliance": {
      "total_certificates": 827,
      "valid_certificates": 723,
      "expired_certificates": 42,
      "revoked_certificates": 62,
      "company_compliance": {
        "Pacific Blue Aquaculture": 92.8,
        "Green Ocean Farms": 87.3,
        "Sustainable Seafood Co.": 94.5
      },
      "standards_compliance": {
        "ASC": 92.3,
        "ISO9001": 88.7,
        "GlobalG.A.P": 85.1,
        "BAP": 90.5
      },
      "regional_compliance": {
        "North Vietnam": 88.2,
        "Central Vietnam": 92.7,
        "South Vietnam": 85.3
      },
      "last_updated": "2023-05-21T15:18:42Z"
    },
    "blockchain": {
      "total_nodes": 5,
      "active_nodes": 5,
      "network_health": "healthy",
      "consensus_status": "running",
      "average_block_time_ms": 2500,
      "transactions_per_second": 15.7,
      "pending_transactions": 23,
      "node_latencies": {
        "node-1": 45,
        "node-2": 62,
        "node-3": 38,
        "node-4": 72,
        "node-5": 55
      },
      "chain_health": {
        "tracepost-main": "healthy",
        "cosmos-ibc": "healthy",
        "polkadot": "syncing"
      },
      "cross_chain_transactions": {
        "tracepost-main->cosmos-ibc": 137,
        "cosmos-ibc->tracepost-main": 92,
        "tracepost-main->polkadot": 43,
        "polkadot->tracepost-main": 38
      },
      "last_updated": "2023-05-21T15:19:15Z"
    },
    "user_activity": {
      "active_users_by_role": {
        "admin": 3,
        "hatchery_manager": 24,
        "inspector": 17,
        "logistics": 32,
        "viewer": 52
      },
      "login_frequency": {
        "today": 128,
        "yesterday": 115,
        "last_7_days": 742,
        "last_30_days": 2857
      },
      "api_endpoint_usage": {
        "/api/v1/batches": 1257,
        "/api/v1/hatcheries": 892,
        "/api/v1/documents": 745,
        "/api/v1/auth/login": 684,
        "/api/v1/blockchain/batch": 523
      },
      "most_active_users": [
        {
          "user_id": 1,
          "username": "admin",
          "request_count": 248,
          "last_active": "2023-05-21T14:15:30Z"
        },
        {
          "user_id": 5,
          "username": "farm_manager1",
          "request_count": 187,
          "last_active": "2023-05-21T14:45:30Z"
        }
      ],
      "user_growth": {
        "today": 3,
        "yesterday": 2,
        "last_7_days": 15,
        "last_30_days": 42,
        "last_90_days": 98
      },
      "last_updated": "2023-05-21T15:20:00Z"
    },
    "batch": {
      "total_batches_produced": 1052,
      "active_batches": 287,
      "batches_by_status": {
        "active": 287,
        "shipped": 142,
        "completed": 578,
        "rejected": 45
      },
      "batches_by_region": {
        "North Vietnam": 287,
        "Central Vietnam": 452,
        "South Vietnam": 339
      },
      "batches_by_species": {
        "Litopenaeus vannamei": 725,
        "Penaeus monodon": 287,
        "Macrobrachium rosenbergii": 40
      },
      "batches_by_hatchery": {
        "Pacific Blue Aquaculture": 312,
        "Green Ocean Farms": 287,
        "Sustainable Seafood Co.": 265,
        "Vietnam Aquatic Industries": 188
      },
      "production_trend": {
        "last_6_months": [125, 142, 160, 155, 172, 183]
      },
      "average_shipment_time": {
        "North to Central": 5.2,
        "Central to South": 4.8,
        "North to South": 10.5,
        "Farm to Processing": 3.2,
        "Processing to Distribution": 6.7
      },
      "last_updated": "2023-05-21T15:17:25Z"
    },
    "timestamp": "2023-05-21T15:20:30Z"
  }
}
```

## System Metrics

`GET /api/v1/admin/analytics/system`

Retrieves system performance metrics including user counts, batch counts, API performance metrics, and server health data.

**Response:**

```json
{
  "success": true,
  "message": "System metrics retrieved successfully",
  "data": {
    "active_users": 128,
    "total_batches": 1052,
    "blockchain_tx_count": 5283,
    "api_requests_per_hour": 287,
    "avg_response_time_ms": 156.8,
    "system_health": "healthy",
    "server_cpu_usage": 35.5,
    "server_memory_usage": 45.2,
    "db_connections": 8,
    "last_updated": "2023-05-21T15:20:30Z"
  }
}
```

## Blockchain Analytics

`GET /api/v1/admin/analytics/blockchain`

Retrieves blockchain network performance metrics, including node status, transaction throughput, and cross-chain transaction counts.

**Response:**

```json
{
  "success": true,
  "message": "Blockchain analytics retrieved successfully",
  "data": {
    "total_nodes": 5,
    "active_nodes": 5,
    "network_health": "healthy",
    "consensus_status": "running",
    "average_block_time_ms": 2500,
    "transactions_per_second": 15.7,
    "pending_transactions": 23,
    "node_latencies": {
      "node-1": 45,
      "node-2": 62,
      "node-3": 38,
      "node-4": 72,
      "node-5": 55
    },
    "chain_health": {
      "tracepost-main": "healthy",
      "cosmos-ibc": "healthy",
      "polkadot": "syncing"
    },
    "cross_chain_transactions": {
      "tracepost-main->cosmos-ibc": 137,
      "cosmos-ibc->tracepost-main": 92,
      "tracepost-main->polkadot": 43,
      "polkadot->tracepost-main": 38
    },
    "last_updated": "2023-05-21T15:19:15Z"
  }
}
```

## Compliance Analytics

`GET /api/v1/admin/analytics/compliance`

Retrieves compliance metrics, including certificates status, company compliance percentages, and compliance trends over time.

**Response:**

```json
{
  "success": true,
  "message": "Compliance analytics retrieved successfully",
  "data": {
    "total_certificates": 827,
    "valid_certificates": 723,
    "expired_certificates": 42,
    "revoked_certificates": 62,
    "company_compliance": {
      "Pacific Blue Aquaculture": 92.8,
      "Green Ocean Farms": 87.3,
      "Sustainable Seafood Co.": 94.5
    },
    "standards_compliance": {
      "ASC": 92.3,
      "ISO9001": 88.7,
      "GlobalG.A.P": 85.1,
      "BAP": 90.5
    },
    "regional_compliance": {
      "North Vietnam": 88.2,
      "Central Vietnam": 92.7,
      "South Vietnam": 85.3
    },
    "compliance_trends": {
      "last_6_months": [83.2, 84.5, 86.1, 87.2, 88.5, 89.3]
    },
    "last_updated": "2023-05-21T15:18:42Z"
  }
}
```

## User Activity Analytics

`GET /api/v1/admin/analytics/users`

Retrieves user activity metrics, including active users by role, login frequency, and API usage patterns.

**Response:**

```json
{
  "success": true,
  "message": "User activity analytics retrieved successfully",
  "data": {
    "active_users_by_role": {
      "admin": 3,
      "hatchery_manager": 24,
      "inspector": 17,
      "logistics": 32,
      "viewer": 52
    },
    "login_frequency": {
      "today": 128,
      "yesterday": 115,
      "last_7_days": 742,
      "last_30_days": 2857
    },
    "api_endpoint_usage": {
      "/api/v1/batches": 1257,
      "/api/v1/hatcheries": 892,
      "/api/v1/documents": 745,
      "/api/v1/auth/login": 684,
      "/api/v1/blockchain/batch": 523
    },
    "most_active_users": [
      {
        "user_id": 1,
        "username": "admin",
        "request_count": 248,
        "last_active": "2023-05-21T14:15:30Z"
      },
      {
        "user_id": 5,
        "username": "farm_manager1",
        "request_count": 187,
        "last_active": "2023-05-21T14:45:30Z"
      }
    ],
    "user_growth": {
      "today": 3,
      "yesterday": 2,
      "last_7_days": 15,
      "last_30_days": 42,
      "last_90_days": 98
    },
    "last_updated": "2023-05-21T15:20:00Z"
  }
}
```

## Batch Analytics

`GET /api/v1/admin/analytics/batches`

Retrieves batch production and tracking metrics, including status counts, regional distribution, and shipment timing data.

**Response:**

```json
{
  "success": true,
  "message": "Batch analytics retrieved successfully",
  "data": {
    "total_batches_produced": 1052,
    "active_batches": 287,
    "batches_by_status": {
      "active": 287,
      "shipped": 142,
      "completed": 578,
      "rejected": 45
    },
    "batches_by_region": {
      "North Vietnam": 287,
      "Central Vietnam": 452,
      "South Vietnam": 339
    },
    "batches_by_species": {
      "Litopenaeus vannamei": 725,
      "Penaeus monodon": 287,
      "Macrobrachium rosenbergii": 40
    },
    "batches_by_hatchery": {
      "Pacific Blue Aquaculture": 312,
      "Green Ocean Farms": 287,
      "Sustainable Seafood Co.": 265,
      "Vietnam Aquatic Industries": 188
    },
    "production_trend": {
      "last_6_months": [125, 142, 160, 155, 172, 183]
    },
    "average_shipment_time": {
      "North to Central": 5.2,
      "Central to South": 4.8,
      "North to South": 10.5,
      "Farm to Processing": 3.2,
      "Processing to Distribution": 6.7
    },
    "last_updated": "2023-05-21T15:17:25Z"
  }
}
```

## Export Analytics Data

`GET /api/v1/admin/analytics/export`

Exports all analytics data as a downloadable JSON file. The response will be a file download rather than a JSON response.

**Response:**

A downloadable JSON file containing all analytics data combined from the various endpoints.

## Refresh Analytics Data

`POST /api/v1/admin/analytics/refresh`

Forces a refresh of all analytics data. This is useful when immediate updates are needed before the scheduled refresh.

**Response:**

```json
{
  "success": true,
  "message": "Analytics data refresh triggered",
  "data": {
    "triggered_at": "2023-05-21T15:25:30Z",
    "status": "processing"
  }
}
```

## Implementation Notes

- Analytics data is automatically refreshed every 5 minutes
- Historical analytics data is retained for 90 days by default
- The API supports filtering by date ranges (use query parameters `start_date` and `end_date` in ISO 8601 format)
- For real-time monitoring, consider using WebSocket connections instead of polling these endpoints
