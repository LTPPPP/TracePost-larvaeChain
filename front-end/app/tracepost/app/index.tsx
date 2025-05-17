import { Text, View } from "react-native";
import "../global.css"

export default function Index() {
  return (
    <View
      style={{
        flex: 1,
        justifyContent: "center",
        alignItems: "center",
      }}
    >
      <Text className="text-blue-400">Con tom vui ve, tracepost vui ve</Text>
    </View>
  );
}
