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

interface BatchFormData {
  hatcheryId: string;
  species: string;
  quantity: string;
  startDate: string;
  estimatedDuration: string;
  temperature: string;
  ph: string;
  salinity: string;
  feedingSchedule: string;
  notes: string;
  tags: string;
}

interface Hatchery {
  id: number;
  name: string;
  location: string;
  capacity: number;
  currentStock: number;
}

export default function CreateBatchScreen() {
  const [formData, setFormData] = useState<BatchFormData>({
    hatcheryId: "",
    species: "",
    quantity: "",
    startDate: "",
    estimatedDuration: "",
    temperature: "",
    ph: "",
    salinity: "",
    feedingSchedule: "",
    notes: "",
    tags: "",
  });

  const [errors, setErrors] = useState<Partial<BatchFormData>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [currentStep, setCurrentStep] = useState(1);
  const [selectedHatchery, setSelectedHatchery] = useState<Hatchery | null>(
    null,
  );

  const router = useRouter();
  const { userData } = useRole();

  // Mock hatcheries data
  const availableHatcheries: Hatchery[] = [
    {
      id: 1,
      name: "Main Breeding Facility",
      location: "Mekong Delta, Vietnam",
      capacity: 15000,
      currentStock: 12500,
    },
    {
      id: 2,
      name: "Secondary Hatchery",
      location: "Can Tho, Vietnam",
      capacity: 10000,
      currentStock: 8200,
    },
    {
      id: 3,
      name: "Research & Development Center",
      location: "Ho Chi Minh City, Vietnam",
      capacity: 5000,
      currentStock: 2100,
    },
    {
      id: 4,
      name: "Coastal Breeding Station",
      location: "Phan Thiet, Vietnam",
      capacity: 8000,
      currentStock: 6800,
    },
  ];

  const speciesOptions = [
    "Penaeus vannamei (Pacific White Shrimp)",
    "Penaeus monodon (Giant Tiger Prawn)",
    "Penaeus japonicus (Kuruma Prawn)",
    "Penaeus merguiensis (Banana Prawn)",
  ];

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
        newErrors.quantity = "Quantity must be a valid number";
      } else if (selectedHatchery) {
        const availableCapacity =
          selectedHatchery.capacity - selectedHatchery.currentStock;
        if (Number(formData.quantity) > availableCapacity) {
          newErrors.quantity = `Exceeds available capacity (${availableCapacity.toLocaleString()})`;
        }
      }
      if (!formData.startDate.trim()) {
        newErrors.startDate = "Start date is required";
      }
    } else if (step === 2) {
      if (!formData.temperature.trim()) {
        newErrors.temperature = "Temperature is required";
      } else if (isNaN(Number(formData.temperature))) {
        newErrors.temperature = "Temperature must be a valid number";
      }
      if (!formData.ph.trim()) {
        newErrors.ph = "pH level is required";
      } else if (isNaN(Number(formData.ph))) {
        newErrors.ph = "pH must be a valid number";
      }
      if (!formData.salinity.trim()) {
        newErrors.salinity = "Salinity is required";
      } else if (isNaN(Number(formData.salinity))) {
        newErrors.salinity = "Salinity must be a valid number";
      }
      if (!formData.estimatedDuration.trim()) {
        newErrors.estimatedDuration = "Estimated duration is required";
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
    const hatcheryCode = selectedHatchery?.name.split(" ")[0].charAt(0) || "H";
    const randomNum = Math.floor(Math.random() * 1000)
      .toString()
      .padStart(3, "0");
    return `SH-${year}-${month}-${hatcheryCode}${randomNum}`;
  };

  const handleSubmit = async () => {
    if (!validateStep(currentStep)) return;

    setIsSubmitting(true);

    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 2000));

      const batchId = generateBatchId();

      // Show success message
      Alert.alert(
        "Batch Created Successfully",
        `Batch ${batchId} has been created and registered on the blockchain.`,
        [
          {
            text: "View Batch",
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
                startDate: "",
                estimatedDuration: "",
                temperature: "",
                ph: "",
                salinity: "",
                feedingSchedule: "",
                notes: "",
                tags: "",
              });
              setSelectedHatchery(null);
              setCurrentStep(1);
            },
          },
        ],
      );
    } catch (error) {
      console.error("Error creating batch:", error);
      Alert.alert("Error", "Failed to create batch. Please try again.");
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
        {availableHatcheries.map((hatchery) => {
          const availableCapacity = hatchery.capacity - hatchery.currentStock;
          const utilizationPercent = Math.round(
            (hatchery.currentStock / hatchery.capacity) * 100,
          );

          return (
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
                {hatchery.location}
              </Text>
              <View className="flex-row justify-between items-center">
                <Text className="text-sm text-gray-600">
                  Available: {availableCapacity.toLocaleString()} larvae
                </Text>
                <Text className="text-sm text-gray-600">
                  {utilizationPercent}% utilized
                </Text>
              </View>
              <View className="mt-2 h-2 bg-gray-200 rounded-full overflow-hidden">
                <View
                  className="h-full bg-primary rounded-full"
                  style={{ width: `${utilizationPercent}%` }}
                />
              </View>
            </TouchableOpacity>
          );
        })}
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
      <View className="mb-4">
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
        {selectedHatchery && (
          <Text className="text-gray-500 text-xs mt-1">
            Available capacity:{" "}
            {(
              selectedHatchery.capacity - selectedHatchery.currentStock
            ).toLocaleString()}{" "}
            larvae
          </Text>
        )}
        {errors.quantity && (
          <Text className="text-red-500 text-xs mt-1">{errors.quantity}</Text>
        )}
      </View>

      {/* Start Date */}
      <View className="mb-6">
        <Text className="font-medium text-gray-700 mb-1">Start Date *</Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.startDate ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="YYYY-MM-DD"
          value={formData.startDate}
          onChangeText={(text) => updateFormData("startDate", text)}
        />
        {errors.startDate && (
          <Text className="text-red-500 text-xs mt-1">{errors.startDate}</Text>
        )}
      </View>
    </View>
  );

  const renderStep2 = () => (
    <View>
      <Text className="text-lg font-semibold mb-4">
        Environmental Parameters
      </Text>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          Target Temperature (°C) *
        </Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.temperature ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="e.g., 28.5"
          value={formData.temperature}
          onChangeText={(text) => updateFormData("temperature", text)}
          keyboardType="decimal-pad"
        />
        {errors.temperature && (
          <Text className="text-red-500 text-xs mt-1">
            {errors.temperature}
          </Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          Target pH Level *
        </Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.ph ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="e.g., 7.2"
          value={formData.ph}
          onChangeText={(text) => updateFormData("ph", text)}
          keyboardType="decimal-pad"
        />
        {errors.ph && (
          <Text className="text-red-500 text-xs mt-1">{errors.ph}</Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          Target Salinity (ppt) *
        </Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.salinity ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="e.g., 15"
          value={formData.salinity}
          onChangeText={(text) => updateFormData("salinity", text)}
          keyboardType="decimal-pad"
        />
        {errors.salinity && (
          <Text className="text-red-500 text-xs mt-1">{errors.salinity}</Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          Estimated Duration (days) *
        </Text>
        <TextInput
          className={`p-3 border rounded-xl bg-white ${
            errors.estimatedDuration ? "border-red-500" : "border-gray-300"
          }`}
          placeholder="e.g., 45"
          value={formData.estimatedDuration}
          onChangeText={(text) => updateFormData("estimatedDuration", text)}
          keyboardType="numeric"
        />
        {errors.estimatedDuration && (
          <Text className="text-red-500 text-xs mt-1">
            {errors.estimatedDuration}
          </Text>
        )}
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">
          Feeding Schedule (Optional)
        </Text>
        <TextInput
          className="p-3 border border-gray-300 rounded-xl bg-white"
          placeholder="e.g., 3 times daily"
          value={formData.feedingSchedule}
          onChangeText={(text) => updateFormData("feedingSchedule", text)}
        />
      </View>

      <View className="mb-4">
        <Text className="font-medium text-gray-700 mb-1">Tags (Optional)</Text>
        <TextInput
          className="p-3 border border-gray-300 rounded-xl bg-white"
          placeholder="e.g., premium, research, export"
          value={formData.tags}
          onChangeText={(text) => updateFormData("tags", text)}
        />
        <Text className="text-gray-500 text-xs mt-1">
          Separate multiple tags with commas
        </Text>
      </View>

      <View className="mb-6">
        <Text className="font-medium text-gray-700 mb-1">
          Additional Notes (Optional)
        </Text>
        <TextInput
          className="p-3 border border-gray-300 rounded-xl bg-white"
          placeholder="Any additional information about this batch"
          value={formData.notes}
          onChangeText={(text) => updateFormData("notes", text)}
          multiline
          numberOfLines={3}
          textAlignVertical="top"
        />
      </View>
    </View>
  );

  const renderStep3 = () => {
    const estimatedCompletion =
      formData.startDate && formData.estimatedDuration
        ? new Date(
            new Date(formData.startDate).getTime() +
              Number(formData.estimatedDuration) * 24 * 60 * 60 * 1000,
          )
            .toISOString()
            .split("T")[0]
        : null;

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
              {selectedHatchery?.location}
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
            <Text className="text-gray-600 text-sm">Duration</Text>
            <Text className="font-medium">
              {formData.startDate} to {estimatedCompletion} (
              {formData.estimatedDuration} days)
            </Text>
          </View>

          <View className="mb-3">
            <Text className="text-gray-600 text-sm">
              Environmental Parameters
            </Text>
            <Text className="font-medium">
              Temperature: {formData.temperature}°C, pH: {formData.ph},
              Salinity: {formData.salinity} ppt
            </Text>
          </View>

          {formData.feedingSchedule && (
            <View className="mb-3">
              <Text className="text-gray-600 text-sm">Feeding Schedule</Text>
              <Text className="font-medium">{formData.feedingSchedule}</Text>
            </View>
          )}

          {formData.tags && (
            <View className="mb-3">
              <Text className="text-gray-600 text-sm">Tags</Text>
              <Text className="font-medium">{formData.tags}</Text>
            </View>
          )}

          {formData.notes && (
            <View>
              <Text className="text-gray-600 text-sm">Notes</Text>
              <Text className="font-medium">{formData.notes}</Text>
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
            This batch will be registered on the blockchain with a unique NFT
            for complete traceability throughout its lifecycle.
          </Text>
        </View>

        <View className="bg-green-50 p-4 rounded-xl mb-6">
          <View className="flex-row items-center mb-2">
            <TablerIconComponent name="qrcode" size={20} color="#10b981" />
            <Text className="font-semibold text-green-800 ml-2">
              QR Code Generation
            </Text>
          </View>
          <Text className="text-green-700 text-sm">
            A unique QR code will be generated for this batch, enabling easy
            tracking and verification by customers.
          </Text>
        </View>
      </View>
    );
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
                  Create New Batch
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
                    {currentStep === 3 ? "Create Batch" : "Next"}
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
