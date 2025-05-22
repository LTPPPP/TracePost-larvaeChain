import React from "react";
import { Tabs } from "expo-router";
import { StatusBar, View } from "react-native";
import NavigationBar from "@/components/navigation/NavigationBar";
import { SafeAreaProvider } from "react-native-safe-area-context";

export default function TabLayout() {
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
        <Tabs.Screen
          name="(home)"
          options={{
            title: "Home",
          }}
        />
        <Tabs.Screen
          name="(report)"
          options={{
            title: "Report",
          }}
        />
        <Tabs.Screen
          name="(track)"
          options={{
            title: "Track",
          }}
        />
      </Tabs>
    </SafeAreaProvider>
  );
}
