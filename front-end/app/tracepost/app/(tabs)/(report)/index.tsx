import React, { useState, useEffect } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  Image,
  TextInput,
  ActivityIndicator,
  Alert,
  RefreshControl,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import TablerIconComponent from "@/components/icon";
import * as ImagePicker from "expo-image-picker";
import { useRole } from "@/contexts/RoleContext";
import { makeAuthenticatedRequest } from "@/api/auth";
import { getAllBatches, BatchData } from "@/api/batch";
import "@/global.css";

// Define interfaces for the report data
interface CreateEventRequest {
  actor_id: number;
  batch_id: number;
  event_type: "feeding" | "disease" | "harvest";
  location: string;
  metadata: Record<string, string>;
}

interface Event {
  id: number;
  batch_id: number;
  batch_info: {
    quantity: number;
    species: string;
    status: string;
  };
  event_type: string;
  facility_info: {
    company_name: string;
    hatchery_name: string;
  };
  is_active: boolean;
  location: string;
  metadata: Record<string, any>;
  timestamp: string;
  updated_at: string;
}

interface EventResponse {
  success: boolean;
  message: string;
  data: Event[] | Event | null; // Add null as a possible type
}

export default function ReportScreen() {
  const [activeTab, setActiveTab] = useState("feeding");
  const [images, setImages] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSigning, setIsSigning] = useState(false);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isLoadingBatches, setIsLoadingBatches] = useState(true);
  const [isLoadingEvents, setIsLoadingEvents] = useState(true);
  const [userEvents, setUserEvents] = useState<Event[]>([]);
  const [availableBatches, setAvailableBatches] = useState<BatchData[]>([]);

  const [formData, setFormData] = useState({
    batch_id: "",
    feedType: "",
    amount: "",
    notes: "",
    diseaseType: "",
    severity: "",
    treatmentApplied: "",
    harvestAmount: "",
    harvestQuality: "",
  });

  const { userData } = useRole();

  useEffect(() => {
    loadBatches();
    loadUserEvents();
  }, []);

  // Load batches for selection
  const loadBatches = async () => {
    try {
      setIsLoadingBatches(true);
      const response = await getAllBatches();
      if (response.success && response.data) {
        // Handle potential null data
        const batchesArray = Array.isArray(response.data) ? response.data : [];
        setAvailableBatches(batchesArray);
      } else {
        setAvailableBatches([]);
      }
    } catch (error) {
      console.error("Error loading batches:", error);
      setAvailableBatches([]); // Set empty array on error
      Alert.alert("Error", "Could not load batches. Please try again later.");
    } finally {
      setIsLoadingBatches(false);
    }
  };

  // Load user's previous events
  const loadUserEvents = async () => {
    try {
      setIsLoadingEvents(true);
      const response = await makeAuthenticatedRequest(
        `${process.env.EXPO_PUBLIC_API_URL}/events`,
        { method: "GET" },
      );

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || "Failed to fetch events");
      }

      const data: EventResponse = await response.json();

      if (data.success) {
        // Handle null data by providing empty array fallback
        const eventsArray = Array.isArray(data.data) ? data.data : [];

        // Filter events to only show those created by the current user
        const userEventsFiltered = eventsArray.filter(
          (event) => event.event_type !== "batch_created",
        );
        setUserEvents(userEventsFiltered);
      } else {
        // If success is false, set empty array
        setUserEvents([]);
      }
    } catch (error) {
      console.error("Error loading events:", error);
      setUserEvents([]); // Set empty array on error
      Alert.alert(
        "Error",
        "Could not load your reports. Please try again later.",
      );
    } finally {
      setIsLoadingEvents(false);
    }
  };

  const refreshData = async () => {
    setIsRefreshing(true);
    await Promise.all([loadBatches(), loadUserEvents()]);
    setIsRefreshing(false);
  };

  const pickImage = async () => {
    // No permissions request is necessary for launching the image library
    let result = await ImagePicker.launchImageLibraryAsync({
      mediaTypes: ImagePicker.MediaTypeOptions.Images,
      allowsEditing: true,
      aspect: [4, 3],
      quality: 0.8,
    });

    if (!result.canceled) {
      setImages([...images, result.assets[0].uri]);
    }
  };

  const signWithWallet = async () => {
    if (!isFormValid()) {
      Alert.alert("Form Error", "Please fill out all required fields");
      return;
    }

    setIsSigning(true);

    try {
      // In a real app, here you would sign the transaction with a blockchain wallet
      setTimeout(() => {
        setIsSigning(false);
        handleSubmit();
      }, 1500);
    } catch (error) {
      setIsSigning(false);
      Alert.alert("Error", "Signing failed. Please try again.");
    }
  };

  const handleSubmit = async () => {
    if (!userData?.user_id) {
      Alert.alert("Error", "User information not available");
      return;
    }

    setIsSubmitting(true);

    try {
      // Build request based on selected tab
      const request: CreateEventRequest = {
        actor_id: userData.user_id,
        batch_id: parseInt(formData.batch_id),
        event_type: activeTab as "feeding" | "disease" | "harvest",
        location: "Batch area",
        metadata: {},
      };

      // Add metadata based on report type
      if (activeTab === "feeding") {
        request.metadata = {
          food_type: formData.feedType || "",
          amount: formData.amount || "",
          notes: formData.notes || "",
        };
      } else if (activeTab === "disease") {
        request.metadata = {
          disease_type: formData.diseaseType || "",
          severity: formData.severity || "",
          treatment: formData.treatmentApplied || "",
          notes: formData.notes || "",
        };
      } else if (activeTab === "harvest") {
        request.metadata = {
          amount: formData.harvestAmount || "",
          quality: formData.harvestQuality || "",
          notes: formData.notes || "",
        };
      }

      // Call the API
      const response = await makeAuthenticatedRequest(
        `${process.env.EXPO_PUBLIC_API_URL}/events`,
        {
          method: "POST",
          body: JSON.stringify(request),
        },
      );

      const responseData = await response.json();

      if (!response.ok) {
        throw new Error(responseData.message || "Failed to create event");
      }

      // Reset form
      setFormData({
        batch_id: "",
        feedType: "",
        amount: "",
        notes: "",
        diseaseType: "",
        severity: "",
        treatmentApplied: "",
        harvestAmount: "",
        harvestQuality: "",
      });
      setImages([]);

      // Reload events to show the newly created one
      await loadUserEvents();

      // Show success message with transaction details
      const eventId = responseData.data?.id || "Unknown";
      Alert.alert(
        "Report Submitted",
        `Your ${activeTab} report has been recorded successfully. Event ID: ${eventId}`,
        [{ text: "OK" }],
      );
    } catch (error) {
      console.error("Error submitting report:", error);
      Alert.alert(
        "Error",
        error instanceof Error
          ? error.message
          : "Failed to submit report. Please try again.",
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  const isFormValid = () => {
    if (!formData.batch_id) {
      return false;
    }

    if (activeTab === "feeding") {
      return formData.feedType && formData.amount;
    } else if (activeTab === "disease") {
      return formData.diseaseType && formData.severity;
    } else if (activeTab === "harvest") {
      return formData.harvestAmount;
    }
    return false;
  };

  const getEventIcon = (eventType: string) => {
    switch (eventType) {
      case "feeding":
        return "bucket";
      case "disease":
        return "virus";
      case "harvest":
        return "scale";
      default:
        return "circle";
    }
  };

  const getEventColor = (eventType: string) => {
    switch (eventType) {
      case "feeding":
        return "#f97316";
      case "disease":
        return "#ef4444";
      case "harvest":
        return "#10b981";
      default:
        return "#6b7280";
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const renderBatchSelector = () => (
    <View className="mb-4">
      <Text className="font-medium text-gray-700 mb-1">Select Batch *</Text>
      {isLoadingBatches ? (
        <View className="p-4 border border-gray-300 rounded-xl bg-gray-50 items-center">
          <ActivityIndicator size="small" color="#3b82f6" />
          <Text className="text-gray-500 mt-2">Loading batches...</Text>
        </View>
      ) : availableBatches.length === 0 ? (
        <View className="p-4 border border-gray-300 rounded-xl bg-gray-50 items-center">
          <Text className="text-gray-500">No batches available</Text>
        </View>
      ) : (
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          className="pb-2"
        >
          {availableBatches.map((batch) => (
            <TouchableOpacity
              key={batch.id}
              className={`p-3 border rounded-xl mr-3 mb-1 min-w-40 ${
                formData.batch_id === batch.id.toString()
                  ? "border-primary bg-primary/10"
                  : "border-gray-300 bg-white"
              }`}
              onPress={() =>
                setFormData({ ...formData, batch_id: batch.id.toString() })
              }
            >
              <View className="flex-row justify-between items-start">
                <Text className="font-semibold">Batch #{batch.id}</Text>
                <View
                  className={`w-4 h-4 rounded-full border ${
                    formData.batch_id === batch.id.toString()
                      ? "border-primary bg-primary"
                      : "border-gray-300"
                  }`}
                />
              </View>
              <Text className="text-gray-600 text-sm" numberOfLines={1}>
                {batch.species}
              </Text>
              <Text className="text-gray-600 text-xs mt-1">
                Qty: {batch.quantity.toLocaleString()}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      )}
    </View>
  );

  return (
    <SafeAreaView className="flex-1 bg-white">
      <ScrollView
        contentContainerStyle={{ paddingBottom: 100 }}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={refreshData}
            colors={["#3b82f6"]}
            tintColor="#3b82f6"
          />
        }
      >
        <View className="px-5 pt-4 pb-6">
          {/* Header */}
          <View className="flex-row items-center justify-between mb-6">
            <View>
              <Text className="text-2xl font-bold text-gray-800">
                Report Events
              </Text>
              <Text className="text-gray-500">
                Record farm activities & issues
              </Text>

              {userData && (
                <Text className="text-xs text-gray-400 mt-1">
                  {userData.username} • Farm User
                </Text>
              )}
            </View>
            <TouchableOpacity
              className="h-10 w-10 rounded-full bg-secondary/10 items-center justify-center"
              onPress={refreshData}
            >
              <TablerIconComponent name="refresh" size={20} color="#4338ca" />
            </TouchableOpacity>
          </View>

          {/* Blockchain Badge */}
          <View className="bg-green-50 px-4 py-3 rounded-xl mb-6 flex-row items-center">
            <TablerIconComponent
              name="currency-ethereum"
              size={20}
              color="#10b981"
            />
            <Text className="ml-2 text-green-700 flex-1">
              Reports are securely recorded on the blockchain for traceability
              and immutability
            </Text>
          </View>

          {/* Tabs */}
          <View className="mb-6">
            <Text className="text-sm font-medium text-gray-700 mb-2">
              Report Type
            </Text>
            <View className="flex-row">
              {["feeding", "disease", "harvest"].map((tab) => (
                <TouchableOpacity
                  key={tab}
                  className={`px-4 py-3 rounded-full mr-3 ${
                    activeTab === tab ? "bg-primary" : "bg-gray-100"
                  }`}
                  onPress={() => setActiveTab(tab)}
                >
                  <Text
                    className={`font-medium capitalize ${
                      activeTab === tab ? "text-white" : "text-gray-600"
                    }`}
                  >
                    {tab}
                  </Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {/* Form Section */}
          <View className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm mb-6">
            {/* Batch Selection - Common for all report types */}
            {renderBatchSelector()}

            {activeTab === "feeding" && (
              <>
                <Text className="text-lg font-semibold mb-4">
                  Feeding Report
                </Text>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Feed Type *
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="bucket"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter feed type"
                      value={formData.feedType}
                      onChangeText={(text) =>
                        setFormData({ ...formData, feedType: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Amount (kg) *
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="scale"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter amount"
                      keyboardType="numeric"
                      value={formData.amount}
                      onChangeText={(text) =>
                        setFormData({ ...formData, amount: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Notes (Optional)
                  </Text>
                  <TextInput
                    className="p-3 border border-gray-300 rounded-lg bg-white"
                    placeholder="Additional information"
                    multiline
                    numberOfLines={3}
                    textAlignVertical="top"
                    value={formData.notes}
                    onChangeText={(text) =>
                      setFormData({ ...formData, notes: text })
                    }
                  />
                </View>
              </>
            )}

            {activeTab === "disease" && (
              <>
                <Text className="text-lg font-semibold mb-4">
                  Disease Report
                </Text>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Disease Type *
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="virus"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter disease type"
                      value={formData.diseaseType}
                      onChangeText={(text) =>
                        setFormData({ ...formData, diseaseType: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Severity *
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="alert-triangle"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Low / Medium / High"
                      value={formData.severity}
                      onChangeText={(text) =>
                        setFormData({ ...formData, severity: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Treatment Applied
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="medicine"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter treatment if any"
                      value={formData.treatmentApplied}
                      onChangeText={(text) =>
                        setFormData({ ...formData, treatmentApplied: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-6">
                  <Text className="font-medium text-gray-700 mb-3">Images</Text>

                  <View className="flex-row flex-wrap">
                    {images.map((uri, index) => (
                      <View
                        key={index}
                        className="w-20 h-20 m-1 rounded-lg overflow-hidden"
                      >
                        <Image source={{ uri }} className="w-full h-full" />
                        <TouchableOpacity
                          className="absolute top-1 right-1 bg-red-500 rounded-full p-1"
                          onPress={() => {
                            const newImages = [...images];
                            newImages.splice(index, 1);
                            setImages(newImages);
                          }}
                        >
                          <TablerIconComponent
                            name="x"
                            size={12}
                            color="white"
                          />
                        </TouchableOpacity>
                      </View>
                    ))}

                    <TouchableOpacity
                      className="w-20 h-20 border-2 border-dashed border-gray-300 rounded-lg m-1 items-center justify-center"
                      onPress={pickImage}
                    >
                      <TablerIconComponent
                        name="plus"
                        size={24}
                        color="#9ca3af"
                      />
                    </TouchableOpacity>
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">Notes</Text>
                  <TextInput
                    className="p-3 border border-gray-300 rounded-lg bg-white"
                    placeholder="Additional information"
                    multiline
                    numberOfLines={3}
                    textAlignVertical="top"
                    value={formData.notes}
                    onChangeText={(text) =>
                      setFormData({ ...formData, notes: text })
                    }
                  />
                </View>
              </>
            )}

            {activeTab === "harvest" && (
              <>
                <Text className="text-lg font-semibold mb-4">
                  Harvest Report
                </Text>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Harvest Amount (kg) *
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="scale"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter amount"
                      keyboardType="numeric"
                      value={formData.harvestAmount}
                      onChangeText={(text) =>
                        setFormData({ ...formData, harvestAmount: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Quality Grade
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="star"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="A / B / C Grade"
                      value={formData.harvestQuality}
                      onChangeText={(text) =>
                        setFormData({ ...formData, harvestQuality: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-6">
                  <Text className="font-medium text-gray-700 mb-3">Images</Text>

                  <View className="flex-row flex-wrap">
                    {images.map((uri, index) => (
                      <View
                        key={index}
                        className="w-20 h-20 m-1 rounded-lg overflow-hidden"
                      >
                        <Image source={{ uri }} className="w-full h-full" />
                        <TouchableOpacity
                          className="absolute top-1 right-1 bg-red-500 rounded-full p-1"
                          onPress={() => {
                            const newImages = [...images];
                            newImages.splice(index, 1);
                            setImages(newImages);
                          }}
                        >
                          <TablerIconComponent
                            name="x"
                            size={12}
                            color="white"
                          />
                        </TouchableOpacity>
                      </View>
                    ))}

                    <TouchableOpacity
                      className="w-20 h-20 border-2 border-dashed border-gray-300 rounded-lg m-1 items-center justify-center"
                      onPress={pickImage}
                    >
                      <TablerIconComponent
                        name="plus"
                        size={24}
                        color="#9ca3af"
                      />
                    </TouchableOpacity>
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">Notes</Text>
                  <TextInput
                    className="p-3 border border-gray-300 rounded-lg bg-white"
                    placeholder="Additional information"
                    multiline
                    numberOfLines={3}
                    textAlignVertical="top"
                    value={formData.notes}
                    onChangeText={(text) =>
                      setFormData({ ...formData, notes: text })
                    }
                  />
                </View>
              </>
            )}

            {/* Web3 signature notice */}
            <View className="bg-indigo-50 p-3 rounded-lg mb-4">
              <Text className="text-sm text-indigo-700">
                <TablerIconComponent
                  name="info-circle"
                  size={16}
                  color="#4338ca"
                  style={{ marginRight: 4 }}
                />
                This report will be signed with your wallet and recorded on the
                blockchain for transparency and traceability
              </Text>
            </View>

            <TouchableOpacity
              className={`rounded-xl py-4 ${!isFormValid() ? "bg-gray-300" : isSigning ? "bg-indigo-300" : isSubmitting ? "bg-primary/60" : "bg-primary"} items-center mt-4`}
              onPress={signWithWallet}
              disabled={!isFormValid() || isSigning || isSubmitting}
            >
              {isSigning ? (
                <View className="flex-row items-center">
                  <ActivityIndicator color="white" style={{ marginRight: 8 }} />
                  <Text className="font-bold text-white text-base">
                    SIGNING WITH WALLET
                  </Text>
                </View>
              ) : isSubmitting ? (
                <View className="flex-row items-center">
                  <ActivityIndicator color="white" style={{ marginRight: 8 }} />
                  <Text className="font-bold text-white text-base">
                    SUBMITTING TO BLOCKCHAIN
                  </Text>
                </View>
              ) : (
                <View className="flex-row items-center">
                  <TablerIconComponent
                    name="wallet"
                    size={20}
                    color="white"
                    style={{ marginRight: 8 }}
                  />
                  <Text className="font-bold text-white text-base">
                    SIGN & SUBMIT REPORT
                  </Text>
                </View>
              )}
            </TouchableOpacity>
          </View>

          {/* Recent Reports */}
          <Text className="text-lg font-semibold mb-4">
            Your Recent Reports
          </Text>

          {isLoadingEvents ? (
            <View className="p-8 bg-white border border-gray-200 rounded-xl items-center justify-center">
              <ActivityIndicator color="#3b82f6" />
              <Text className="text-gray-500 mt-3">
                Loading your reports...
              </Text>
            </View>
          ) : userEvents.length === 0 ? (
            <View className="p-8 bg-white border border-gray-200 rounded-xl items-center justify-center">
              <TablerIconComponent
                name="clipboard-off"
                size={48}
                color="#9ca3af"
              />
              <Text className="text-gray-500 font-medium mt-4 mb-2">
                No reports yet
              </Text>
              <Text className="text-gray-400 text-center">
                Your submitted reports will appear here
              </Text>
            </View>
          ) : (
            userEvents.map((event) => (
              <View
                key={event.id}
                className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
              >
                <View className="flex-row items-center mb-2">
                  <View
                    className="h-10 w-10 rounded-full items-center justify-center mr-3"
                    style={{
                      backgroundColor: `${getEventColor(event.event_type)}20`,
                    }}
                  >
                    <TablerIconComponent
                      name={getEventIcon(event.event_type)}
                      size={20}
                      color={getEventColor(event.event_type)}
                    />
                  </View>
                  <View className="flex-1">
                    <Text className="font-medium capitalize">
                      {event.event_type} - Batch #{event.batch_id}
                    </Text>
                    <Text className="text-gray-500 text-xs">
                      {formatDate(event.timestamp)}
                    </Text>
                  </View>

                  {/* Show relevant data based on event type */}
                  <View className="bg-gray-100 px-2 py-1 rounded">
                    {event.event_type === "feeding" &&
                      event.metadata.amount && (
                        <Text className="text-xs text-gray-600">
                          {event.metadata.amount}kg
                        </Text>
                      )}
                    {event.event_type === "disease" &&
                      event.metadata.severity && (
                        <Text className="text-xs text-gray-600">
                          {event.metadata.severity}
                        </Text>
                      )}
                    {event.event_type === "harvest" &&
                      event.metadata.amount && (
                        <Text className="text-xs text-gray-600">
                          {event.metadata.amount}kg
                        </Text>
                      )}
                  </View>
                </View>

                {/* Display metadata */}
                {event.metadata && Object.keys(event.metadata).length > 0 && (
                  <View className="bg-gray-50 p-2 rounded-lg my-2">
                    {Object.entries(event.metadata).map(
                      ([key, value]) =>
                        key !== "amount" &&
                        key !== "severity" &&
                        value != null && ( // Check for null/undefined values
                          <Text key={key} className="text-xs text-gray-600">
                            <Text className="font-medium capitalize">
                              {key.replace(/_/g, " ")}:
                            </Text>{" "}
                            {String(value)}
                          </Text>
                        ),
                    )}
                  </View>
                )}

                {/* Batch info with null checks */}
                {event.batch_info && (
                  <View className="bg-blue-50 p-2 rounded-lg my-2">
                    <Text className="text-xs text-blue-700">
                      <Text className="font-medium">Species:</Text>{" "}
                      {event.batch_info.species || "Unknown"}
                    </Text>
                    <Text className="text-xs text-blue-700">
                      <Text className="font-medium">Status:</Text>{" "}
                      {event.batch_info.status || "Unknown"}
                    </Text>
                    <Text className="text-xs text-blue-700">
                      <Text className="font-medium">Quantity:</Text>{" "}
                      {event.batch_info.quantity?.toLocaleString() || "Unknown"}
                    </Text>
                  </View>
                )}

                {/* Blockchain verification tag */}
                <View className="flex-row items-center mt-2 pt-2 border-t border-gray-100">
                  <TablerIconComponent
                    name="currency-ethereum"
                    size={12}
                    color="#9ca3af"
                  />
                  <Text className="text-xs text-gray-500 ml-1">
                    Verified on blockchain
                  </Text>
                </View>
              </View>
            ))
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
