import React, { useState, useEffect } from "react";
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
import { createBatch } from "@/api/batch";
import { getHatcheries } from "@/api/hatchery";
import "@/global.css";

interface BatchFormData {
  hatcheryId: string;
  species: string;
  quantity: string;
}

interface Hatchery {
  id: number;
  name: string;
  company_id: number;
  company: {
    id: number;
    name: string;
    type: string;
    location: string;
    contact_info: string;
    created_at: string;
    updated_at: string;
    is_active: boolean;
  };
  created_at: string;
  updated_at: string;
  is_active: boolean;
}

export default function CreateBatchScreen() {
  const [formData, setFormData] = useState<BatchFormData>({
    hatcheryId: "",
    species: "",
    quantity: "",
  });

  const [errors, setErrors] = useState<Partial<BatchFormData>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [currentStep, setCurrentStep] = useState(1);
  const [selectedHatchery, setSelectedHatchery] = useState<Hatchery | null>(
    null,
  );
  const [availableHatcheries, setAvailableHatcheries] = useState<Hatchery[]>(
    [],
  );
  const [isLoadingHatcheries, setIsLoadingHatcheries] = useState(true);

  const router = useRouter();
  const { userData } = useRole();

  // Predefined species options
  const speciesOptions = [
    "Penaeus vannamei",
    "Penaeus monodon",
    "Penaeus japonicus",
    "Penaeus merguiensis",
    "Litopenaeus vannamei",
    "Macrobrachium rosenbergii",
  ];

  // Load available hatcheries
  const loadHatcheries = async () => {
    try {
      setIsLoadingHatcheries(true);
      const response = await getHatcheries();

      if (response.success) {
        // Filter only active hatcheries
        const activeHatcheries = response.data.filter((h) => h.is_active);
        setAvailableHatcheries(activeHatcheries);
      } else {
        throw new Error(response.message);
      }
    } catch (error) {
      console.error("Error loading hatcheries:", error);
      Alert.alert(
        "Error",
        error instanceof Error ? error.message : "Failed to load hatcheries",
      );
    } finally {
      setIsLoadingHatcheries(false);
    }
  };

  useEffect(() => {
    loadHatcheries();
  }, []);

  const updateFormData = (field: keyof BatchFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: "" }));
    }
  };

  const selectHatchery = (hatchery: Hatchery) => {
    setSelectedHatchery(hatchery);
    updateFormData("hatcheryId", hatchery.id.toString());
  };

  const validateStep = (step: number): boolean => {
    const newErrors: Partial<BatchFormData> = {};

    if (step === 1) {
      if (!formData.hatcheryId) {
        newErrors.hatcheryId = "Please select a hatchery";
      }
      if (!formData.species.trim()) {
        newErrors.species = "Species selection is required";
      }
      if (!formData.quantity.trim()) {
        newErrors.quantity = "Quantity is required";
      } else if (
        isNaN(Number(formData.quantity)) ||
        Number(formData.quantity) <= 0
      ) {
        newErrors.quantity = "Quantity must be a valid positive number";
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

  const generateBatchId = (): string => {
    const date = new Date();
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, "0");
    const day = String(date.getDate()).padStart(2, "0");
    const hatcheryCode = selectedHatchery?.name.charAt(0).toUpperCase() || "H";
    const randomNum = Math.floor(Math.random() * 1000)
      .toString()
      .padStart(3, "0");
    return `SH-${year}-${month}-${day}-${hatcheryCode}${randomNum}`;
  };

  const handleSubmit = async () => {
    if (!validateStep(currentStep)) return;

    setIsSubmitting(true);

    try {
      const batchData = {
        hatchery_id: Number(formData.hatcheryId),
        species: formData.species,
        quantity: Number(formData.quantity),
      };

      const response = await createBatch(batchData);

      if (response.success) {
        const { batch, blockchain } = response.data;

        // Show success message with blockchain info
        Alert.alert(
          "Batch Created Successfully",
          `Batch ${batch.id} has been created and recorded on the blockchain.\n\nTransaction IDs:\n${blockchain.transaction_ids.join("\n")}`,
          [
            {
              text: "View Batches",
              onPress: () => {
                router.replace("/(tabs)/(batches)");
              },
            },
            {
              text: "Create Another",
              onPress: () => {
                // Reset form
                setFormData({
                  hatcheryId: "",
                  species: "",
                  quantity: "",
                });
                setSelectedHatchery(null);
                setCurrentStep(1);
                setErrors({});
              },
            },
          ],
        );
      } else {
        Alert.alert("Error", response.message || "Failed to create batch");
      }
    } catch (error) {
      console.error("Error creating batch:", error);
      let errorMessage = "Failed to create batch. Please try again.";

      if (error instanceof Error) {
        errorMessage = error.message;
      }

      Alert.alert("Error", errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  const renderStep1 = () => (
    <View>
      <Text className="text-lg font-semibold mb-4">Basic Information</Text>

      {/* Hatchery Selection */}
      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-2">
          Select Hatchery *
        </Text>
        {isLoadingHatcheries ? (
          <View className="p-4 border border-gray-300 rounded-xl items-center">
            <ActivityIndicator size="small" color="#f97316" />
            <Text className="text-gray-500 mt-2">Loading hatcheries...</Text>
          </View>
        ) : availableHatcheries.length === 0 ? (
          <View className="p-4 border border-gray-300 rounded-xl items-center">
            <Text className="text-gray-500">No active hatcheries found</Text>
            <TouchableOpacity
              className="mt-2 bg-primary px-4 py-2 rounded-lg"
              onPress={() => router.push("/(tabs)/(hatchery)/create")}
            >
              <Text className="text-white text-sm">Create Hatchery</Text>
            </TouchableOpacity>
          </View>
        ) : (
          availableHatcheries.map((hatchery) => (
            <TouchableOpacity
              key={hatchery.id}
              className={`p-4 border rounded-xl mb-3 ${
                selectedHatchery?.id === hatchery.id
                  ? "border-primary bg-primary/5"
                  : "border-gray-300 bg-white"
              }`}
              onPress={() => selectHatchery(hatchery)}
            >
              <View className="flex-row justify-between items-start mb-2">
                <Text className="font-semibold text-gray-800">
                  {hatchery.name}
                </Text>
                <View
                  className={`w-4 h-4 rounded-full border-2 ${
                    selectedHatchery?.id === hatchery.id
                      ? "border-primary bg-primary"
                      : "border-gray-300"
                  }`}
                >
                  {selectedHatchery?.id === hatchery.id && (
                    <View className="w-full h-full rounded-full bg-white scale-50" />
                  )}
                </View>
              </View>
              <Text className="text-gray-500 text-sm mb-2">
                {hatchery.company.location}
              </Text>
              <View className="flex-row justify-between items-center">
                <Text className="text-sm text-gray-600">
                  Company: {hatchery.company.name}
                </Text>
                <Text className="text-sm text-gray-600">
                  Type: {hatchery.company.type}
                </Text>
              </View>
            </TouchableOpacity>
          ))
        )}
        {errors.hatcheryId && (
          <Text className="text-red-500 text-xs mt-1">{errors.hatcheryId}</Text>
        )}
      </View>

      {/* Species Selection */}
      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-2">Species *</Text>
        {speciesOptions.map((species) => (
          <TouchableOpacity
            key={species}
            className={`p-3 border rounded-xl mb-2 flex-row items-center ${
              formData.species === species
                ? "border-primary bg-primary/5"
                : "border-gray-300 bg-white"
            }`}
            onPress={() => updateFormData("species", species)}
          >
            <View
              className={`w-4 h-4 rounded-full border-2 mr-3 ${
                formData.species === species
                  ? "border-primary bg-primary"
                  : "border-gray-300"
              }`}
            >
              {formData.species === species && (
                <View className="w-full h-full rounded-full bg-white scale-50" />
              )}
            </View>
            <Text className="flex-1">{species}</Text>
          </TouchableOpacity>
        ))}
        {errors.species && (
          <Text className="text-red-500 text-xs mt-1">{errors.species}</Text>
        )}
      </View>

      {/* Quantity */}
      <View className="mb-6">
        <Text className="font-medium text-gray-700 mb-1">
          Initial Quantity (larvae) *
        </Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.quantity ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="Enter number of larvae"
          value={formData.quantity}
          onChangeText={(text) => updateFormData("quantity", text)}
          keyboardType="numeric"
        />
        {errors.quantity && (
          <Text className="text-red-500 text-xs mt-1">{errors.quantity}</Text>
        )}
        <Text className="text-gray-500 text-xs mt-1">
          Enter the initial number of larvae for this batch
        </Text>
      </View>
    </View>
  );

  const renderStep2 = () => {
    return (
      <View>
        <Text className="text-lg font-semibold mb-4">Review & Submit</Text>

        <View className="bg-gray-50 p-4 rounded-xl mb-6">
          <Text className="font-semibold text-gray-800 mb-3">
            Batch Summary
          </Text>

          <View className="mb-3">
            <Text className="text-gray-600 text-sm">Hatchery</Text>
            <Text className="font-medium">{selectedHatchery?.name}</Text>
            <Text className="text-gray-500 text-sm">
              {selectedHatchery?.company.location}
            </Text>
          </View>

          <View className="mb-3">
            <Text className="text-gray-600 text-sm">Species</Text>
            <Text className="font-medium">{formData.species}</Text>
          </View>

          <View className="mb-3">
            <Text className="text-gray-600 text-sm">Initial Quantity</Text>
            <Text className="font-medium">
              {Number(formData.quantity).toLocaleString()} larvae
            </Text>
          </View>

          <View className="mb-3">
            <Text className="text-gray-600 text-sm">Company</Text>
            <Text className="font-medium">
              {selectedHatchery?.company.name} ({selectedHatchery?.company.type}
              )
            </Text>
          </View>
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
            This batch will be registered on the blockchain with unique
            transaction IDs for complete traceability throughout its lifecycle.
          </Text>
        </View>

        <View className="bg-green-50 p-4 rounded-xl mb-6">
          <View className="flex-row items-center mb-2">
            <TablerIconComponent
              name="check-circle"
              size={20}
              color="#10b981"
            />
            <Text className="font-semibold text-green-800 ml-2">
              Automatic Processing
            </Text>
          </View>
          <Text className="text-green-700 text-sm">
            The batch will be automatically assigned an ID and status upon
            creation. All data will be securely stored and tracked.
          </Text>
        </View>
      </View>
    );
  };

  if (isLoadingHatcheries) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <View className="flex-1 justify-center items-center">
          <ActivityIndicator size="large" color="#f97316" />
          <Text className="text-gray-500 mt-4">Loading hatcheries...</Text>
        </View>
      </SafeAreaView>
    );
  }

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
                  Create New Batch
                </Text>
                <Text className="text-gray-500 text-center text-sm">
                  Step {currentStep} of 2
                </Text>
              </View>
              <View className="w-10" />
            </View>

            {/* Progress Indicator */}
            <View className="flex-row mb-8">
              {[1, 2].map((step) => (
                <View
                  key={step}
                  className={`flex-1 h-2 mx-1 rounded-full ${
                    step <= currentStep ? "bg-primary" : "bg-gray-200"
                  }`}
                />
              ))}
            </View>

            {/* User Info */}
            {userData && (
              <View className="bg-blue-50 p-4 rounded-xl mb-6">
                <Text className="text-blue-800 font-medium">
                  User Information
                </Text>
                <Text className="text-blue-700 text-sm">
                  Creating batch as: {userData.username}
                </Text>
                <Text className="text-blue-600 text-xs">
                  Available hatcheries: {availableHatcheries.length}
                </Text>
              </View>
            )}

            {/* Form Steps */}
            {currentStep === 1 && renderStep1()}
            {currentStep === 2 && renderStep2()}

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
                onPress={currentStep === 2 ? handleSubmit : handleNext}
                disabled={
                  isSubmitting ||
                  (currentStep === 1 && availableHatcheries.length === 0)
                }
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
                    {currentStep === 2 ? "Create Batch" : "Next"}
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
