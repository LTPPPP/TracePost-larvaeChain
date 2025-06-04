const endpoint = process.env.NEXT_PUBLIC_API_URL;

// Batch types
import { ApiCompany } from './company';

export interface ApiBatch {
  id: number;
  hatchery_id: number;
  hatchery: {
    id: number;
    name: string;
    company_id: number;
    company: ApiCompany;
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

export interface ApiEnvironment {
  id: number;
  age?: number;
  batch_id: number;
  batch_info: {
    quantity: number;
    species: string;
    status: string;
  };
  density?: number;
  facility_info?: {
    company_name?: string;
    hatchery_name?: string;
  };
  is_active: boolean;
  ph?: number;
  salinity?: number;
  temperature?: number;
  timestamp: string;
  updated_at: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T[] | T;
}

export async function getBatches(): Promise<ApiResponse<ApiBatch[]>> {
  try {
    const response = await fetch(`${endpoint}/batches`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching batches:', error);
    throw error;
  }
}

export async function getBatchesByCompanyId(companyId: number): Promise<ApiResponse<ApiBatch[]>> {
  try {
    const response = await fetch(`${endpoint}/batches?company_id=${companyId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching batches by company ID:', error);
    throw error;
  }
}

export async function getEnvironment(batchId: number): Promise<ApiResponse<ApiEnvironment[]>> {
  try {
    const response = await fetch(`${endpoint}/environment?batch_id=${batchId}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;
  } catch (error) {
    console.error('Error fetching environment:', error);
    throw error;
  }
}

export async function createBatch(data: { hatchery_id: number; quantity: number; species: string }) {
  try {
    console.log('Creating batch with data:', data);

    const response = await fetch(`${endpoint}/batches`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });

    console.log('Batch API response status:', response.status);
    console.log('Batch API response headers:', Object.fromEntries(response.headers.entries()));

    const responseText = await response.text();
    console.log('Batch API raw response:', responseText);

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}, response: ${responseText}`);
    }

    const result = JSON.parse(responseText);
    console.log('Batch API parsed result:', result);

    return result;
  } catch (error) {
    console.error('Error creating batch:', error);
    throw error;
  }
}

export async function createEnvironment(data: {
  age: number;
  batch_id: number;
  density: number;
  ph: number;
  salinity: number;
  temperature: number;
}) {
  try {
    console.log('Creating environment with data:', data);

    const response = await fetch(`${endpoint}/environment`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });

    console.log('Environment API response status:', response.status);
    console.log('Environment API response headers:', Object.fromEntries(response.headers.entries()));

    const responseText = await response.text();
    console.log('Environment API raw response:', responseText);

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}, response: ${responseText}`);
    }

    const result = JSON.parse(responseText);
    console.log('Environment API parsed result:', result);

    return result;
  } catch (error) {
    console.error('Error creating environment:', error);
    throw error;
  }
}
