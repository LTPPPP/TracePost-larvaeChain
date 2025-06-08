import React, { useState, useEffect } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  TextInput,
  ActivityIndicator,
  Alert,
  RefreshControl,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import TablerIconComponent from "@/components/icon";
import { useRouter } from "expo-router";
import { useRole } from "@/contexts/RoleContext";
import { getAllBatches, getBatchesByHatchery, BatchData } from "@/api/batch";
import { getHatcheries } from "@/api/hatchery";
import "@/global.css";

// Interface for grouped batches
interface GroupedBatches {
  [hatcheryId: string]: {
    hatchery: BatchData["hatchery"];
    batches: BatchData[];
  };
}

export default function BatchesScreen() {
  const [batches, setBatches] = useState<BatchData[]>([]);
  const [groupedBatches, setGroupedBatches] = useState<GroupedBatches>({});
  const [filteredGroupedBatches, setFilteredGroupedBatches] =
    useState<GroupedBatches>({});
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedStatus, setSelectedStatus] = useState("all");
  const [selectedHatchery, setSelectedHatchery] = useState("all");
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [viewMode, setViewMode] = useState<"grid" | "list">("list");
  const [availableHatcheries, setAvailableHatcheries] = useState<string[]>([]);

  const router = useRouter();
  const { userData, getCompanyId } = useRole();

  const loadBatches = async () => {
    setIsLoading(true);
    try {
      const response = await getAllBatches();

      if (response.success) {
        setBatches(response.data);

        // Group batches by hatchery
        const grouped = groupBatchesByHatchery(response.data);
        setGroupedBatches(grouped);
        setFilteredGroupedBatches(grouped);

        // Extract unique hatchery names for filter
        const hatcheryNames = Object.values(grouped).map(
          (group) => group.hatchery.name,
        );
        setAvailableHatcheries(hatcheryNames);
      } else {
        throw new Error(response.message);
      }
    } catch (error) {
      console.error("Error loading batches:", error);
      Alert.alert(
        "Error",
        error instanceof Error ? error.message : "Failed to load batches",
      );
    } finally {
      setIsLoading(false);
    }
  };

  const refreshBatches = async () => {
    setIsRefreshing(true);
    try {
      const response = await getAllBatches();

      if (response.success) {
        setBatches(response.data);
        const grouped = groupBatchesByHatchery(response.data);
        setGroupedBatches(grouped);
        applyFilters(grouped, searchQuery, selectedStatus, selectedHatchery);

        // Update available hatcheries
        const hatcheryNames = Object.values(grouped).map(
          (group) => group.hatchery.name,
        );
        setAvailableHatcheries(hatcheryNames);
      } else {
        throw new Error(response.message);
      }
    } catch (error) {
      console.error("Error refreshing batches:", error);
    } finally {
      setIsRefreshing(false);
    }
  };

  // Group batches by hatchery
  const groupBatchesByHatchery = (batchList: BatchData[]): GroupedBatches => {
    const grouped: GroupedBatches = {};

    batchList.forEach((batch) => {
      const hatcheryId = batch.hatchery_id.toString();

      if (!grouped[hatcheryId]) {
        grouped[hatcheryId] = {
          hatchery: batch.hatchery,
          batches: [],
        };
      }

      grouped[hatcheryId].batches.push(batch);
    });

    // Sort batches within each group by creation date (newest first)
    Object.keys(grouped).forEach((hatcheryId) => {
      grouped[hatcheryId].batches.sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
      );
    });

    return grouped;
  };

  const applyFilters = (
    grouped: GroupedBatches,
    search: string,
    status: string,
    hatchery: string,
  ) => {
    let filteredGrouped: GroupedBatches = {};

    Object.keys(grouped).forEach((hatcheryId) => {
      let filteredBatches = grouped[hatcheryId].batches;

      // Apply hatchery filter
      if (
        hatchery !== "all" &&
        grouped[hatcheryId].hatchery.name !== hatchery
      ) {
        return; // Skip this hatchery group
      }

      // Apply status filter
      if (status !== "all") {
        filteredBatches = filteredBatches.filter(
          (b) => b.status.toLowerCase() === status,
        );
      }

      // Apply search filter
      if (search) {
        filteredBatches = filteredBatches.filter(
          (b) =>
            b.id.toString().includes(search) ||
            b.species.toLowerCase().includes(search.toLowerCase()) ||
            b.hatchery.name.toLowerCase().includes(search.toLowerCase()),
        );
      }

      // Only include the group if it has batches after filtering
      if (filteredBatches.length > 0) {
        filteredGrouped[hatcheryId] = {
          hatchery: grouped[hatcheryId].hatchery,
          batches: filteredBatches,
        };
      }
    });

    setFilteredGroupedBatches(filteredGrouped);
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    applyFilters(groupedBatches, query, selectedStatus, selectedHatchery);
  };

  const handleStatusFilter = (status: string) => {
    setSelectedStatus(status);
    applyFilters(groupedBatches, searchQuery, status, selectedHatchery);
  };

  const handleHatcheryFilter = (hatchery: string) => {
    setSelectedHatchery(hatchery);
    applyFilters(groupedBatches, searchQuery, selectedStatus, hatchery);
  };

  const navigateToBatch = (batchId: number) => {
    router.push(`/(tabs)/(batches)/${batchId}`);
  };

  const createNewBatch = () => {
    router.push("/(tabs)/(batches)/create");
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  // Calculate statistics
  const totalBatches = batches.length;
  const activeBatches = batches.filter((b) => b.is_active).length;
  const totalLarvae = batches.reduce((sum, b) => sum + b.quantity, 0);
  const uniqueHatcheries = Object.keys(groupedBatches).length;

  useEffect(() => {
    loadBatches();
  }, []);

  if (isLoading) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <View className="flex-1 justify-center items-center">
          <ActivityIndicator size="large" color="#f97316" />
          <Text className="text-gray-500 mt-4">Loading batches...</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView className="flex-1 bg-white">
      <ScrollView
        contentContainerStyle={{ paddingBottom: 100 }}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={refreshBatches}
            colors={["#f97316"]}
            tintColor="#f97316"
          />
        }
      >
        <View className="px-5 pt-4 pb-6">
          {/* Header */}
          <View className="flex-row items-center justify-between mb-6">
            <View>
              <Text className="text-2xl font-bold text-gray-800">
                Batch Management
              </Text>
              <Text className="text-gray-500">
                Track and manage your breeding batches
              </Text>
              {userData && (
                <Text className="text-xs text-gray-400 mt-1">
                  {userData.username} • Company #{getCompanyId()}
                </Text>
              )}
            </View>
            <View className="flex-row">
              <TouchableOpacity
                className={`h-10 w-10 rounded-full mr-2 items-center justify-center ${
                  viewMode === "list" ? "bg-primary" : "bg-gray-100"
                }`}
                onPress={() => setViewMode("list")}
              >
                <TablerIconComponent
                  name="list"
                  size={20}
                  color={viewMode === "list" ? "white" : "#9ca3af"}
                />
              </TouchableOpacity>
              <TouchableOpacity
                className={`h-10 w-10 rounded-full mr-2 items-center justify-center ${
                  viewMode === "grid" ? "bg-primary" : "bg-gray-100"
                }`}
                onPress={() => setViewMode("grid")}
              >
                <TablerIconComponent
                  name="grid-3x3"
                  size={20}
                  color={viewMode === "grid" ? "white" : "#9ca3af"}
                />
              </TouchableOpacity>
              <TouchableOpacity
                className="h-10 w-10 rounded-full bg-primary items-center justify-center"
                onPress={createNewBatch}
              >
                <TablerIconComponent name="plus" size={20} color="white" />
              </TouchableOpacity>
            </View>
          </View>

          {/* Stats Overview */}
          <View className="flex-row flex-wrap mb-6">
            <View className="w-1/2 pr-2 mb-4">
              <View className="bg-blue-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="package"
                    size={20}
                    color="#3b82f6"
                  />
                  <Text className="text-blue-700 font-medium ml-2">
                    Total Batches
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-blue-800">
                  {totalBatches}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2 mb-4">
              <View className="bg-green-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent name="check" size={20} color="#10b981" />
                  <Text className="text-green-700 font-medium ml-2">
                    Active
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-green-800">
                  {activeBatches}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pr-2">
              <View className="bg-orange-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent name="fish" size={20} color="#f97316" />
                  <Text className="text-orange-700 font-medium ml-2">
                    Total Larvae
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-orange-800">
                  {totalLarvae.toLocaleString()}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2">
              <View className="bg-indigo-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="building-factory-2"
                    size={20}
                    color="#4338ca"
                  />
                  <Text className="text-indigo-700 font-medium ml-2">
                    Hatcheries
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-indigo-800">
                  {uniqueHatcheries}
                </Text>
              </View>
            </View>
          </View>

          {/* Search Bar */}
          <View className="mb-4">
            <View className="flex-row bg-gray-100 rounded-xl p-3 items-center">
              <TablerIconComponent name="search" size={20} color="#9ca3af" />
              <TextInput
                className="flex-1 ml-3 text-gray-700"
                placeholder="Search batches, species, or hatcheries..."
                value={searchQuery}
                onChangeText={handleSearch}
              />
              {searchQuery ? (
                <TouchableOpacity onPress={() => handleSearch("")}>
                  <TablerIconComponent name="x" size={20} color="#9ca3af" />
                </TouchableOpacity>
              ) : null}
            </View>
          </View>

          {/* Filter Buttons */}
          <View className="mb-4">
            <Text className="text-sm font-medium text-gray-700 mb-2">
              Filter by Status
            </Text>
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
              {["all", "created", "active", "completed", "failed"].map(
                (status) => (
                  <TouchableOpacity
                    key={status}
                    className={`px-4 py-2 rounded-full mr-3 ${
                      selectedStatus === status ? "bg-secondary" : "bg-gray-100"
                    }`}
                    onPress={() => handleStatusFilter(status)}
                  >
                    <Text
                      className={`font-medium capitalize ${
                        selectedStatus === status
                          ? "text-white"
                          : "text-gray-600"
                      }`}
                    >
                      {status}
                    </Text>
                  </TouchableOpacity>
                ),
              )}
            </ScrollView>
          </View>

          {/* Hatchery Filter */}
          {availableHatcheries.length > 1 && (
            <View className="mb-6">
              <Text className="text-sm font-medium text-gray-700 mb-2">
                Filter by Hatchery
              </Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false}>
                <TouchableOpacity
                  className={`px-4 py-2 rounded-full mr-3 ${
                    selectedHatchery === "all" ? "bg-primary" : "bg-gray-100"
                  }`}
                  onPress={() => handleHatcheryFilter("all")}
                >
                  <Text
                    className={`font-medium ${
                      selectedHatchery === "all"
                        ? "text-white"
                        : "text-gray-600"
                    }`}
                  >
                    All Hatcheries
                  </Text>
                </TouchableOpacity>
                {availableHatcheries.map((hatcheryName) => (
                  <TouchableOpacity
                    key={hatcheryName}
                    className={`px-4 py-2 rounded-full mr-3 ${
                      selectedHatchery === hatcheryName
                        ? "bg-primary"
                        : "bg-gray-100"
                    }`}
                    onPress={() => handleHatcheryFilter(hatcheryName)}
                  >
                    <Text
                      className={`font-medium ${
                        selectedHatchery === hatcheryName
                          ? "text-white"
                          : "text-gray-600"
                      }`}
                    >
                      {hatcheryName}
                    </Text>
                  </TouchableOpacity>
                ))}
              </ScrollView>
            </View>
          )}

          {/* Grouped Batches List */}
          {Object.keys(filteredGroupedBatches).length === 0 ? (
            <View className="bg-gray-50 p-8 rounded-xl items-center">
              <TablerIconComponent name="package" size={48} color="#9ca3af" />
              <Text className="text-gray-500 font-medium mt-4 mb-2">
                No batches found
              </Text>
              <Text className="text-gray-400 text-center">
                {searchQuery ||
                selectedStatus !== "all" ||
                selectedHatchery !== "all"
                  ? "Try adjusting your search or filters"
                  : "Create your first batch to get started"}
              </Text>
              {!searchQuery &&
                selectedStatus === "all" &&
                selectedHatchery === "all" && (
                  <TouchableOpacity
                    className="bg-primary px-6 py-3 rounded-xl mt-4"
                    onPress={createNewBatch}
                  >
                    <Text className="text-white font-bold">Create Batch</Text>
                  </TouchableOpacity>
                )}
            </View>
          ) : (
            Object.keys(filteredGroupedBatches)
              .sort((a, b) => {
                // Sort hatchery groups by hatchery name
                const nameA = filteredGroupedBatches[a].hatchery.name;
                const nameB = filteredGroupedBatches[b].hatchery.name;
                return nameA.localeCompare(nameB);
              })
              .map((hatcheryId) => {
                const group = filteredGroupedBatches[hatcheryId];
                return (
                  <View key={hatcheryId} className="mb-6">
                    {/* Hatchery Header */}
                    <View className="bg-gradient-to-r from-primary/10 to-primary/5 p-4 rounded-xl mb-4 border-l-4 border-primary">
                      <View className="flex-row items-center justify-between">
                        <View className="flex-1">
                          <View className="flex-row items-center mb-2">
                            <TablerIconComponent
                              name="building-factory-2"
                              size={20}
                              color="#f97316"
                            />
                            <Text className="text-lg font-bold text-gray-800 ml-2">
                              {group.hatchery.name}
                            </Text>
                          </View>
                          <View className="flex-row items-center justify-between">
                            <View>
                              <Text className="text-gray-600 text-sm">
                                {group.hatchery.company.name} •{" "}
                                {group.hatchery.company.location}
                              </Text>
                              <Text className="text-gray-500 text-xs">
                                Company Type: {group.hatchery.company.type}
                              </Text>
                            </View>
                            <View className="items-end">
                              <Text className="text-primary font-bold text-lg">
                                {group.batches.length}
                              </Text>
                              <Text className="text-gray-600 text-xs">
                                {group.batches.length === 1
                                  ? "batch"
                                  : "batches"}
                              </Text>
                            </View>
                          </View>
                        </View>
                      </View>

                      {/* Hatchery Stats */}
                      <View className="flex-row items-center mt-3 pt-3 border-t border-primary/20">
                        <View className="flex-1">
                          <Text className="text-gray-600 text-xs">
                            Total Larvae
                          </Text>
                          <Text className="font-bold text-gray-800">
                            {group.batches
                              .reduce((sum, batch) => sum + batch.quantity, 0)
                              .toLocaleString()}
                          </Text>
                        </View>
                        <View className="flex-1">
                          <Text className="text-gray-600 text-xs">
                            Active Batches
                          </Text>
                          <Text className="font-bold text-gray-800">
                            {
                              group.batches.filter((batch) => batch.is_active)
                                .length
                            }
                          </Text>
                        </View>
                        <View className="flex-1">
                          <Text className="text-gray-600 text-xs">
                            Latest Batch
                          </Text>
                          <Text className="font-bold text-gray-800">
                            {formatDate(group.batches[0]?.created_at || "")}
                          </Text>
                        </View>
                      </View>
                    </View>

                    {/* Batches in this Hatchery */}
                    {group.batches.map((batch) => (
                      <TouchableOpacity
                        key={batch.id}
                        className="bg-white border border-gray-200 rounded-xl p-4 mb-3 shadow-sm ml-4"
                        onPress={() => navigateToBatch(batch.id)}
                      >
                        {/* Batch Header */}
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
                            className={`px-3 py-1 rounded-full ${getStatusColor(
                              batch.status,
                            )}`}
                          >
                            <Text className="text-xs font-medium capitalize">
                              {batch.status}
                            </Text>
                          </View>
                        </View>

                        {/* Batch Stats */}
                        <View className="flex-row flex-wrap mb-3">
                          <View className="w-1/3 mb-2">
                            <Text className="text-gray-500 text-xs">
                              Quantity
                            </Text>
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
                            <Text className="text-gray-500 text-xs">
                              Species
                            </Text>
                            <Text
                              className="font-medium text-sm"
                              numberOfLines={1}
                            >
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
                            <Text className="text-gray-500 text-xs">
                              Status
                            </Text>
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
                          <TouchableOpacity className="flex-row items-center">
                            <Text className="font-medium text-primary mr-1 text-sm">
                              Details
                            </Text>
                            <TablerIconComponent
                              name="chevron-right"
                              size={16}
                              color="#f97316"
                            />
                          </TouchableOpacity>
                        </View>
                      </TouchableOpacity>
                    ))}
                  </View>
                );
              })
          )}

          {/* Quick Actions */}
          <View className="mt-6">
            <Text className="text-lg font-semibold mb-4">Quick Actions</Text>
            <View className="flex-row flex-wrap gap-3">
              <TouchableOpacity
                className="flex-1 bg-primary/10 p-4 rounded-xl items-center min-w-[45%]"
                onPress={createNewBatch}
              >
                <TablerIconComponent name="package" size={24} color="#f97316" />
                <Text className="text-primary font-medium mt-2">New Batch</Text>
              </TouchableOpacity>

              <TouchableOpacity className="flex-1 bg-green-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent
                  name="download"
                  size={24}
                  color="#10b981"
                />
                <Text className="text-green-600 font-medium mt-2">
                  Export Data
                </Text>
              </TouchableOpacity>

              <TouchableOpacity className="flex-1 bg-blue-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent
                  name="chart-bar"
                  size={24}
                  color="#3b82f6"
                />
                <Text className="text-blue-600 font-medium mt-2">
                  Analytics
                </Text>
              </TouchableOpacity>

              <TouchableOpacity
                className="flex-1 bg-indigo-50 p-4 rounded-xl items-center min-w-[45%]"
                onPress={() => router.push("/(tabs)/(hatchery)")}
              >
                <TablerIconComponent
                  name="building-factory-2"
                  size={24}
                  color="#4338ca"
                />
                <Text className="text-indigo-600 font-medium mt-2">
                  Hatcheries
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
