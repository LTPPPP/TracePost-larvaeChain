import React from "react";
import { Text, Pressable, View, StyleSheet } from "react-native";
import TablerIconComponent from "@/components/icon";
import { LinearGradient } from "expo-linear-gradient";

type NavigationButtonProps = {
  onPress: () => void;
  onLongPress: () => void;
  isFocused: boolean;
  icon: string;
  label: string;
  compactMode?: boolean;
  maxWidth?: number;
};

export default function NavigationButton({
  onPress,
  onLongPress,
  isFocused,
  icon,
  label,
  compactMode = false,
  maxWidth,
}: NavigationButtonProps) {
  // Determine sizes based on compact mode
  const iconSize = 24;
  const iconContainerSize = compactMode ? 44 : 48;
  const fontSize = compactMode ? 10 : 12;
  const verticalPadding = 12;
  const horizontalPadding = compactMode ? 12 : 20;

  // Create dynamic styles
  const dynamicStyles = StyleSheet.create({
    button: {
      paddingVertical: verticalPadding,
      paddingHorizontal: horizontalPadding,
      maxWidth: maxWidth,
      minWidth: compactMode ? 60 : 80,
    },
    iconContainer: {
      width: iconContainerSize,
      height: iconContainerSize,
      alignItems: "center",
      justifyContent: "center",
      borderRadius: compactMode ? 12 : 16,
      backgroundColor: "rgba(255,255,255,0.1)",
    },
    gradient: {
      width: iconContainerSize,
      height: iconContainerSize,
      alignItems: "center",
      justifyContent: "center",
      borderRadius: compactMode ? 12 : 16,
      shadowColor: "#4338ca",
      shadowOffset: {
        width: 0,
        height: compactMode ? 2 : 4,
      },
      shadowOpacity: 0.3,
      shadowRadius: compactMode ? 3 : 6,
      elevation: compactMode ? 4 : 8,
    },
    label: {
      marginTop: compactMode ? 3 : 6,
      fontSize: fontSize,
      color: "rgba(255,255,255,0.6)",
      textAlign: "center",
    },
    focusedLabel: {
      marginTop: compactMode ? 3 : 6,
      fontSize: fontSize,
      fontWeight: "bold",
      color: "#ffffff",
      textAlign: "center",
    },
    indicator: {
      width: compactMode ? 8 : 12,
      height: 2,
      backgroundColor: "#6366f1",
      borderRadius: 1,
      marginTop: compactMode ? 2 : 4,
    },
  });

  return (
    <Pressable
      onPress={onPress}
      onLongPress={onLongPress}
      style={dynamicStyles.button}
    >
      {isFocused ? (
        <View style={styles.focusedContainer}>
          <LinearGradient
            colors={["#4338ca", "#6366f1"]} // indigo gradient
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
            style={dynamicStyles.gradient}
          >
            <TablerIconComponent name={icon} size={iconSize} color="#ffffff" />
          </LinearGradient>
          <Text
            style={dynamicStyles.focusedLabel}
            numberOfLines={1}
            adjustsFontSizeToFit={compactMode}
          >
            {label}
          </Text>
          <View style={dynamicStyles.indicator} />
        </View>
      ) : (
        <View style={styles.container}>
          <View style={dynamicStyles.iconContainer}>
            <TablerIconComponent
              name={icon}
              size={iconSize}
              color="rgba(255,255,255,0.8)"
            />
          </View>
          <Text
            style={dynamicStyles.label}
            numberOfLines={1}
            adjustsFontSizeToFit={compactMode}
          >
            {label}
          </Text>
        </View>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: "center",
    justifyContent: "center",
  },
  focusedContainer: {
    alignItems: "center",
    justifyContent: "center",
  },
});
