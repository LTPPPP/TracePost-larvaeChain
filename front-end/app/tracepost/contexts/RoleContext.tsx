import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  ReactNode,
} from "react";
import { StorageService, UserData } from "@/utils/storage";

export type UserRole = "user" | "hatchery";

interface RoleContextType {
  // Role state
  currentRole: UserRole | null;
  userData: UserData | null;
  isLoading: boolean;

  // Role checking helpers
  isHatchery: boolean;
  isUser: boolean;

  // Methods
  checkRole: () => Promise<void>;
  hasPermission: (requiredRole: UserRole) => boolean;
  getCompanyId: () => number | null;
  refreshUserData: () => Promise<void>;
}

const RoleContext = createContext<RoleContextType | undefined>(undefined);

interface RoleProviderProps {
  children: ReactNode;
}

export function RoleProvider({ children }: RoleProviderProps) {
  const [currentRole, setCurrentRole] = useState<UserRole | null>(null);
  const [userData, setUserData] = useState<UserData | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Computed properties
  const isHatchery = currentRole === "hatchery";
  const isUser = currentRole === "user";

  // Check and set role from storage
  const checkRole = async () => {
    try {
      console.log("🔍 Checking role from storage...");
      setIsLoading(true);

      // First check if user is logged in
      const isLoggedIn = await StorageService.isLoggedIn();
      if (!isLoggedIn) {
        console.log("❌ User not logged in");
        setCurrentRole(null);
        setUserData(null);
        setIsLoading(false);
        return;
      }

      const user = await StorageService.getUserData();
      console.log("👤 Retrieved user data:", user);

      if (user && user.role) {
        setCurrentRole(user.role);
        setUserData(user);
        console.log("✅ Role set to:", user.role);
      } else {
        console.log("❌ No valid user data found");
        setCurrentRole(null);
        setUserData(null);
      }
    } catch (error) {
      console.error("❌ Error checking role:", error);
      setCurrentRole(null);
      setUserData(null);
    } finally {
      setIsLoading(false);
    }
  };

  // Refresh user data from storage
  const refreshUserData = async () => {
    try {
      const user = await StorageService.getUserData();
      setUserData(user);
      setCurrentRole(user?.role || null);
    } catch (error) {
      console.error("Error refreshing user data:", error);
    }
  };

  // Check if user has required permission
  const hasPermission = (requiredRole: UserRole): boolean => {
    return currentRole === requiredRole;
  };

  // Get company ID
  const getCompanyId = (): number | null => {
    return userData?.company_id || null;
  };

  // Initialize role on mount
  useEffect(() => {
    checkRole();
  }, []);

  const value: RoleContextType = {
    // State
    currentRole,
    userData,
    isLoading,

    // Computed
    isHatchery,
    isUser,

    // Methods
    checkRole,
    hasPermission,
    getCompanyId,
    refreshUserData,
  };

  return <RoleContext.Provider value={value}>{children}</RoleContext.Provider>;
}

// Custom hook to use role context
export function useRole() {
  const context = useContext(RoleContext);
  if (context === undefined) {
    throw new Error("useRole must be used within a RoleProvider");
  }
  return context;
}

// Higher-order component for role-based access control
export function withRoleAccess<T extends object>(
  Component: React.ComponentType<T>,
  requiredRole: UserRole,
) {
  return function RoleProtectedComponent(props: T) {
    const { hasPermission, isLoading } = useRole();

    if (isLoading) {
      return null; // or a loading component
    }

    if (!hasPermission(requiredRole)) {
      return null; // or an unauthorized component
    }

    return <Component {...props} />;
  };
}

// Hook for role-based conditional rendering
export function useRolePermissions() {
  const { isHatchery, isUser, hasPermission, currentRole } = useRole();

  return {
    // Role checks
    isHatchery,
    isUser,
    currentRole,

    // Permission checker
    hasPermission,

    // Conditional rendering helpers
    renderForHatchery: (component: React.ReactNode) =>
      isHatchery ? component : null,

    renderForUser: (component: React.ReactNode) => (isUser ? component : null),

    renderForRole: (role: UserRole, component: React.ReactNode) =>
      hasPermission(role) ? component : null,
  };
}
