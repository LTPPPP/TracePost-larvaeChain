import { makeAuthenticatedRequest } from "./auth";

const apiUrl = process.env.EXPO_PUBLIC_API_URL;

export interface CreateHatcheryRequest {
  company_id: number;
  name: string;
}

export interface CreateHatcheryResponse {
  success: boolean;
  message: string;
  data: {
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
}

export interface GetHatcheriesResponse {
  success: boolean;
  message: string;
  data: Array<{
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
  }>;
}

export interface GetHatcheryResponse {
  success: boolean;
  message: string;
  data: {
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
}

export interface UpdateHatcheryRequest {
  name: string;
}

export interface UpdateHatcheryResponse {
  success: boolean;
  message: string;
  data: {
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
}

export async function createHatchery(
  hatcheryData: CreateHatcheryRequest,
): Promise<CreateHatcheryResponse> {
  try {
    const response = await makeAuthenticatedRequest(`${apiUrl}/hatcheries`, {
      method: "POST",
      body: JSON.stringify(hatcheryData),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to create hatchery");
    }

    const data: CreateHatcheryResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Create Hatchery API error:", error);
    throw error;
  }
}

export async function getHatcheries(): Promise<GetHatcheriesResponse> {
  try {
    const response = await makeAuthenticatedRequest(`${apiUrl}/hatcheries`, {
      method: "GET",
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch hatcheries");
    }

    const data: GetHatcheriesResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Get Hatcheries API error:", error);
    throw error;
  }
}

export async function getHatcheryById(
  hatcheryId: number,
): Promise<GetHatcheryResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${apiUrl}/hatcheries/${hatcheryId}`,
      {
        method: "GET",
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to fetch hatchery");
    }

    const data: GetHatcheryResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Get Hatchery API error:", error);
    throw error;
  }
}

export async function updateHatchery(
  hatcheryId: number,
  updateData: UpdateHatcheryRequest,
): Promise<UpdateHatcheryResponse> {
  try {
    const response = await makeAuthenticatedRequest(
      `${apiUrl}/hatcheries/${hatcheryId}`,
      {
        method: "PUT",
        body: JSON.stringify(updateData),
      },
    );

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.message || "Failed to update hatchery");
    }

    const data: UpdateHatcheryResponse = await response.json();
    return data;
  } catch (error) {
    console.error("Update Hatchery API error:", error);
    throw error;
  }
}
