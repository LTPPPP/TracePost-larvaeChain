import React, { ReactNode } from "react";
import { View } from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";

type ScreenWithTabBarProps = {
  children: ReactNode;
  className?: string;
};

export default function ScreenWithTabBar({
  children,
  className = "",
}: ScreenWithTabBarProps) {
  const insets = useSafeAreaInsets();

  return (
    <View className={`flex-1 ${className}`} style={{ paddingBottom: 70 }}>
      {children}
    </View>
  );
}
