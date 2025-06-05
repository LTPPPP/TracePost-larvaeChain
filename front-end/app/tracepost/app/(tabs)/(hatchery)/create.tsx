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
import "@/global.css";

interface HatcheryFormData {
  name: string;
  location: string;
  address: string;
  capacity: string;
  managerName: string;
  managerPhone: string;
  managerEmail: string;
  establishedDate: string;
  licenseNumber: string;
  description: string;
  coordinates: string;
}

export default function CreateHatcheryScreen() {
  const [formData, setFormData] = useState<HatcheryFormData>({
    name: "",
    location: "",
    address: "",
    capacity: "",
    managerName: "",
    managerPhone: "",
    managerEmail: "",
    establishedDate: "",
    licenseNumber: "",
    description: "",
    coordinates: "",
  });

  const [errors, setErrors] = useState<Partial<HatcheryFormData>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [currentStep, setCurrentStep] = useState(1);

  const router = useRouter();
  const { userData } = useRole();

  const updateFormData = (field: keyof HatcheryFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: "" }));
    }
  };

  const validateStep = (step: number): boolean => {
    const newErrors: Partial<HatcheryFormData> = {};

    if (step === 1) {
      if (!formData.name.trim()) {
        newErrors.name = "Hatchery name is required";
      }
      if (!formData.location.trim()) {
        newErrors.location = "Location is required";
      }
      if (!formData.address.trim()) {
        newErrors.address = "Address is required";
      }
      if (!formData.capacity.trim()) {
        newErrors.capacity = "Capacity is required";
      } else if (
        isNaN(Number(formData.capacity)) ||
        Number(formData.capacity) <= 0
      ) {
        newErrors.capacity = "Capacity must be a valid number";
      }
    } else if (step === 2) {
      if (!formData.managerName.trim()) {
        newErrors.managerName = "Manager name is required";
      }
      if (!formData.managerPhone.trim()) {
        newErrors.managerPhone = "Manager phone is required";
      }
      if (!formData.managerEmail.trim()) {
        newErrors.managerEmail = "Manager email is required";
      } else if (!/\S+@\S+\.\S+/.test(formData.managerEmail)) {
        newErrors.managerEmail = "Please enter a valid email address";
      }
      if (!formData.establishedDate.trim()) {
        newErrors.establishedDate = "Established date is required";
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleNext = () => {
    if (validateStep(currentStep)) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleBack = () => {
    setCurrentStep(currentStep - 1);
  };

  const handleSubmit = async () => {
    if (!validateStep(currentStep)) return;

    setIsSubmitting(true);

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Show success message
      Alert.alert(
        "Hatchery Created Successfully",
        `${formData.name} has been created and registered on the blockchain.`,
        [
          {
            text: "View Hatchery",
            onPress: () => {
              router.replace("/(tabs)/(hatchery)");
            },
          },
          {
            text: "Create Another",
            onPress: () => {
              // Reset form
              setFormData({
                name: "",
                location: "",
                address: "",
                capacity: "",
                managerName: "",
                managerPhone: "",
                managerEmail: "",
                establishedDate: "",
                licenseNumber: "",
                description: "",
                coordinates: "",
              });
              setCurrentStep(1);
            },
          },
        ],
      );
    } catch (error) {
      console.error("Error creating hatchery:", error);
      Alert.alert("Error", "Failed to create hatchery. Please try again.");
    } finally {
      setIsSubmitting(false);
    }
  };

  const renderStep1 = () => (
    <View>
      <Text className="text-lg font-semibold mb-4">Basic Information</Text>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">Hatchery Name *</Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.name ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="Enter hatchery name"
          value={formData.name}
          onChangeText={(text) => updateFormData("name", text)}
        />
        {errors.name && (
          <Text className="text-red-500 text-xs mt-1">{errors.name}</Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">Location *</Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.location ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="City, Province/State"
          value={formData.location}
          onChangeText={(text) => updateFormData("location", text)}
        />
        {errors.location && (
          <Text className="text-red-500 text-xs mt-1">{errors.location}</Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">Full Address *</Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.address ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="Complete street address"
          value={formData.address}
          onChangeText={(text) => updateFormData("address", text)}
          multiline
          numberOfLines={2}
        />
        {errors.address && (
          <Text className="text-red-500 text-xs mt-1">{errors.address}</Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          Capacity (larvae) *
        </Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.capacity ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="Maximum larvae capacity"
          value={formData.capacity}
          onChangeText={(text) => updateFormData("capacity", text)}
          keyboardType="numeric"
        />
        {errors.capacity && (
          <Text className="text-red-500 text-xs mt-1">{errors.capacity}</Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          GPS Coordinates (Optional)
        </Text>
        <TextInput
          className="p-3 border border-gray-300 rounded-xl bg-white"
          placeholder="Latitude, Longitude"
          value={formData.coordinates}
          onChangeText={(text) => updateFormData("coordinates", text)}
        />
      </View>

      <View className="mb-6">
        <Text className="font-medium text-gray-700 mb-1">
          Description (Optional)
        </Text>
        <TextInput
          className="p-3 border border-gray-300 rounded-xl bg-white"
          placeholder="Brief description of the hatchery"
          value={formData.description}
          onChangeText={(text) => updateFormData("description", text)}
          multiline
          numberOfLines={3}
          textAlignVertical="top"
        />
      </View>
    </View>
  );

  const renderStep2 = () => (
    <View>
      <Text className="text-lg font-semibold mb-4">Management & Legal</Text>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">Manager Name *</Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.managerName ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="Full name of hatchery manager"
          value={formData.managerName}
          onChangeText={(text) => updateFormData("managerName", text)}
        />
        {errors.managerName && (
          <Text className="text-red-500 text-xs mt-1">
            {errors.managerName}
          </Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">Manager Phone *</Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.managerPhone ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="+84 123 456 789"
          value={formData.managerPhone}
          onChangeText={(text) => updateFormData("managerPhone", text)}
          keyboardType="phone-pad"
        />
        {errors.managerPhone && (
          <Text className="text-red-500 text-xs mt-1">
            {errors.managerPhone}
          </Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">Manager Email *</Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.managerEmail ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="manager@example.com"
          value={formData.managerEmail}
          onChangeText={(text) => updateFormData("managerEmail", text)}
          keyboardType="email-address"
          autoCapitalize="none"
        />
        {errors.managerEmail && (
          <Text className="text-red-500 text-xs mt-1">
            {errors.managerEmail}
          </Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          Established Date *
        </Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.establishedDate ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="YYYY-MM-DD"
          value={formData.establishedDate}
          onChangeText={(text) => updateFormData("establishedDate", text)}
        />
        {errors.establishedDate && (
          <Text className="text-red-500 text-xs mt-1">
            {errors.establishedDate}
          </Text>
        )}
      </View>

      <View className="mb-6">
        <Text className="font-medium text-gray-700 mb-1">
          License Number (Optional)
        </Text>
        <TextInput
          className="p-3 border border-gray-300 rounded-xl bg-white"
          placeholder="Government license number"
          value={formData.licenseNumber}
          onChangeText={(text) => updateFormData("licenseNumber", text)}
        />
      </View>
    </View>
  );

  const renderStep3 = () => (
    <View>
      <Text className="text-lg font-semibold mb-4">Review & Submit</Text>

      <View className="bg-gray-50 p-4 rounded-xl mb-6">
        <Text className="font-semibold text-gray-800 mb-3">
          Hatchery Summary
        </Text>

        <View className="mb-3">
          <Text className="text-gray-600 text-sm">Name</Text>
          <Text className="font-medium">{formData.name}</Text>
        </View>

        <View className="mb-3">
          <Text className="text-gray-600 text-sm">Location</Text>
          <Text className="font-medium">{formData.location}</Text>
        </View>

        <View className="mb-3">
          <Text className="text-gray-600 text-sm">Address</Text>
          <Text className="font-medium">{formData.address}</Text>
        </View>

        <View className="mb-3">
          <Text className="text-gray-600 text-sm">Capacity</Text>
          <Text className="font-medium">
            {Number(formData.capacity).toLocaleString()} larvae
          </Text>
        </View>

        <View className="mb-3">
          <Text className="text-gray-600 text-sm">Manager</Text>
          <Text className="font-medium">{formData.managerName}</Text>
          <Text className="text-gray-500 text-sm">{formData.managerEmail}</Text>
          <Text className="text-gray-500 text-sm">{formData.managerPhone}</Text>
        </View>

        <View className="mb-3">
          <Text className="text-gray-600 text-sm">Established</Text>
          <Text className="font-medium">{formData.establishedDate}</Text>
        </View>

        {formData.licenseNumber && (
          <View className="mb-3">
            <Text className="text-gray-600 text-sm">License</Text>
            <Text className="font-medium">{formData.licenseNumber}</Text>
          </View>
        )}

        {formData.description && (
          <View>
            <Text className="text-gray-600 text-sm">Description</Text>
            <Text className="font-medium">{formData.description}</Text>
          </View>
        )}
      </View>

      <View className="bg-indigo-50 p-4 rounded-xl mb-6">
        <View className="flex-row items-center mb-2">
          <TablerIconComponent
            name="currency-ethereum"
            size={20}
            color="#4338ca"
          />
          <Text className="font-semibold text-indigo-800 ml-2">
            Blockchain Registration
          </Text>
        </View>
        <Text className="text-indigo-700 text-sm">
          This hatchery will be registered on the blockchain with a unique smart
          contract for complete transparency and traceability.
        </Text>
      </View>
    </View>
  );

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
                  Step {currentStep} of 3
                </Text>
              </View>
              <View className="w-10" />
            </View>

            {/* Progress Indicator */}
            <View className="flex-row mb-8">
              {[1, 2, 3].map((step) => (
                <View
                  key={step}
                  className={`flex-1 h-2 mx-1 rounded-full ${
                    step <= currentStep ? "bg-primary" : "bg-gray-200"
                  }`}
                />
              ))}
            </View>

            {/* Form Steps */}
            {currentStep === 1 && renderStep1()}
            {currentStep === 2 && renderStep2()}
            {currentStep === 3 && renderStep3()}

            {/* Navigation Buttons */}
            <View className="flex-row gap-3 mt-6">
              {currentStep > 1 && (
                <TouchableOpacity
                  className="flex-1 bg-gray-100 py-4 rounded-xl items-center"
                  onPress={handleBack}
                  disabled={isSubmitting}
                >
                  <Text className="font-bold text-gray-700">Back</Text>
                </TouchableOpacity>
              )}

              <TouchableOpacity
                className={`flex-1 py-4 rounded-xl items-center ${
                  isSubmitting ? "bg-primary/60" : "bg-primary"
                }`}
                onPress={currentStep === 3 ? handleSubmit : handleNext}
                disabled={isSubmitting}
              >
                {isSubmitting ? (
                  <View className="flex-row items-center">
                    <ActivityIndicator color="white" size="small" />
                    <Text className="font-bold text-white ml-2">
                      Creating...
                    </Text>
                  </View>
                ) : (
                  <Text className="font-bold text-white">
                    {currentStep === 3 ? "Create Hatchery" : "Next"}
                  </Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
