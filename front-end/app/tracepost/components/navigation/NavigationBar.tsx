import React from "react";
import { View } from "react-native";
import { BottomTabBarProps } from "@react-navigation/bottom-tabs";
import NavigationButton from "./NavigationButton";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { BlurView } from "expo-blur";
import { useRole } from "@/contexts/RoleContext";

export default function NavigationBar({
  state,
  descriptors,
  navigation,
  isVisible = true,
}: BottomTabBarProps & { isVisible?: boolean }) {
  const insets = useSafeAreaInsets();
  const { currentRole, isLoading } = useRole();

  // Don't show navigation bar while loading or if no role
  if (!isVisible || isLoading || !currentRole) {
    return null;
  }

  // Map route names to role-specific icons and labels
  const getIconInfo = (routeName: string, role: "user" | "hatchery") => {
    const roleBasedRouteMap: {
      [key in "user" | "hatchery"]: {
        [route: string]: { icon: string; label: string };
      };
    } = {
      hatchery: {
        "(home)/index": { icon: "dashboard", label: "Dashboard" },
        "(hatchery)/index": { icon: "building-factory-2", label: "Hatcheries" },
        "(batches)/index": { icon: "package", label: "Batches" },
      },
      user: {
        "(home)/index": { icon: "chart-dots-2", label: "Dashboard" },
        "(report)/index": { icon: "file-plus", label: "Report" },
        "(track)/index": { icon: "qrcode", label: "Track" },
      },
    };

    return (
      roleBasedRouteMap[role]?.[routeName] || {
        icon: "circle",
        label: "Unknown",
      }
    );
  };

  // Filter routes based on role - using the correct path structure
  const getVisibleRoutes = () => {
    return state.routes.filter((route) => {
      if (currentRole === "hatchery") {
        return ["(home)/index", "(hatchery)/index", "(batches)/index"].includes(
          route.name,
        );
      } else if (currentRole === "user") {
        return ["(home)/index", "(report)/index", "(track)/index"].includes(
          route.name,
        );
      }
      return false;
    });
  };

  const visibleRoutes = getVisibleRoutes();

  // Find the active route index among visible routes
  const getActiveIndex = () => {
    const activeRouteName = state.routes[state.index]?.name;
    return visibleRoutes.findIndex((route) => route.name === activeRouteName);
  };

  const activeIndex = getActiveIndex();

  return (
    <View
      className="absolute bottom-0 left-0 right-0 z-50"
      style={{ paddingBottom: Math.max(insets.bottom, 8) }}
    >
      <BlurView
        intensity={40}
        tint="dark"
        className="rounded-3xl mx-4 mb-3 overflow-hidden border border-white/15"
      >
        <View className="flex-row justify-around py-2">
          {visibleRoutes.map((route, index) => {
            const { options } = descriptors[route.key];
            const isFocused = activeIndex === index;
            const { icon, label } = getIconInfo(route.name, currentRole);

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
