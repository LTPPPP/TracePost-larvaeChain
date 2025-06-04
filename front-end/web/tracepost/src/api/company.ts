// API Types File - Improved Version
const endpoint = process.env.NEXT_PUBLIC_API_URL;

interface ApiCompany {
  id: number;
  name: string;
  type: string;
  location: string;
  contact_info: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  hatcheries?: ApiHatchery[]; // Optional property
}

interface ApiHatchery {
  id: number;
  name: string;
  company_id: number;
  company: ApiCompany;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

interface ApiBatch {
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

interface ApiEnvironment {
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

interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T[] | T;
}

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

export async function getListHatcheries(): Promise<ApiResponse<ApiHatchery[]>> {
  try {
    const response = await fetch(`${endpoint}/hatcheries`, {
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
    console.error('Error fetching hatcheries:', error);
    throw error;
  }
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

export function countHatcheriesByCompany(hatcheries: ApiHatchery[], companyId: number): number {
  return hatcheries.filter((hatchery) => hatchery.company_id === companyId).length;
}

export type { ApiCompany, ApiHatchery, ApiResponse, ApiBatch, ApiEnvironment };
