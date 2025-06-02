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

interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T[];
}

export async function getListCompany(): Promise<ApiResponse<ApiCompany>> {
  try {
    const response = await fetch(`${endpoint}/companies`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
        // Thêm authorization header nếu cần
        // 'Authorization': `Bearer ${token}`
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

export async function getListHatcheries(): Promise<ApiResponse<ApiHatchery>> {
  try {
    const response = await fetch(`${endpoint}/hatcheries`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
        // Thêm authorization header nếu cần
        // 'Authorization': `Bearer ${token}`
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

export type { ApiCompany, ApiHatchery, ApiResponse };
