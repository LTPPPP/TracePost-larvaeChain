import AsyncStorage from "@react-native-async-storage/async-storage";

const STORAGE_KEYS = {
  ACCESS_TOKEN: "access_token",
  USER_DATA: "user_data",
  TOKEN_EXPIRY: "token_expiry",
} as const;

export interface UserData {
  user_id: number;
  username: string;
  role: "user" | "hatchery";
  company_id?: number;
}

export interface LoginResponse {
  success: boolean;
  message: string;
  data: {
    access_token: string;
    token_type: string;
    expires_in: number;
    user_id: number;
    role: "user" | "hatchery";
    username?: string;
    company_id?: number;
  };
}

// Helper function to decode JWT token
function decodeJWT(token: string): any {
  try {
    const base64Url = token.split(".")[1];
    const base64 = base64Url.replace(/-/g, "+").replace(/_/g, "/");
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split("")
        .map(function (c) {
          return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2);
        })
        .join(""),
    );
    return JSON.parse(jsonPayload);
  } catch (error) {
    console.error("Error decoding JWT:", error);
    return null;
  }
}

export const StorageService = {
  // Store login data
  async storeLoginData(loginResponse: LoginResponse): Promise<void> {
    try {
      const { access_token, expires_in, user_id, role } = loginResponse.data;

      // Calculate expiry timestamp
      const expiryTime = Date.now() + expires_in * 1000;

      // Store token
      await AsyncStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, access_token);

      // Store expiry time
      await AsyncStorage.setItem(
        STORAGE_KEYS.TOKEN_EXPIRY,
        expiryTime.toString(),
      );

      // Decode JWT to get additional user info
      const decodedToken = decodeJWT(access_token);
      console.log("🔓 Decoded JWT token:", decodedToken);

      // Store user data with role information from JWT + API response
      const userData: UserData = {
        user_id: user_id,
        role: role,
        username: decodedToken?.username || loginResponse.data.username || "",
        company_id: decodedToken?.company_id || loginResponse.data.company_id,
      };

      console.log("💾 Storing user data:", userData);

      await AsyncStorage.setItem(
        STORAGE_KEYS.USER_DATA,
        JSON.stringify(userData),
      );
    } catch (error) {
      console.error("Error storing login data:", error);
      throw new Error("Failed to store login data");
    }
  },

  // Get access token
  async getAccessToken(): Promise<string | null> {
    try {
      const token = await AsyncStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN);

      // Check if token is expired
      if (token && (await this.isTokenExpired())) {
        await this.clearLoginData();
        return null;
      }

      return token;
    } catch (error) {
      console.error("Error getting access token:", error);
      return null;
    }
  },

  // Get user data
  async getUserData(): Promise<UserData | null> {
    try {
      const userData = await AsyncStorage.getItem(STORAGE_KEYS.USER_DATA);
      const result = userData ? JSON.parse(userData) : null;
      console.log("📖 Retrieved user data from storage:", result);
      return result;
    } catch (error) {
      console.error("Error getting user data:", error);
      return null;
    }
  },

  // Get user role
  async getUserRole(): Promise<"user" | "hatchery" | null> {
    try {
      const userData = await this.getUserData();
      const role = userData?.role || null;
      console.log("🏷️ Retrieved user role:", role);
      return role;
    } catch (error) {
      console.error("Error getting user role:", error);
      return null;
    }
  },

  // Check if user has specific role
  async hasRole(role: "user" | "hatchery"): Promise<boolean> {
    try {
      const userRole = await this.getUserRole();
      return userRole === role;
    } catch (error) {
      console.error("Error checking user role:", error);
      return false;
    }
  },

  // Check if user is hatchery
  async isHatchery(): Promise<boolean> {
    return this.hasRole("hatchery");
  },

  // Check if user is regular user
  async isUser(): Promise<boolean> {
    return this.hasRole("user");
  },

  // Check if token is expired
  async isTokenExpired(): Promise<boolean> {
    try {
      const expiryTime = await AsyncStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRY);
      if (!expiryTime) return true;

      return Date.now() > parseInt(expiryTime);
    } catch (error) {
      console.error("Error checking token expiry:", error);
      return true;
    }
  },

  // Check if user is logged in
  async isLoggedIn(): Promise<boolean> {
    try {
      const token = await this.getAccessToken();
      return !!token;
    } catch (error) {
      console.error("Error checking login status:", error);
      return false;
    }
  },

  // Get company ID for the user
  async getCompanyId(): Promise<number | null> {
    try {
      const userData = await this.getUserData();
      return userData?.company_id || null;
    } catch (error) {
      console.error("Error getting company ID:", error);
      return null;
    }
  },

  // Clear all login data
  async clearLoginData(): Promise<void> {
    try {
      await AsyncStorage.multiRemove([
        STORAGE_KEYS.ACCESS_TOKEN,
        STORAGE_KEYS.USER_DATA,
        STORAGE_KEYS.TOKEN_EXPIRY,
      ]);
      console.log("🧹 Cleared all login data");
    } catch (error) {
      console.error("Error clearing login data:", error);
    }
  },
};
