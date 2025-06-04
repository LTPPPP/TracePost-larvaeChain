const endpoint = process.env.NEXT_PUBLIC_API_URL;

// Company types
export interface ApiCompany {
  id: number;
  name: string;
  type: string;
  location: string;
  contact_info: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  hatcheries?: ApiHatchery[];
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T[] | T;
}

import { ApiHatchery } from './hatchery';
import { ApiBatch, ApiEnvironment } from './batch';

export async function getListCompany(): Promise<ApiResponse<ApiCompany[]>> {
  try {
    const response = await fetch(`${endpoint}/companies`, {
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
    console.error('Error fetching companies:', error);
    throw error;
  }
}

export async function getCompanyById(
  companyId: number
): Promise<{ success: boolean; message: string; data: ApiCompany }> {
  try {
    const response = await fetch(`${endpoint}/companies/${companyId}`, {
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
    console.error('Error fetching company by ID:', error);
    throw error;
  }
}

export async function getCompanyBatchesWithEnvironment(companyId: number) {
  try {
    const { getBatches, getEnvironment } = await import('./batch');

    const batchesResponse = await getBatches();

    if (!batchesResponse.success || !batchesResponse.data) {
      throw new Error('Failed to fetch batches');
    }

    const batchesData = Array.isArray(batchesResponse.data) ? batchesResponse.data : [batchesResponse.data];

    const companyBatches = batchesData.filter((batch: ApiBatch) => batch.hatchery?.company_id === companyId);

    const batchesWithEnvironment = await Promise.all(
      companyBatches.map(async (batch: ApiBatch) => {
        try {
          const envResponse = await getEnvironment(batch.id);

          if (envResponse.success && Array.isArray(envResponse.data) && envResponse.data.length > 0) {
            const envData = envResponse.data[0] as ApiEnvironment;

            return {
              id: batch.id.toString(),
              name: envData.facility_info?.hatchery_name || batch.hatchery?.name || 'Unknown Hatchery',
              temperature: envData.temperature ?? 0,
              ph: envData.ph ?? 0,
              salinity: envData.salinity ?? 0,
              density: envData.density ?? 0,
              age: envData.age ?? 0,
              species: batch.species || 'Unknown Species',
              quantity: batch.quantity || 0,
              status: batch.status,
              batchId: batch.id,
              hatcheryId: batch.hatchery_id
            };
          }

          return {
            id: batch.id.toString(),
            name: batch.hatchery?.name || 'Unknown Hatchery',
            temperature: 0,
            ph: 0,
            salinity: 0,
            density: 0,
            age: 0,
            species: batch.species || 'Unknown Species',
            quantity: batch.quantity || 0,
            status: batch.status,
            batchId: batch.id,
            hatcheryId: batch.hatchery_id
          };
        } catch (error) {
          console.error(`Error fetching environment for batch ${batch.id}:`, error);

          return {
            id: batch.id.toString(),
            name: batch.hatchery?.name || 'Unknown Hatchery',
            temperature: 0,
            ph: 0,
            salinity: 0,
            density: 0,
            age: 0,
            species: batch.species || 'Unknown Species',
            quantity: batch.quantity || 0,
            status: batch.status,
            batchId: batch.id,
            hatcheryId: batch.hatchery_id
          };
        }
      })
    );

    return {
      success: true,
      data: batchesWithEnvironment.filter((batch) => batch !== null)
    };
  } catch (error) {
    console.error('Error fetching company batches with environment:', error);
    return {
      success: false,
      data: [],
      error: error instanceof Error ? error.message : 'Unknown error'
    };
  }
}

export async function createHatchery(data: { company_id: number; name: string }) {
  try {
    console.log('Creating hatchery with data:', data);

    const response = await fetch(`${endpoint}/hatcheries`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(data)
    });

    console.log('Hatchery API response status:', response.status);

    const responseText = await response.text();
    console.log('Hatchery API raw response:', responseText);

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}, response: ${responseText}`);
    }

    const result = JSON.parse(responseText);
    console.log('Hatchery API parsed result:', result);

    return result;
  } catch (error) {
    console.error('Error creating hatchery:', error);
    throw error;
  }
}
