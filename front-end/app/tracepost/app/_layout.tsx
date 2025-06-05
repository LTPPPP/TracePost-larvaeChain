import { Stack, useRouter } from "expo-router";
import React, { useEffect } from "react";
import { View } from "react-native";
import * as SplashScreen from "expo-splash-screen";
import { useFonts } from "expo-font";
import { StorageService } from "@/utils/storage";
import { RoleProvider } from "@/contexts/RoleContext";

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const router = useRouter();
  const [fontsLoaded] = useFonts({
    TablerIcons: require("../assets/tabler-icons/tabler-icons.ttf"),
  });

  useEffect(() => {
    const initializeApp = async () => {
      if (!fontsLoaded) return;

      try {
        // Check if user is logged in
        const isLoggedIn = await StorageService.isLoggedIn();

        // Hide splash screen
        await SplashScreen.hideAsync();

        // Add a small delay for better UX
        await new Promise((resolve) => setTimeout(resolve, 500));

        if (isLoggedIn) {
          // Get user role to determine navigation
          const userRole = await StorageService.getUserRole();

          if (userRole) {
            // Navigate based on user role
            switch (userRole) {
              case "hatchery":
                router.replace("/(tabs)/(home)");
                break;
              case "user":
                router.replace("/(tabs)/(home)");
                break;
              default:
                // Unknown role, go to login
                router.replace("/(auth)/login");
                break;
            }
          } else {
            // No role found, clear storage and go to login
            await StorageService.clearLoginData();
            router.replace("/(auth)/login");
          }
        } else {
          // Not logged in, go to login
          router.replace("/(auth)/login");
        }
      } catch (error) {
        console.error("App initialization error:", error);
        // On error, clear any corrupted data and go to login
        await StorageService.clearLoginData();
        router.replace("/(auth)/login");
      }
    };

    initializeApp();
  }, [fontsLoaded, router]);

  // Don't render anything until fonts are loaded and navigation is handled
  if (!fontsLoaded) {
    return null;
  }

  return (
    <RoleProvider>
      <View className="flex-1 bg-white">
        <Stack
          screenOptions={{
            headerShown: false,
            animation: "slide_from_bottom",
            animationDuration: 500,
            contentStyle: {
              backgroundColor: "#FFFFFF",
            },
          }}
        >
          {/* Auth Screens */}
          <Stack.Screen
            name="(auth)"
            options={{
              headerShown: false,
              gestureEnabled: false,
            }}
          />

          {/* Main App Screens */}
          <Stack.Screen
            name="(tabs)"
            options={{
              headerShown: false,
              gestureEnabled: false,
            }}
          />

          {/* Modal Screens */}
          <Stack.Screen
            name="hatchery/create"
            options={{
              presentation: "modal",
              headerShown: false,
              animation: "slide_from_bottom",
            }}
          />

          <Stack.Screen
            name="batch/create"
            options={{
              presentation: "modal",
              headerShown: false,
              animation: "slide_from_bottom",
            }}
          />

          {/* Detail Screens */}
          <Stack.Screen
            name="hatchery/[id]"
            options={{
              headerShown: false,
              animation: "slide_from_right",
            }}
          />

          <Stack.Screen
            name="batch/[id]"
            options={{
              headerShown: false,
              animation: "slide_from_right",
            }}
          />
        </Stack>
      </View>
    </RoleProvider>
  );
}
