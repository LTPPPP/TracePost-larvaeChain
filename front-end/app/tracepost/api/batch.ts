import { makeAuthenticatedRequest } from "./auth";

const apiUrl = process.env.EXPO_PUBLIC_API_URL;

export interface CreateBatchRequest {
  hatchery_id: number;
  quantity: number;
  species: string;
}

export interface BatchData {
  id: number;
  hatchery_id: number;
  hatchery: {
    id: number;
    name: string;
    company_id: number;
    company: {
      id: number;
      name: string;
      type: string;
      location: string;
      contact_info: string;
      created_at: string;
      updated_at: string;
      is_active: boolean;
    };
    created_at: string;
    updated_at: string;
    is_active: boolean;
  };
  species: string;
  quantity: number;
  status: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export interface CreateBatchResponse {
  success: boolean;
  message: string;
  data: {
    batch: BatchData;
    blockchain: {
      errors: any[];
      success: boolean;
      transaction_ids: string[];
    };
  };
}

export interface GetBatchesResponse {
  success: boolean;
  message: string;
  data: BatchData[];
}

export interface GetBatchResponse {
  success: boolean;
  message: string;
  data: BatchData;
}

export interface BlockchainTransaction {
  tx_id: string;
  type: string;
  timestamp: string;
  payload: any;
  validated_at: string;
  sender?: string;
}

export interface BlockchainState {
  batch_id: string;
  hatchery_id: string;
  quantity: number;
  species: string;
  status: string;
}

export interface BatchBlockchainData {
  batch_id: string;
  first_tx: string;
  latest_tx: string;
  state: BlockchainState;
  tx_count: number;
  txs: BlockchainTransaction[];
}

export interface GetBatchBlockchainResponse {
  success: boolean;
  message: string;
  data: BatchBlockchainData;
}

export interface BatchEvent {
  event_type: string;
  id: number;
  metadata: {
    blockchain_errors: any[];
    blockchain_success: boolean;
  };
  timestamp: string;
}

export interface DbRecord {
  created_at: string;
  id: number;
  metadata_hash: string;
  tx_id: string;
}

export interface BatchHistoryData {
  batch_events: BatchEvent[];
  batch_id: number;
  blockchain_transactions: BlockchainTransaction[];
  db_records: DbRecord[];
  verifiable_history: boolean;
}

export interface GetBatchHistoryResponse {
  success: boolean;
  message: string;
  data: BatchHistoryData;
}

export interface EnvironmentData {
  age: number;
  density: number;
  is_active: boolean;
  ph: number;
  salinity: number;
  temperature: number;
  timestamp: string;
  updated_at: string;
}

export interface BatchEnvironmentRecord {
  id: number;
  batch_info: {
    id: number;
    quantity: number;
    species: string;
    status: string;
  };
  environment_data: EnvironmentData;
  facility_info: {
    company_location: string;
    company_name: string;
    hatchery_name: string;
  };
}

export interface GetBatchEnvironmentResponse {
  success: boolean;
  message: string;
  data: BatchEnvironmentRecord[];
}

export interface CreateEnvironmentRequest {
  age: number;
  batch_id: number;
  density: number;
  ph: number;
  salinity: number;
  temperature: number;
}

export interface CreateEnvironmentResponse {
  success: boolean;
  message: string;
  data: {
    id: number;
    batch_id: number;
    temperature: number;
    ph: number;
    salinity: number;
    density: number;
    age: number;
    timestamp: string;
    updated_at: string;
    is_active: boolean;
  };
}

export interface Event {
  id: number;
  batch_id: number;
  event_type: string;
  location: string;
  metadata: Record<string, any>;
  timestamp: string;
  updated_at: string;
  is_active: boolean;
  facility_info: {
    company_name: string;
    hatchery_name: string;
  };
  batch_info: {
    quantity: number;
    species: string;
    status: string;
  };
}

export interface GetBatchEventsResponse {
  success: boolean;
  message: string;
  data: Event[];
}

export async function createBatch(
  batchData: CreateBatchRequest,
): Promise<CreateBatchResponse> {
  try {
    const response = await makeAuthenticatedRequest(`${apiUrl}/batches`, {
      method: "POST",
      body: JSON.stringify(batchData),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to create batch");
    }

    const data: CreateBatchResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Create Batch API error:", error);
    throw error;
  }
}

export async function getBatchesByHatchery(
  hatcheryId: number,
): Promise<GetBatchesResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${apiUrl}/hatcheries/${hatcheryId}/batches`,
      {
        method: "GET",
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch batches");
    }

    const data: GetBatchesResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Get Batches API error:", error);
    throw error;
  }
}

export async function getAllBatches(): Promise<GetBatchesResponse> {
  try {
    const response = await makeAuthenticatedRequest(`${apiUrl}/batches`, {
      method: "GET",
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch batches");
    }

    const data: GetBatchesResponse = await response.json();

    // Handle null data case
    if (data.success && data.data === null) {
      return {
        success: true,
        message: data.message,
        data: [], // Convert null to empty array
      };
    }

    return data;
  } catch (error) {
    console.error("Get All Batches API error:", error);
    throw error;
  }
}

export async function getBatchById(batchId: number): Promise<GetBatchResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${apiUrl}/batches/${batchId}`,
      {
        method: "GET",
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch batch");
    }

    const data: GetBatchResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Get Batch API error:", error);
    throw error;
  }
}

export async function getBatchBlockchainData(
  batchId: number,
): Promise<GetBatchBlockchainResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${apiUrl}/batches/${batchId}/blockchain`,
      {
        method: "GET",
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch blockchain data");
    }

    const data: GetBatchBlockchainResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Get Batch Blockchain Data API error:", error);
    throw error;
  }
}

export async function getBatchHistory(
  batchId: number,
): Promise<GetBatchHistoryResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${apiUrl}/batches/${batchId}/history`,
      {
        method: "GET",
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch batch history");
    }

    const data: GetBatchHistoryResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Get Batch History API error:", error);
    throw error;
  }
}

export async function getBatchEnvironment(
  batchId: number,
): Promise<GetBatchEnvironmentResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${apiUrl}/batches/${batchId}/environment`,
      {
        method: "GET",
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch environment data");
    }

    const data: GetBatchEnvironmentResponse = await response.json();

    // Handle null data case
    if (data.success && data.data === null) {
      return {
        success: true,
        message: data.message,
        data: [], // Convert null to empty array
      };
    }

    return data;
  } catch (error) {
    console.error("Get Batch Environment API error:", error);
    throw error;
  }
}

export async function createEnvironmentData(
  environmentData: CreateEnvironmentRequest,
): Promise<CreateEnvironmentResponse> {
  try {
    const response = await makeAuthenticatedRequest(`${apiUrl}/environment`, {
      method: "POST",
      body: JSON.stringify(environmentData),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to create environment data");
    }

    const data: CreateEnvironmentResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Create Environment Data API error:", error);
    throw error;
  }
}

export async function getBatchEvents(
  batchId: number,
): Promise<GetBatchEventsResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${process.env.EXPO_PUBLIC_API_URL}/events?batch_id=${batchId}`,
      {
        method: "GET",
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch batch events");
    }

    const data: GetBatchEventsResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Get Batch Events API error:", error);
    throw error;
  }
}
