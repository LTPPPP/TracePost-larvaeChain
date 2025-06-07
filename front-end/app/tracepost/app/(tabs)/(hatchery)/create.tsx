import React, { useState } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  TextInput,
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Platform,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import TablerIconComponent from "@/components/icon";
import { useRouter } from "expo-router";
import { useRole } from "@/contexts/RoleContext";
import { createHatchery } from "@/api/hatchery";
import "@/global.css";

interface HatcheryFormData {
  name: string;
}

export default function CreateHatcheryScreen() {
  const [formData, setFormData] = useState<HatcheryFormData>({
    name: "",
  });

  const [errors, setErrors] = useState<Partial<HatcheryFormData>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const router = useRouter();
  const { userData, getCompanyId } = useRole();

  const updateFormData = (field: keyof HatcheryFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: "" }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Partial<HatcheryFormData> = {};

    if (!formData.name.trim()) {
      newErrors.name = "Hatchery name is required";
    } else if (formData.name.trim().length < 3) {
      newErrors.name = "Hatchery name must be at least 3 characters";
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async () => {
    if (!validateForm()) return;

    const companyId = getCompanyId();
    if (!companyId) {
      Alert.alert("Error", "Company ID not found. Please log in again.");
      return;
    }

    setIsSubmitting(true);

    try {
      const hatcheryData = {
        company_id: companyId,
        name: formData.name.trim(),
      };

      const response = await createHatchery(hatcheryData);

      if (response.success) {
        // Show success message
        Alert.alert(
          "Hatchery Created Successfully",
          `${response.data.name} has been created successfully.`,
          [
            {
              text: "View Hatcheries",
              onPress: () => {
                router.replace("/(tabs)/(hatchery)");
              },
            },
            {
              text: "Create Another",
              onPress: () => {
                // Reset form
                setFormData({ name: "" });
                setErrors({});
              },
            },
          ],
        );
      } else {
        Alert.alert("Error", response.message || "Failed to create hatchery");
      }
    } catch (error) {
      console.error("Error creating hatchery:", error);
      let errorMessage = "Failed to create hatchery. Please try again.";

      if (error instanceof Error) {
        errorMessage = error.message;
      }

      Alert.alert("Error", errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-white">
      <KeyboardAvoidingView
        className="flex-1"
        behavior={Platform.OS === "ios" ? "padding" : "height"}
      >
        <ScrollView
          contentContainerStyle={{ paddingBottom: 100 }}
          showsVerticalScrollIndicator={false}
        >
          <View className="px-5 pt-4 pb-6">
            {/* Header */}
            <View className="flex-row items-center justify-between mb-6">
              <TouchableOpacity
                className="h-10 w-10 rounded-full bg-gray-100 items-center justify-center"
                onPress={() => router.back()}
              >
                <TablerIconComponent name="arrow-left" size={20} color="#000" />
              </TouchableOpacity>
              <View className="flex-1 mx-4">
                <Text className="text-xl font-bold text-gray-800 text-center">
                  Create New Hatchery
                </Text>
                <Text className="text-gray-500 text-center text-sm">
                  Register a new breeding facility
                </Text>
              </View>
              <View className="w-10" />
            </View>

            {/* User Info */}
            {userData && (
              <View className="bg-blue-50 p-4 rounded-xl mb-6">
                <Text className="text-blue-800 font-medium">
                  Company Information
                </Text>
                <Text className="text-blue-700 text-sm">
                  Creating hatchery for Company ID: {getCompanyId()}
                </Text>
                <Text className="text-blue-600 text-xs">
                  Logged in as: {userData.username}
                </Text>
              </View>
            )}

            {/* Form */}
            <View className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
              <Text className="text-lg font-semibold mb-4">
                Hatchery Information
              </Text>

              <View className="mb-6">
                <Text className="font-medium text-gray-700 mb-1">
                  Hatchery Name *
                </Text>
                <TextInput
                  className={`p-3 border rounded-xl bg-white ${
                    errors.name ? "border-red-500" : "border-gray-300"
                  }`}
                  placeholder="Enter hatchery name"
                  value={formData.name}
                  onChangeText={(text) => updateFormData("name", text)}
                  editable={!isSubmitting}
                />
                {errors.name && (
                  <Text className="text-red-500 text-xs mt-1">
                    {errors.name}
                  </Text>
                )}
                <Text className="text-gray-500 text-xs mt-1">
                  Choose a descriptive name for your breeding facility
                </Text>
              </View>

              {/* API Information */}
              <View className="bg-indigo-50 p-4 rounded-xl mb-6">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="info-circle"
                    size={20}
                    color="#4338ca"
                  />
                  <Text className="font-semibold text-indigo-800 ml-2">
                    System Integration
                  </Text>
                </View>
                <Text className="text-indigo-700 text-sm">
                  This hatchery will be registered in the system and can be used
                  to create and manage breeding batches.
                </Text>
              </View>

              <TouchableOpacity
                className={`py-4 rounded-xl items-center ${
                  isSubmitting || !formData.name.trim()
                    ? "bg-gray-300"
                    : "bg-primary"
                }`}
                onPress={handleSubmit}
                disabled={isSubmitting || !formData.name.trim()}
              >
                {isSubmitting ? (
                  <View className="flex-row items-center">
                    <ActivityIndicator color="white" size="small" />
                    <Text className="font-bold text-white ml-2">
                      Creating Hatchery...
                    </Text>
                  </View>
                ) : (
                  <Text className="font-bold text-white text-lg">
                    Create Hatchery
                  </Text>
                )}
              </TouchableOpacity>
            </View>

            {/* Additional Information */}
            <View className="mt-6 bg-gray-50 p-4 rounded-xl">
              <Text className="font-medium text-gray-700 mb-2">
                What happens next?
              </Text>
              <View className="space-y-2">
                <View className="flex-row items-start">
                  <Text className="text-gray-600 mr-2">•</Text>
                  <Text className="text-gray-600 text-sm flex-1">
                    Your hatchery will be registered in the system
                  </Text>
                </View>
                <View className="flex-row items-start">
                  <Text className="text-gray-600 mr-2">•</Text>
                  <Text className="text-gray-600 text-sm flex-1">
                    You can start creating breeding batches
                  </Text>
                </View>
                <View className="flex-row items-start">
                  <Text className="text-gray-600 mr-2">•</Text>
                  <Text className="text-gray-600 text-sm flex-1">
                    All activities will be tracked for traceability
                  </Text>
                </View>
              </View>
            </View>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
