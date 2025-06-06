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
import { useRouter } from "expo-router";
import { useRole } from "@/contexts/RoleContext";
import "@/global.css";

interface Batch {
  id: number;
  batchId: string;
  hatcheryId: number;
  hatcheryName: string;
  stage:
    | "Breeding"
    | "Larvae"
    | "Post-Larvae"
    | "Ready"
    | "Completed"
    | "Failed";
  species: string;
  quantity: number;
  startDate: string;
  estimatedCompletion?: string;
  actualCompletion?: string;
  temperature: number;
  ph: number;
  salinity: number;
  status: "Active" | "Completed" | "Failed" | "On Hold";
  progress: number; // 0-100
  manager: string;
  notes?: string;
  blockchainVerified: boolean;
  nftMinted: boolean;
  contractAddress?: string;
  qrCode?: string;
  lastUpdated: string;
}

export default function BatchesScreen() {
  const [batches, setBatches] = useState<Batch[]>([]);
  const [filteredBatches, setFilteredBatches] = useState<Batch[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedStage, setSelectedStage] = useState("all");
  const [selectedStatus, setSelectedStatus] = useState("all");
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [viewMode, setViewMode] = useState<"grid" | "list">("list");

  const router = useRouter();
  const { userData, getCompanyId } = useRole();

  // Mock data - In real app, this would come from API
  const mockBatches: Batch[] = [
    {
      id: 1,
      batchId: "SH-2023-11-H001",
      hatcheryId: 1,
      hatcheryName: "Main Breeding Facility",
      stage: "Post-Larvae",
      species: "Penaeus vannamei",
      quantity: 50000,
      startDate: "2023-10-01",
      estimatedCompletion: "2023-11-15",
      temperature: 28.5,
      ph: 7.2,
      salinity: 15,
      status: "Active",
      progress: 75,
      manager: "Nguyen Van A",
      notes: "Excellent growth rate, above average survival",
      blockchainVerified: true,
      nftMinted: true,
      contractAddress: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
      qrCode: "QR_SH2023_H001",
      lastUpdated: "2023-10-20T14:30:00Z",
    },
    {
      id: 2,
      batchId: "SH-2023-11-H002",
      hatcheryId: 1,
      hatcheryName: "Main Breeding Facility",
      stage: "Larvae",
      species: "Penaeus vannamei",
      quantity: 75000,
      startDate: "2023-10-10",
      estimatedCompletion: "2023-11-25",
      temperature: 29.0,
      ph: 7.4,
      salinity: 16,
      status: "Active",
      progress: 45,
      manager: "Nguyen Van A",
      notes: "Monitoring feeding patterns closely",
      blockchainVerified: true,
      nftMinted: false,
      contractAddress: "0x3a4e813ea3bf9913613ee7a1bea26e02e85f9ea9",
      lastUpdated: "2023-10-20T10:15:00Z",
    },
    {
      id: 3,
      batchId: "SH-2023-10-H015",
      hatcheryId: 2,
      hatcheryName: "Secondary Hatchery",
      stage: "Completed",
      species: "Penaeus vannamei",
      quantity: 45000,
      startDate: "2023-09-01",
      actualCompletion: "2023-10-18",
      temperature: 28.0,
      ph: 7.1,
      salinity: 14,
      status: "Completed",
      progress: 100,
      manager: "Tran Thi B",
      notes: "Successful batch, 92% survival rate",
      blockchainVerified: true,
      nftMinted: true,
      contractAddress: "0x7b91b7c1d8b9a89c8a65e06e4a4f8f0c9c6f6d4c",
      qrCode: "QR_SH2023_H015",
      lastUpdated: "2023-10-18T16:45:00Z",
    },
    {
      id: 4,
      batchId: "SH-2023-11-H003",
      hatcheryId: 3,
      hatcheryName: "Research & Development Center",
      stage: "Breeding",
      species: "Penaeus monodon",
      quantity: 25000,
      startDate: "2023-10-15",
      estimatedCompletion: "2023-12-01",
      temperature: 27.5,
      ph: 7.0,
      salinity: 18,
      status: "On Hold",
      progress: 15,
      manager: "Le Van C",
      notes: "Research batch - testing new breeding techniques",
      blockchainVerified: false,
      nftMinted: false,
      lastUpdated: "2023-10-19T09:20:00Z",
    },
    {
      id: 5,
      batchId: "SH-2023-10-H012",
      hatcheryId: 2,
      hatcheryName: "Secondary Hatchery",
      stage: "Failed",
      species: "Penaeus vannamei",
      quantity: 30000,
      startDate: "2023-09-20",
      actualCompletion: "2023-10-10",
      temperature: 30.5,
      ph: 6.8,
      salinity: 20,
      status: "Failed",
      progress: 35,
      manager: "Tran Thi B",
      notes: "Temperature spike caused mortality, investigating cause",
      blockchainVerified: true,
      nftMinted: false,
      contractAddress: "0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0",
      lastUpdated: "2023-10-10T12:30:00Z",
    },
    {
      id: 6,
      batchId: "SH-2023-11-H004",
      hatcheryId: 4,
      hatcheryName: "Coastal Breeding Station",
      stage: "Ready",
      species: "Penaeus vannamei",
      quantity: 60000,
      startDate: "2023-09-15",
      estimatedCompletion: "2023-10-30",
      temperature: 28.8,
      ph: 7.3,
      salinity: 15,
      status: "Active",
      progress: 95,
      manager: "Pham Thi D",
      notes: "Ready for harvest, excellent quality",
      blockchainVerified: true,
      nftMinted: true,
      contractAddress: "0xd9e8f7c6b5a4d3c2b1a0f9e8d7c6b5a4d3c2b1a0",
      qrCode: "QR_SH2023_H004",
      lastUpdated: "2023-10-20T11:00:00Z",
    },
  ];

  const loadBatches = async () => {
    setIsLoading(true);
    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setBatches(mockBatches);
      setFilteredBatches(mockBatches);
    } catch (error) {
      console.error("Error loading batches:", error);
      Alert.alert("Error", "Failed to load batches");
    } finally {
      setIsLoading(false);
    }
  };

  const refreshBatches = async () => {
    setIsRefreshing(true);
    try {
      // Simulate API refresh
      await new Promise((resolve) => setTimeout(resolve, 1500));
      setBatches(mockBatches);
      applyFilters(mockBatches, searchQuery, selectedStage, selectedStatus);
    } catch (error) {
      console.error("Error refreshing batches:", error);
    } finally {
      setIsRefreshing(false);
    }
  };

  const applyFilters = (
    batchList: Batch[],
    search: string,
    stage: string,
    status: string,
  ) => {
    let filtered = batchList;

    // Apply stage filter
    if (stage !== "all") {
      filtered = filtered.filter((b) => b.stage.toLowerCase() === stage);
    }

    // Apply status filter
    if (status !== "all") {
      filtered = filtered.filter((b) => b.status.toLowerCase() === status);
    }

    // Apply search filter
    if (search) {
      filtered = filtered.filter(
        (b) =>
          b.batchId.toLowerCase().includes(search.toLowerCase()) ||
          b.hatcheryName.toLowerCase().includes(search.toLowerCase()) ||
          b.species.toLowerCase().includes(search.toLowerCase()) ||
          b.manager.toLowerCase().includes(search.toLowerCase()),
      );
    }

    setFilteredBatches(filtered);
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    applyFilters(batches, query, selectedStage, selectedStatus);
  };

  const handleStageFilter = (stage: string) => {
    setSelectedStage(stage);
    applyFilters(batches, searchQuery, stage, selectedStatus);
  };

  const handleStatusFilter = (status: string) => {
    setSelectedStatus(status);
    applyFilters(batches, searchQuery, selectedStage, status);
  };

  const navigateToBatch = (batchId: number) => {
    router.push(`/batch/${batchId}`);
  };

  const createNewBatch = () => {
    router.push("/batch/create");
  };

  const getStageColor = (stage: string) => {
    switch (stage) {
      case "Breeding":
        return "bg-purple-100 text-purple-700";
      case "Larvae":
        return "bg-blue-100 text-blue-700";
      case "Post-Larvae":
        return "bg-cyan-100 text-cyan-700";
      case "Ready":
        return "bg-green-100 text-green-700";
      case "Completed":
        return "bg-gray-100 text-gray-700";
      case "Failed":
        return "bg-red-100 text-red-700";
      default:
        return "bg-gray-100 text-gray-700";
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "Active":
        return "bg-green-100 text-green-700";
      case "Completed":
        return "bg-blue-100 text-blue-700";
      case "Failed":
        return "bg-red-100 text-red-700";
      case "On Hold":
        return "bg-yellow-100 text-yellow-700";
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

  const getDaysRemaining = (estimatedCompletion?: string) => {
    if (!estimatedCompletion) return null;
    const today = new Date();
    const completion = new Date(estimatedCompletion);
    const diffTime = completion.getTime() - today.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  };

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
                    Active Batches
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-blue-800">
                  {batches.filter((b) => b.status === "Active").length}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2 mb-4">
              <View className="bg-green-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="check-circle"
                    size={20}
                    color="#10b981"
                  />
                  <Text className="text-green-700 font-medium ml-2">
                    Completed
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-green-800">
                  {batches.filter((b) => b.status === "Completed").length}
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
                  {batches
                    .filter((b) => b.status === "Active")
                    .reduce((sum, b) => sum + b.quantity, 0)
                    .toLocaleString()}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2">
              <View className="bg-indigo-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="shield-check"
                    size={20}
                    color="#4338ca"
                  />
                  <Text className="text-indigo-700 font-medium ml-2">
                    Verified
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-indigo-800">
                  {batches.filter((b) => b.blockchainVerified).length}
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
                placeholder="Search batches, hatcheries, or species..."
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
              Stage
            </Text>
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
              {[
                "all",
                "breeding",
                "larvae",
                "post-larvae",
                "ready",
                "completed",
                "failed",
              ].map((stage) => (
                <TouchableOpacity
                  key={stage}
                  className={`px-4 py-2 rounded-full mr-3 ${
                    selectedStage === stage ? "bg-primary" : "bg-gray-100"
                  }`}
                  onPress={() => handleStageFilter(stage)}
                >
                  <Text
                    className={`font-medium capitalize ${
                      selectedStage === stage ? "text-white" : "text-gray-600"
                    }`}
                  >
                    {stage === "post-larvae" ? "Post-Larvae" : stage}
                  </Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          </View>

          <View className="mb-6">
            <Text className="text-sm font-medium text-gray-700 mb-2">
              Status
            </Text>
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
              {["all", "active", "completed", "failed", "on hold"].map(
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

          {/* Batches List/Grid */}
          {filteredBatches.length === 0 ? (
            <View className="bg-gray-50 p-8 rounded-xl items-center">
              <TablerIconComponent name="package" size={48} color="#9ca3af" />
              <Text className="text-gray-500 font-medium mt-4 mb-2">
                No batches found
              </Text>
              <Text className="text-gray-400 text-center">
                {searchQuery ||
                selectedStage !== "all" ||
                selectedStatus !== "all"
                  ? "Try adjusting your search or filters"
                  : "Create your first batch to get started"}
              </Text>
              {!searchQuery &&
                selectedStage === "all" &&
                selectedStatus === "all" && (
                  <TouchableOpacity
                    className="bg-primary px-6 py-3 rounded-xl mt-4"
                    onPress={createNewBatch}
                  >
                    <Text className="text-white font-bold">Create Batch</Text>
                  </TouchableOpacity>
                )}
            </View>
          ) : viewMode === "list" ? (
            // List View
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
                      {batch.batchId}
                    </Text>
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="building-factory-2"
                        size={14}
                        color="#9ca3af"
                      />
                      <Text className="text-gray-500 text-sm ml-1">
                        {batch.hatcheryName}
                      </Text>
                    </View>
                  </View>
                  <View className="flex-row">
                    <View
                      className={`px-3 py-1 rounded-full mr-2 ${getStageColor(
                        batch.stage,
                      )}`}
                    >
                      <Text className="text-xs font-medium">{batch.stage}</Text>
                    </View>
                    <View
                      className={`px-3 py-1 rounded-full ${getStatusColor(
                        batch.status,
                      )}`}
                    >
                      <Text className="text-xs font-medium">
                        {batch.status}
                      </Text>
                    </View>
                  </View>
                </View>

                {/* Progress Bar */}
                <View className="mb-3">
                  <View className="flex-row justify-between items-center mb-1">
                    <Text className="text-gray-600 text-sm">Progress</Text>
                    <Text className="text-gray-800 font-medium text-sm">
                      {batch.progress}%
                    </Text>
                  </View>
                  <View className="h-2 bg-gray-200 rounded-full overflow-hidden">
                    <View
                      className={`h-full rounded-full ${
                        batch.status === "Failed"
                          ? "bg-red-500"
                          : batch.status === "Completed"
                            ? "bg-green-500"
                            : "bg-primary"
                      }`}
                      style={{ width: `${batch.progress}%` }}
                    />
                  </View>
                </View>

                {/* Stats Grid */}
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

                  <View className="w-1/3 mb-2">
                    <Text className="text-gray-500 text-xs">Species</Text>
                    <Text className="font-medium text-sm" numberOfLines={1}>
                      {batch.species.split(" ")[1] || batch.species}
                    </Text>
                  </View>

                  <View className="w-1/3 mb-2">
                    <Text className="text-gray-500 text-xs">Manager</Text>
                    <Text className="font-medium text-sm" numberOfLines={1}>
                      {batch.manager}
                    </Text>
                  </View>

                  <View className="w-1/2">
                    <Text className="text-gray-500 text-xs">Started</Text>
                    <Text className="font-medium text-sm">
                      {formatDate(batch.startDate)}
                    </Text>
                  </View>

                  <View className="w-1/2">
                    {batch.status === "Active" && batch.estimatedCompletion ? (
                      <>
                        <Text className="text-gray-500 text-xs">Days Left</Text>
                        <Text className="font-medium text-sm">
                          {getDaysRemaining(batch.estimatedCompletion)} days
                        </Text>
                      </>
                    ) : batch.actualCompletion ? (
                      <>
                        <Text className="text-gray-500 text-xs">Completed</Text>
                        <Text className="font-medium text-sm">
                          {formatDate(batch.actualCompletion)}
                        </Text>
                      </>
                    ) : null}
                  </View>
                </View>

                {/* Footer */}
                <View className="flex-row items-center justify-between pt-3 border-t border-gray-100">
                  <View className="flex-row items-center">
                    {batch.blockchainVerified && (
                      <View className="flex-row items-center mr-3">
                        <TablerIconComponent
                          name="shield-check"
                          size={14}
                          color="#10b981"
                        />
                        <Text className="text-green-600 text-xs ml-1">
                          Verified
                        </Text>
                      </View>
                    )}
                    {batch.nftMinted && (
                      <View className="flex-row items-center mr-3">
                        <TablerIconComponent
                          name="certificate"
                          size={14}
                          color="#4338ca"
                        />
                        <Text className="text-indigo-600 text-xs ml-1">
                          NFT
                        </Text>
                      </View>
                    )}
                    <Text className="text-gray-500 text-xs">
                      Updated {new Date(batch.lastUpdated).toLocaleDateString()}
                    </Text>
                  </View>
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
            ))
          ) : (
            // Grid View
            <View className="flex-row flex-wrap -mx-2">
              {filteredBatches.map((batch) => (
                <TouchableOpacity
                  key={batch.id}
                  className="w-1/2 px-2 mb-4"
                  onPress={() => navigateToBatch(batch.id)}
                >
                  <View className="bg-white border border-gray-200 rounded-xl p-3 shadow-sm">
                    <View className="flex-row justify-between items-start mb-2">
                      <Text
                        className="font-bold text-sm text-gray-800 flex-1"
                        numberOfLines={1}
                      >
                        {batch.batchId}
                      </Text>
                      {batch.blockchainVerified && (
                        <TablerIconComponent
                          name="shield-check"
                          size={16}
                          color="#10b981"
                        />
                      )}
                    </View>

                    <View
                      className={`self-start px-2 py-1 rounded-full mb-2 ${getStageColor(
                        batch.stage,
                      )}`}
                    >
                      <Text className="text-xs font-medium">{batch.stage}</Text>
                    </View>

                    <Text className="text-gray-500 text-xs mb-1">
                      {batch.hatcheryName}
                    </Text>

                    <View className="flex-row items-center mb-2">
                      <TablerIconComponent
                        name="fish"
                        size={12}
                        color="#f97316"
                      />
                      <Text className="ml-1 text-xs font-medium">
                        {batch.quantity.toLocaleString()}
                      </Text>
                    </View>

                    <View className="mb-2">
                      <View className="h-1.5 bg-gray-200 rounded-full overflow-hidden">
                        <View
                          className={`h-full rounded-full ${
                            batch.status === "Failed"
                              ? "bg-red-500"
                              : batch.status === "Completed"
                                ? "bg-green-500"
                                : "bg-primary"
                          }`}
                          style={{ width: `${batch.progress}%` }}
                        />
                      </View>
                      <Text className="text-xs text-gray-500 mt-1">
                        {batch.progress}% complete
                      </Text>
                    </View>

                    <View
                      className={`self-start px-2 py-1 rounded ${getStatusColor(
                        batch.status,
                      )}`}
                    >
                      <Text className="text-xs font-medium">
                        {batch.status}
                      </Text>
                    </View>
                  </View>
                </TouchableOpacity>
              ))}
            </View>
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

              <TouchableOpacity className="flex-1 bg-indigo-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent name="qrcode" size={24} color="#4338ca" />
                <Text className="text-indigo-600 font-medium mt-2">
                  QR Codes
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
