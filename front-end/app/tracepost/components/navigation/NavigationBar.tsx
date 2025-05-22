import React from "react";
import { View } from "react-native";
import { BottomTabBarProps } from "@react-navigation/bottom-tabs";
import NavigationButton from "./NavigationButton";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { BlurView } from "expo-blur";

export default function NavigationBar({
  state,
  descriptors,
  navigation,
  isVisible = true,
}: BottomTabBarProps & { isVisible?: boolean }) {
  const insets = useSafeAreaInsets();

  if (!isVisible) {
    return null;
  }

  // Map route names to more appropriate Web3 themed icons and labels
  const getIconInfo = (routeName: string) => {
    const routeMap: { [key: string]: { icon: string; label: string } } = {
      "(home)/index": { icon: "chart-dots-2", label: "Info" },
      "(report)/index": { icon: "currency-ethereum", label: "Report" },
      "(track)/index": { icon: "qrcode", label: "Track" },
    };

    return routeMap[routeName] || { icon: "circle", label: routeName };
  };

  return (
    <View
      className="absolute bottom-0 left-0 right-0 z-50"
      style={{ paddingBottom: 8 }}
    >
      <BlurView
        intensity={40}
        tint="dark"
        className="rounded-3xl mx-4 mb-3 overflow-hidden border border-white/15"
      >
        <View className="flex-row justify-around py-2">
          {state.routes.map((route, index) => {
            const { options } = descriptors[route.key];
            const isFocused = state.index === index;
            const { icon, label } = getIconInfo(route.name);

            const onPress = () => {
              const event = navigation.emit({
                type: "tabPress",
                target: route.key,
                canPreventDefault: true,
              });

              if (!isFocused && !event.defaultPrevented) {
                navigation.navigate(route.name, route.params);
              }
            };

            const onLongPress = () => {
              navigation.emit({
                type: "tabLongPress",
                target: route.key,
              });
            };

            return (
              <NavigationButton
                key={route.name}
                onPress={onPress}
                onLongPress={onLongPress}
                isFocused={isFocused}
                icon={icon}
                label={label}
              />
            );
          })}
        </View>
      </BlurView>
    </View>
  );
}
