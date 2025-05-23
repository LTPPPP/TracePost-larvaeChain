import { Stack, useRouter } from "expo-router";
import React, { useEffect } from "react";
import { View } from "react-native";
import * as SplashScreen from "expo-splash-screen";
import { useFonts } from "expo-font";

SplashScreen.preventAutoHideAsync();

export default function RootLayout() {
  const router = useRouter();
  const [fontsLoaded] = useFonts({
    TablerIcons: require("../assets/tabler-icons/tabler-icons.ttf"),
  });

  useEffect(() => {
    const hideSplashAndNavigate = async () => {
      SplashScreen.hideAsync();

      await new Promise((resolve) => setTimeout(resolve, 1000)); // Simulate a delay for splash screen

      router.replace("/(tabs)/(home)");
      // router.replace("/(auth)/login");
    };

    hideSplashAndNavigate();
  }, [fontsLoaded]);

  return (
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
      />
    </View>
  );
}
