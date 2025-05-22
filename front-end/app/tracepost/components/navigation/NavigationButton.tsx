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
};

export default function NavigationButton({
  onPress,
  onLongPress,
  isFocused,
  icon,
  label,
}: NavigationButtonProps) {
  // Define different styles based on focused state
  return (
    <Pressable
      onPress={onPress}
      onLongPress={onLongPress}
      style={styles.button}
    >
      {isFocused ? (
        <View style={styles.focusedContainer}>
          <LinearGradient
            colors={["#4338ca", "#6366f1"]} // indigo gradient
            start={{ x: 0, y: 0 }}
            end={{ x: 1, y: 1 }}
            style={styles.gradient}
          >
            <TablerIconComponent name={icon} size={24} color="#ffffff" />
          </LinearGradient>
          <Text style={styles.focusedLabel}>{label}</Text>
          <View style={styles.indicator} />
        </View>
      ) : (
        <View style={styles.container}>
          <View style={styles.iconContainer}>
            <TablerIconComponent
              name={icon}
              size={24}
              color="rgba(255,255,255,0.8)"
            />
          </View>
          <Text style={styles.label}>{label}</Text>
        </View>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  button: {
    paddingVertical: 8,
    paddingHorizontal: 20,
  },
  container: {
    alignItems: "center",
    justifyContent: "center",
  },
  focusedContainer: {
    alignItems: "center",
    justifyContent: "center",
  },
  iconContainer: {
    width: 48,
    height: 48,
    alignItems: "center",
    justifyContent: "center",
    borderRadius: 16,
    backgroundColor: "rgba(255,255,255,0.1)",
  },
  gradient: {
    width: 48,
    height: 48,
    alignItems: "center",
    justifyContent: "center",
    borderRadius: 16,
    shadowColor: "#4338ca",
    shadowOffset: {
      width: 0,
      height: 4,
    },
    shadowOpacity: 0.3,
    shadowRadius: 6,
    elevation: 8,
  },
  label: {
    marginTop: 6,
    fontSize: 12,
    color: "rgba(255,255,255,0.6)",
  },
  focusedLabel: {
    marginTop: 6,
    fontSize: 12,
    fontWeight: "bold",
    color: "#ffffff",
  },
  indicator: {
    width: 12,
    height: 2,
    backgroundColor: "#6366f1",
    borderRadius: 1,
    marginTop: 4,
  },
});
