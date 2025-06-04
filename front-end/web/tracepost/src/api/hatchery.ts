const endpoint = process.env.NEXT_PUBLIC_API_URL;

// Hatchery types
import { ApiCompany } from './company';

export interface ApiHatchery {
  id: number;
  name: string;
  company_id: number;
  company: ApiCompany;
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T[] | T;
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

export function countHatcheriesByCompany(hatcheries: ApiHatchery[], companyId: number): number {
  return hatcheries.filter((hatchery) => hatchery.company_id === companyId).length;
}
