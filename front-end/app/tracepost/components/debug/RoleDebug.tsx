import React from "react";
import { View, Text } from "react-native";
import { useRole } from "@/contexts/RoleContext";

export default function RoleDebug() {
  const { currentRole, userData, isLoading } = useRole();

  if (__DEV__) {
    return (
      <View className="bg-yellow-100 p-2 m-2 rounded">
        <Text className="text-xs">Role: {currentRole || "null"}</Text>
        <Text className="text-xs">Loading: {isLoading.toString()}</Text>
        <Text className="text-xs">User: {userData?.username || "null"}</Text>
      </View>
    );
  }

  return null;
}
