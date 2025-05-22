import React from "react";
import { useFonts } from "expo-font";
import { createIconSet } from "@expo/vector-icons";
import glyphMap from "../../assets/tabler-icons/glyph-map.json";

// create the icon set using the glyph map and font family name
const TablerIcon = createIconSet(glyphMap, "TablerIcons", "tabler-icons.ttf");

export default function TablerIconComponent(props: any) {
  const [fontsLoaded] = useFonts({
    TablerIcons: require("../../assets/tabler-icons/tabler-icons.ttf"),
  });

  if (!fontsLoaded) {
    return null;
  }

  return <TablerIcon {...props} />;
}
