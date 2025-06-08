import React from "react";
import { Tabs } from "expo-router";
import { StatusBar, View, ActivityIndicator } from "react-native";
import NavigationBar from "@/components/navigation/NavigationBar";
import { SafeAreaProvider } from "react-native-safe-area-context";
import { useRole } from "@/contexts/RoleContext";

export default function TabLayout() {
  const { currentRole, isLoading } = useRole();

  // Show loading while role is being determined
  if (isLoading) {
    return (
      <SafeAreaProvider>
        <StatusBar barStyle="dark-content" />
        <View className="flex-1 bg-white justify-center items-center">
          <ActivityIndicator size="large" color="#f97316" />
        </View>
      </SafeAreaProvider>
    );
  }

  // If no role is determined, return empty view (should redirect in _layout.tsx)
  if (!currentRole) {
    return (
      <SafeAreaProvider>
        <StatusBar barStyle="dark-content" />
        <View className="flex-1 bg-white" />
      </SafeAreaProvider>
    );
  }

  return (
    <SafeAreaProvider>
      <StatusBar barStyle="dark-content" />
      <Tabs
        screenOptions={{
          headerShown: false,
          tabBarStyle: { display: "none" }, // Hide the default tab bar
        }}
        tabBar={(props) => <NavigationBar {...props} />}
      >
        {/* Home Tab - Available for all roles */}
        <Tabs.Screen
          name="(home)"
          options={{
            title: "Dashboard",
            href: "/(tabs)/(home)",
          }}
        />

        {/* Hatchery Role Tabs */}
        <Tabs.Screen
          name="(hatchery)"
          options={{
            title: "Hatcheries",
            href: currentRole === "hatchery" ? "/(tabs)/(hatchery)" : null,
          }}
        />
        <Tabs.Screen
          name="(batches)"
          options={{
            title: "Batches",
            href: currentRole === "hatchery" ? "/(tabs)/(batches)" : null,
          }}
        />

        {/* Track Tab - Available for both roles */}
        <Tabs.Screen
          name="(track)"
          options={{
            title: "Track",
            href: "/(tabs)/(track)",
          }}
        />

        {/* User Role Tabs */}
        <Tabs.Screen
          name="(report)"
          options={{
            title: "Report",
            href: currentRole === "user" ? "/(tabs)/(report)" : null,
          }}
        />
      </Tabs>
    </SafeAreaProvider>
  );
}
