import React, { useState, useEffect } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  RefreshControl,
  Dimensions,
  Modal,
  TextInput,
  KeyboardAvoidingView,
  Platform,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useLocalSearchParams, useRouter } from "expo-router";
import { LineChart } from "react-native-chart-kit";
import TablerIconComponent from "@/components/icon";
import { useRole } from "@/contexts/RoleContext";
import {
  getHatcheryById,
  GetHatcheryResponse,
  updateHatchery,
  UpdateHatcheryRequest,
} from "@/api/hatchery";
import {
  getBatchesByHatchery,
  BatchData,
  GetBatchesResponse,
} from "@/api/batch";
import "@/global.css";

const screenWidth = Dimensions.get("window").width;

interface HatcheryStats {
  totalBatches: number;
  activeBatches: number;
  completedBatches: number;
  totalLarvae: number;
  averageQuantity: number;
  successRate: number;
}

interface StatusDistribution {
  [status: string]: number;
}

export default function HatcheryDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { isHatchery, userData } = useRole();

  // State management
  const [hatchery, setHatchery] = useState<GetHatcheryResponse["data"] | null>(
    null,
  );
  const [batches, setBatches] = useState<BatchData[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [activeTab, setActiveTab] = useState<
    "overview" | "batches" | "analytics"
  >("overview");
  const [selectedStatus, setSelectedStatus] = useState("all");

  // Name editing state
  const [isEditingName, setIsEditingName] = useState(false);
  const [editedName, setEditedName] = useState("");
  const [isUpdatingName, setIsUpdatingName] = useState(false);

  // Load hatchery and batch data
  const loadHatcheryData = async () => {
    if (!id) return;

    try {
      setIsLoading(true);
      const hatcheryId = parseInt(id);

      // Load hatchery details
      const hatcheryResponse = await getHatcheryById(hatcheryId);
      if (hatcheryResponse.success) {
        setHatchery(hatcheryResponse.data);
        setEditedName(hatcheryResponse.data.name); // Initialize edited name
      }

      // Load hatchery batches
      try {
        const batchesResponse = await getBatchesByHatchery(hatcheryId);
        if (batchesResponse.success) {
          // Handle null data case - when hatchery has no batches
          if (batchesResponse.data === null) {
            setBatches([]);
          } else {
            setBatches(batchesResponse.data);
          }
        }
      } catch (batchError) {
        console.log("No batches found for this hatchery:", batchError);
        setBatches([]);
      }
    } catch (error) {
      console.error("Error loading hatchery data:", error);
      Alert.alert(
        "Error",
        error instanceof Error ? error.message : "Failed to load hatchery data",
      );
    } finally {
      setIsLoading(false);
    }
  };

  // Refresh data
  const handleRefresh = async () => {
    setIsRefreshing(true);
    await loadHatcheryData();
    setIsRefreshing(false);
  };

  // Handle name editing
  const handleEditName = () => {
    if (!hatchery || !isHatchery) return;
    setEditedName(hatchery.name);
    setIsEditingName(true);
  };

  const handleSaveName = async () => {
    if (
      !hatchery ||
      !editedName.trim() ||
      editedName.trim() === hatchery.name
    ) {
      setIsEditingName(false);
      return;
    }

    setIsUpdatingName(true);
    try {
      const updateData: UpdateHatcheryRequest = {
        name: editedName.trim(),
      };

      const response = await updateHatchery(hatchery.id, updateData);

      if (response.success) {
        // Update local state with new data
        setHatchery(response.data);
        setIsEditingName(false);

        Alert.alert("Success", "Hatchery name updated successfully", [
          { text: "OK" },
        ]);
      }
    } catch (error) {
      console.error("Error updating hatchery name:", error);
      Alert.alert(
        "Error",
        error instanceof Error
          ? error.message
          : "Failed to update hatchery name",
      );
    } finally {
      setIsUpdatingName(false);
    }
  };

  const handleCancelEdit = () => {
    setEditedName(hatchery?.name || "");
    setIsEditingName(false);
  };

  // Navigate to batch detail
  const navigateToBatch = (batchId: number) => {
    router.push(`/batch/${batchId}`);
  };

  // Create new batch for this hatchery
  const createNewBatch = () => {
    router.push("/(tabs)/(batches)/create");
  };

  // Calculate hatchery statistics
  const calculateStats = (): HatcheryStats => {
    const totalBatches = batches.length;
    const activeBatches = batches.filter((b) => b.is_active).length;
    const completedBatches = batches.filter(
      (b) => b.status.toLowerCase() === "completed",
    ).length;
    const totalLarvae = batches.reduce((sum, b) => sum + b.quantity, 0);
    const averageQuantity =
      totalBatches > 0 ? Math.round(totalLarvae / totalBatches) : 0;
    const successRate =
      totalBatches > 0
        ? Math.round((completedBatches / totalBatches) * 100)
        : 0;

    return {
      totalBatches,
      activeBatches,
      completedBatches,
      totalLarvae,
      averageQuantity,
      successRate,
    };
  };

  // Get status distribution for analytics
  const getStatusDistribution = (): StatusDistribution => {
    const distribution: StatusDistribution = {};
    batches.forEach((batch) => {
      const status = batch.status.toLowerCase();
      distribution[status] = (distribution[status] || 0) + 1;
    });
    return distribution;
  };

  // Filter batches by status
  const getFilteredBatches = () => {
    if (selectedStatus === "all") {
      return batches;
    }
    return batches.filter(
      (batch) => batch.status.toLowerCase() === selectedStatus,
    );
  };

  // Generate chart data for batch quantities over time
  const getQuantityChartData = () => {
    const sortedBatches = [...batches]
      .sort(
        (a, b) =>
          new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
      )
      .slice(-6); // Last 6 batches

    if (sortedBatches.length === 0) {
      return {
        labels: ["No Data"],
        datasets: [{ data: [0] }],
      };
    }

    return {
      labels: sortedBatches.map((_, index) => `B${index + 1}`),
      datasets: [
        {
          data: sortedBatches.map((batch) => batch.quantity),
          color: (opacity = 1) => `rgba(249, 115, 22, ${opacity})`,
          strokeWidth: 2,
        },
      ],
      legend: ["Larvae Quantity"],
    };
  };

  // Helper functions
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case "created":
        return "bg-blue-100 text-blue-700";
      case "active":
        return "bg-green-100 text-green-700";
      case "completed":
        return "bg-gray-100 text-gray-700";
      case "failed":
        return "bg-red-100 text-red-700";
      default:
        return "bg-gray-100 text-gray-700";
    }
  };

  const chartConfig = {
    backgroundGradientFrom: "#fff",
    backgroundGradientTo: "#fff",
    decimalPlaces: 0,
    color: (opacity = 1) => `rgba(0, 0, 0, ${opacity})`,
    labelColor: (opacity = 1) => `rgba(0, 0, 0, ${opacity})`,
    style: {
      borderRadius: 16,
    },
    propsForDots: {
      r: "6",
      strokeWidth: "2",
    },
  };

  useEffect(() => {
    loadHatcheryData();
  }, [id]);

  if (isLoading) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <View className="flex-1 justify-center items-center">
          <ActivityIndicator size="large" color="#f97316" />
          <Text className="text-gray-500 mt-4">
            Loading hatchery details...
          </Text>
        </View>
      </SafeAreaView>
    );
  }

  if (!hatchery) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <View className="flex-1 justify-center items-center px-6">
          <TablerIconComponent
            name="building-factory-off"
            size={64}
            color="#9ca3af"
          />
          <Text className="text-gray-500 font-medium mt-4 mb-2 text-center">
            Hatchery not found
          </Text>
          <TouchableOpacity
            className="bg-primary px-6 py-3 rounded-xl"
            onPress={() => router.back()}
          >
            <Text className="text-white font-bold">Go Back</Text>
          </TouchableOpacity>
        </View>
      </SafeAreaView>
    );
  }

  const stats = calculateStats();
  const statusDistribution = getStatusDistribution();
  const filteredBatches = getFilteredBatches();

  return (
    <SafeAreaView className="flex-1 bg-white">
      <ScrollView
        contentContainerStyle={{ paddingBottom: 100 }}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={handleRefresh}
            colors={["#f97316"]}
            tintColor="#f97316"
          />
        }
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
              <View className="flex-row items-center justify-center">
                <Text className="text-xl font-bold text-gray-800 text-center">
                  {hatchery.name}
                </Text>
                {isHatchery && (
                  <TouchableOpacity
                    className="ml-2 p-1"
                    onPress={handleEditName}
                  >
                    <TablerIconComponent
                      name="edit"
                      size={16}
                      color="#f97316"
                    />
                  </TouchableOpacity>
                )}
              </View>
              <Text className="text-gray-500 text-center text-sm">
                Hatchery Details
              </Text>
            </View>
            <TouchableOpacity
              className="h-10 w-10 rounded-full bg-primary/10 items-center justify-center"
              onPress={handleRefresh}
            >
              <TablerIconComponent name="refresh" size={20} color="#f97316" />
            </TouchableOpacity>
          </View>

          {/* Hatchery Overview Card */}
          <View className="bg-sky-300 p-5 rounded-xl mb-6">
            <View className="flex-row justify-between items-start mb-4">
              <View>
                <Text className="text-white/80 text-sm">Hatchery Status</Text>
                <Text className="text-white font-bold text-xl">
                  {hatchery.is_active ? "Active" : "Inactive"}
                </Text>
              </View>
              <View
                className={`px-3 py-1 rounded-full ${
                  hatchery.is_active ? "bg-green-100" : "bg-red-100"
                }`}
              >
                <Text
                  className={`text-xs font-medium ${
                    hatchery.is_active ? "text-green-700" : "text-red-700"
                  }`}
                >
                  {hatchery.is_active ? "Active" : "Inactive"}
                </Text>
              </View>
            </View>

            <View className="flex-row flex-wrap">
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Company</Text>
                <Text className="text-white">{hatchery.company.name}</Text>
              </View>
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Location</Text>
                <Text className="text-white">{hatchery.company.location}</Text>
              </View>
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Type</Text>
                <Text className="text-white">{hatchery.company.type}</Text>
              </View>
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Created</Text>
                <Text className="text-white">
                  {formatDate(hatchery.created_at)}
                </Text>
              </View>
            </View>

            <View className="bg-white/20 p-3 rounded-lg mt-4">
              <View className="flex-row items-center justify-between">
                <View className="flex-row items-center">
                  <TablerIconComponent name="package" size={18} color="white" />
                  <Text className="text-white ml-2 font-medium">
                    {stats.totalBatches} Total Batches
                  </Text>
                </View>
                <Text className="text-white font-bold">
                  {stats.totalLarvae.toLocaleString()} Larvae
                </Text>
              </View>
            </View>
          </View>

          {/* Statistics Cards */}
          <View className="flex-row flex-wrap mb-6">
            <View className="w-1/2 pr-2 mb-4">
              <View className="bg-blue-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="package"
                    size={20}
                    color="#3b82f6"
                  />
                  <Text className="text-blue-700 font-medium ml-2">Active</Text>
                </View>
                <Text className="text-2xl font-bold text-blue-800">
                  {stats.activeBatches}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2 mb-4">
              <View className="bg-green-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent name="check" size={20} color="#10b981" />
                  <Text className="text-green-700 font-medium ml-2">
                    Completed
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-green-800">
                  {stats.completedBatches}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pr-2">
              <View className="bg-orange-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="chart-bar"
                    size={20}
                    color="#f97316"
                  />
                  <Text className="text-orange-700 font-medium ml-2">
                    Avg. Quantity
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-orange-800">
                  {stats.averageQuantity.toLocaleString()}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2">
              <View className="bg-purple-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="percentage"
                    size={20}
                    color="#8b5cf6"
                  />
                  <Text className="text-purple-700 font-medium ml-2">
                    Success Rate
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-purple-800">
                  {stats.successRate}%
                </Text>
              </View>
            </View>
          </View>

          {/* Tabs */}
          <View className="mb-4">
            <Text className="text-sm font-medium text-gray-700 mb-2">
              Details
            </Text>
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
              {["overview", "batches", "analytics"].map((tab) => (
                <TouchableOpacity
                  key={tab}
                  className={`px-4 py-2 rounded-full mr-3 ${
                    activeTab === tab ? "bg-primary" : "bg-gray-100"
                  }`}
                  onPress={() => setActiveTab(tab as any)}
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
            </ScrollView>
          </View>

          {/* Tab Content */}
          {activeTab === "overview" && (
            <View>
              {/* Company Information */}
              <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                <Text className="text-lg font-semibold mb-4">
                  Company Information
                </Text>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Company Name</Text>
                  <Text className="font-medium">{hatchery.company.name}</Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Business Type</Text>
                  <Text className="font-medium">{hatchery.company.type}</Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Location</Text>
                  <Text className="font-medium">
                    {hatchery.company.location}
                  </Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">
                    Contact Information
                  </Text>
                  <Text className="font-medium">
                    {hatchery.company.contact_info}
                  </Text>
                </View>

                <View className="flex-row items-center justify-between pt-3 border-t border-gray-100">
                  <Text className="text-gray-500 text-xs">
                    Company ID: {hatchery.company.id}
                  </Text>
                  <View
                    className={`px-2 py-1 rounded ${
                      hatchery.company.is_active ? "bg-green-100" : "bg-red-100"
                    }`}
                  >
                    <Text
                      className={`text-xs ${
                        hatchery.company.is_active
                          ? "text-green-600"
                          : "text-red-600"
                      }`}
                    >
                      {hatchery.company.is_active ? "Active" : "Inactive"}
                    </Text>
                  </View>
                </View>
              </View>

              {/* Recent Activity */}
              <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                <Text className="text-lg font-semibold mb-4">
                  Recent Batches
                </Text>

                {batches.length > 0 ? (
                  batches.slice(0, 3).map((batch) => (
                    <TouchableOpacity
                      key={batch.id}
                      className="flex-row items-center mb-3 last:mb-0 p-3 bg-gray-50 rounded-lg"
                      onPress={() => navigateToBatch(batch.id)}
                    >
                      <View className="h-10 w-10 rounded-full bg-primary/10 items-center justify-center mr-3">
                        <TablerIconComponent
                          name="package"
                          size={20}
                          color="#f97316"
                        />
                      </View>
                      <View className="flex-1">
                        <Text className="font-medium">Batch #{batch.id}</Text>
                        <Text className="text-gray-500 text-xs">
                          {batch.species} • {batch.quantity.toLocaleString()}{" "}
                          larvae
                        </Text>
                      </View>
                      <View className="items-end">
                        <View
                          className={`px-2 py-1 rounded ${getStatusColor(batch.status)}`}
                        >
                          <Text className="text-xs font-medium capitalize">
                            {batch.status}
                          </Text>
                        </View>
                        <Text className="text-gray-500 text-xs mt-1">
                          {formatDate(batch.created_at)}
                        </Text>
                      </View>
                    </TouchableOpacity>
                  ))
                ) : (
                  <View className="items-center py-8">
                    <TablerIconComponent
                      name="package-off"
                      size={48}
                      color="#9ca3af"
                    />
                    <Text className="text-gray-500 mt-2 text-center">
                      No batches yet
                    </Text>
                    <Text className="text-gray-400 text-center text-sm mt-1">
                      This hatchery hasn't created any batches
                    </Text>
                    {isHatchery && (
                      <TouchableOpacity
                        className="bg-primary px-4 py-2 rounded-lg mt-3"
                        onPress={createNewBatch}
                      >
                        <Text className="text-white font-medium">
                          Create First Batch
                        </Text>
                      </TouchableOpacity>
                    )}
                  </View>
                )}
              </View>
            </View>
          )}

          {activeTab === "batches" && (
            <View>
              {/* Batch Controls */}
              {isHatchery && (
                <View className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm">
                  <View className="flex-row items-center justify-between">
                    <Text className="text-lg font-semibold">
                      Batch Management
                    </Text>
                    <TouchableOpacity
                      className="bg-primary px-4 py-2 rounded-lg flex-row items-center"
                      onPress={createNewBatch}
                    >
                      <TablerIconComponent
                        name="plus"
                        size={16}
                        color="white"
                      />
                      <Text className="text-white ml-1 font-medium">
                        New Batch
                      </Text>
                    </TouchableOpacity>
                  </View>
                </View>
              )}

              {/* Status Filter - Only show if there are batches */}
              {batches.length > 0 &&
                Object.keys(statusDistribution).length > 0 && (
                  <View className="mb-4">
                    <Text className="text-sm font-medium text-gray-700 mb-2">
                      Filter by Status
                    </Text>
                    <ScrollView
                      horizontal
                      showsHorizontalScrollIndicator={false}
                    >
                      {["all", ...Object.keys(statusDistribution)].map(
                        (status) => (
                          <TouchableOpacity
                            key={status}
                            className={`px-4 py-2 rounded-full mr-3 ${
                              selectedStatus === status
                                ? "bg-secondary"
                                : "bg-gray-100"
                            }`}
                            onPress={() => setSelectedStatus(status)}
                          >
                            <Text
                              className={`font-medium capitalize ${
                                selectedStatus === status
                                  ? "text-white"
                                  : "text-gray-600"
                              }`}
                            >
                              {status}{" "}
                              {status !== "all" &&
                                `(${statusDistribution[status]})`}
                            </Text>
                          </TouchableOpacity>
                        ),
                      )}
                    </ScrollView>
                  </View>
                )}

              {/* Batches List */}
              {filteredBatches.length > 0 ? (
                filteredBatches.map((batch) => (
                  <TouchableOpacity
                    key={batch.id}
                    className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
                    onPress={() => navigateToBatch(batch.id)}
                  >
                    {/* Header */}
                    <View className="flex-row justify-between items-start mb-3">
                      <View className="flex-1">
                        <Text className="font-bold text-lg text-gray-800 mb-1">
                          Batch #{batch.id}
                        </Text>
                        <Text className="text-gray-500 text-sm">
                          Created {formatDate(batch.created_at)}
                        </Text>
                      </View>
                      <View
                        className={`px-3 py-1 rounded-full ${getStatusColor(batch.status)}`}
                      >
                        <Text className="text-xs font-medium capitalize">
                          {batch.status}
                        </Text>
                      </View>
                    </View>

                    {/* Stats */}
                    <View className="flex-row flex-wrap mb-3">
                      <View className="w-1/3 mb-2">
                        <Text className="text-gray-500 text-xs">Quantity</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="fish"
                            size={14}
                            color="#f97316"
                          />
                          <Text className="ml-1 font-medium text-sm">
                            {batch.quantity.toLocaleString()}
                          </Text>
                        </View>
                      </View>

                      <View className="w-2/3 mb-2">
                        <Text className="text-gray-500 text-xs">Species</Text>
                        <Text className="font-medium text-sm" numberOfLines={1}>
                          {batch.species}
                        </Text>
                      </View>

                      <View className="w-1/2">
                        <Text className="text-gray-500 text-xs">
                          Last Updated
                        </Text>
                        <Text className="font-medium text-sm">
                          {formatDate(batch.updated_at)}
                        </Text>
                      </View>

                      <View className="w-1/2">
                        <Text className="text-gray-500 text-xs">Status</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name={batch.is_active ? "check" : "x"}
                            size={14}
                            color={batch.is_active ? "#10b981" : "#ef4444"}
                          />
                          <Text
                            className={`text-xs ml-1 ${
                              batch.is_active
                                ? "text-green-600"
                                : "text-red-600"
                            }`}
                          >
                            {batch.is_active ? "Active" : "Inactive"}
                          </Text>
                        </View>
                      </View>
                    </View>

                    {/* Footer */}
                    <View className="flex-row items-center justify-between pt-3 border-t border-gray-100">
                      <Text className="text-gray-500 text-xs">
                        ID: {batch.id} • Updated{" "}
                        {new Date(batch.updated_at).toLocaleDateString()}
                      </Text>
                      <View className="flex-row items-center">
                        <Text className="font-medium text-primary mr-1 text-sm">
                          View Details
                        </Text>
                        <TablerIconComponent
                          name="chevron-right"
                          size={16}
                          color="#f97316"
                        />
                      </View>
                    </View>
                  </TouchableOpacity>
                ))
              ) : (
                <View className="bg-gray-50 p-8 rounded-xl items-center">
                  <TablerIconComponent
                    name="package-off"
                    size={48}
                    color="#9ca3af"
                  />
                  <Text className="text-gray-500 font-medium mt-4 mb-2">
                    {selectedStatus === "all"
                      ? "No batches found"
                      : `No ${selectedStatus} batches`}
                  </Text>
                  <Text className="text-gray-400 text-center">
                    {selectedStatus === "all"
                      ? "This hatchery hasn't created any batches yet"
                      : "Try selecting a different status filter"}
                  </Text>
                  {selectedStatus === "all" && isHatchery && (
                    <TouchableOpacity
                      className="bg-primary px-6 py-3 rounded-xl mt-4"
                      onPress={createNewBatch}
                    >
                      <Text className="text-white font-bold">
                        Create First Batch
                      </Text>
                    </TouchableOpacity>
                  )}
                </View>
              )}
            </View>
          )}

          {activeTab === "analytics" && (
            <View>
              {/* Performance Overview */}
              <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                <Text className="text-lg font-semibold mb-4">
                  Performance Overview
                </Text>

                <View className="flex-row flex-wrap">
                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">Success Rate</Text>
                    <Text className="font-bold text-2xl text-green-600">
                      {stats.successRate}%
                    </Text>
                  </View>

                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">
                      Average Batch Size
                    </Text>
                    <Text className="font-bold text-2xl text-blue-600">
                      {stats.averageQuantity.toLocaleString()}
                    </Text>
                  </View>

                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">
                      Total Production
                    </Text>
                    <Text className="font-bold text-2xl text-orange-600">
                      {stats.totalLarvae.toLocaleString()}
                    </Text>
                  </View>

                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">Active Ratio</Text>
                    <Text className="font-bold text-2xl text-purple-600">
                      {stats.totalBatches > 0
                        ? Math.round(
                            (stats.activeBatches / stats.totalBatches) * 100,
                          )
                        : 0}
                      %
                    </Text>
                  </View>
                </View>
              </View>

              {/* Batch Quantity Trend - Only show if there are batches */}
              {batches.length > 0 && (
                <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                  <Text className="text-lg font-semibold mb-4">
                    Batch Quantity Trend
                  </Text>
                  <LineChart
                    data={getQuantityChartData()}
                    width={screenWidth - 70}
                    height={220}
                    chartConfig={chartConfig}
                    bezier
                    style={{
                      marginVertical: 8,
                      borderRadius: 16,
                    }}
                  />
                </View>
              )}

              {/* Status Distribution - Only show if there are batches */}
              {Object.keys(statusDistribution).length > 0 && (
                <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                  <Text className="text-lg font-semibold mb-4">
                    Status Distribution
                  </Text>

                  {Object.entries(statusDistribution).map(([status, count]) => (
                    <View key={status} className="mb-3 last:mb-0">
                      <View className="flex-row justify-between items-center mb-1">
                        <Text className="font-medium capitalize">{status}</Text>
                        <Text className="text-gray-600">{count} batches</Text>
                      </View>
                      <View className="bg-gray-200 rounded-full h-2">
                        <View
                          className={`h-2 rounded-full ${
                            status === "created"
                              ? "bg-blue-500"
                              : status === "active"
                                ? "bg-green-500"
                                : status === "completed"
                                  ? "bg-gray-500"
                                  : "bg-red-500"
                          }`}
                          style={{
                            width: `${(count / stats.totalBatches) * 100}%`,
                          }}
                        />
                      </View>
                    </View>
                  ))}
                </View>
              )}

              {/* No Data State for Analytics */}
              {batches.length === 0 && (
                <View className="bg-gray-50 p-8 rounded-xl items-center">
                  <TablerIconComponent
                    name="chart-bar-off"
                    size={48}
                    color="#9ca3af"
                  />
                  <Text className="text-gray-500 font-medium mt-4 mb-2">
                    No Analytics Data
                  </Text>
                  <Text className="text-gray-400 text-center">
                    Analytics will be available once batches are created
                  </Text>
                  {isHatchery && (
                    <TouchableOpacity
                      className="bg-primary px-6 py-3 rounded-xl mt-4"
                      onPress={createNewBatch}
                    >
                      <Text className="text-white font-bold">
                        Create First Batch
                      </Text>
                    </TouchableOpacity>
                  )}
                </View>
              )}

              {/* Timeline Summary */}
              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm">
                <Text className="text-lg font-semibold mb-4">
                  Timeline Summary
                </Text>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">
                    Hatchery Established
                  </Text>
                  <Text className="font-medium">
                    {formatDate(hatchery.created_at)}
                  </Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">First Batch</Text>
                  <Text className="font-medium">
                    {batches.length > 0
                      ? formatDate(
                          Math.min(
                            ...batches.map((b) =>
                              new Date(b.created_at).getTime(),
                            ),
                          ).toString(),
                        )
                      : "No batches yet"}
                  </Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Latest Batch</Text>
                  <Text className="font-medium">
                    {batches.length > 0
                      ? formatDate(
                          Math.max(
                            ...batches.map((b) =>
                              new Date(b.created_at).getTime(),
                            ),
                          ).toString(),
                        )
                      : "No batches yet"}
                  </Text>
                </View>

                <View>
                  <Text className="text-gray-600 text-sm">
                    Days in Operation
                  </Text>
                  <Text className="font-medium">
                    {Math.floor(
                      (new Date().getTime() -
                        new Date(hatchery.created_at).getTime()) /
                        (1000 * 60 * 60 * 24),
                    )}{" "}
                    days
                  </Text>
                </View>
              </View>
            </View>
          )}

          {/* Quick Actions */}
          <View className="mt-6">
            <Text className="text-lg font-semibold mb-4">Quick Actions</Text>
            <View className="flex-row flex-wrap gap-3">
              {isHatchery && (
                <TouchableOpacity
                  className="flex-1 bg-primary/10 p-4 rounded-xl items-center min-w-[45%]"
                  onPress={createNewBatch}
                >
                  <TablerIconComponent name="plus" size={24} color="#f97316" />
                  <Text className="text-primary font-medium mt-2">
                    New Batch
                  </Text>
                </TouchableOpacity>
              )}

              <TouchableOpacity
                className="flex-1 bg-blue-50 p-4 rounded-xl items-center min-w-[45%]"
                onPress={handleRefresh}
              >
                <TablerIconComponent name="refresh" size={24} color="#3b82f6" />
                <Text className="text-blue-600 font-medium mt-2">Refresh</Text>
              </TouchableOpacity>

              <TouchableOpacity className="flex-1 bg-green-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent
                  name="download"
                  size={24}
                  color="#10b981"
                />
                <Text className="text-green-600 font-medium mt-2">Export</Text>
              </TouchableOpacity>

              <TouchableOpacity
                className="flex-1 bg-indigo-50 p-4 rounded-xl items-center min-w-[45%]"
                onPress={() => router.push("/(tabs)/(batches)")}
              >
                <TablerIconComponent name="package" size={24} color="#4338ca" />
                <Text className="text-indigo-600 font-medium mt-2">
                  All Batches
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>

      {/* Name Edit Modal */}
      <Modal
        animationType="slide"
        transparent={true}
        visible={isEditingName}
        onRequestClose={handleCancelEdit}
      >
        <KeyboardAvoidingView
          className="flex-1 bg-black/50 justify-center"
          behavior={Platform.OS === "ios" ? "padding" : "height"}
        >
          <View className="bg-white rounded-3xl mx-5 p-6">
            <Text className="text-xl font-bold text-center mb-6">
              Edit Hatchery Name
            </Text>

            <View className="mb-6">
              <Text className="font-medium text-gray-700 mb-2">
                Hatchery Name
              </Text>
              <TextInput
                className="p-4 border border-gray-300 rounded-xl bg-white"
                placeholder="Enter hatchery name"
                value={editedName}
                onChangeText={setEditedName}
                autoFocus
                editable={!isUpdatingName}
              />
            </View>

            <View className="flex-row gap-3">
              <TouchableOpacity
                className="flex-1 bg-gray-100 py-4 rounded-xl items-center"
                onPress={handleCancelEdit}
                disabled={isUpdatingName}
              >
                <Text className="font-bold text-gray-700">Cancel</Text>
              </TouchableOpacity>

              <TouchableOpacity
                className={`flex-1 py-4 rounded-xl items-center ${
                  isUpdatingName ||
                  !editedName.trim() ||
                  editedName.trim() === hatchery.name
                    ? "bg-primary/40"
                    : "bg-primary"
                }`}
                onPress={handleSaveName}
                disabled={
                  isUpdatingName ||
                  !editedName.trim() ||
                  editedName.trim() === hatchery.name
                }
              >
                {isUpdatingName ? (
                  <View className="flex-row items-center">
                    <ActivityIndicator color="white" size="small" />
                    <Text className="font-bold text-white ml-2">Saving...</Text>
                  </View>
                ) : (
                  <Text className="font-bold text-white">Save</Text>
                )}
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      </Modal>
    </SafeAreaView>
  );
}
