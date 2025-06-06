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

interface Hatchery {
  id: number;
  name: string;
  location: string;
  capacity: number;
  currentStock: number;
  activeBatches: number;
  totalBatches: number;
  completedBatches: number;
  status: "Active" | "Maintenance" | "Inactive";
  manager: string;
  phone: string;
  email: string;
  establishedDate: string;
  lastInspection: string;
  certificationStatus: "Valid" | "Expired" | "Pending";
  blockchainVerified: boolean;
  contractAddress?: string;
}

export default function HatcheryScreen() {
  const [hatcheries, setHatcheries] = useState<Hatchery[]>([]);
  const [filteredHatcheries, setFilteredHatcheries] = useState<Hatchery[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedFilter, setSelectedFilter] = useState("all");
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const router = useRouter();
  const { userData, getCompanyId } = useRole();

  // Mock data - In real app, this would come from API
  const mockHatcheries: Hatchery[] = [
    {
      id: 1,
      name: "Main Breeding Facility",
      location: "Mekong Delta, Vietnam",
      capacity: 15000,
      currentStock: 12500,
      activeBatches: 8,
      totalBatches: 24,
      completedBatches: 16,
      status: "Active",
      manager: "Nguyen Van A",
      phone: "+84 123 456 789",
      email: "manager1@hatchery.com",
      establishedDate: "2020-03-15",
      lastInspection: "2023-10-15",
      certificationStatus: "Valid",
      blockchainVerified: true,
      contractAddress: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
    },
    {
      id: 2,
      name: "Secondary Hatchery",
      location: "Can Tho, Vietnam",
      capacity: 10000,
      currentStock: 8200,
      activeBatches: 5,
      totalBatches: 18,
      completedBatches: 13,
      status: "Active",
      manager: "Tran Thi B",
      phone: "+84 987 654 321",
      email: "manager2@hatchery.com",
      establishedDate: "2021-07-20",
      lastInspection: "2023-10-10",
      certificationStatus: "Valid",
      blockchainVerified: true,
      contractAddress: "0x3a4e813ea3bf9913613ee7a1bea26e02e85f9ea9",
    },
    {
      id: 3,
      name: "Research & Development Center",
      location: "Ho Chi Minh City, Vietnam",
      capacity: 5000,
      currentStock: 2100,
      activeBatches: 3,
      totalBatches: 12,
      completedBatches: 9,
      status: "Maintenance",
      manager: "Le Van C",
      phone: "+84 555 123 456",
      email: "research@hatchery.com",
      establishedDate: "2022-01-10",
      lastInspection: "2023-09-30",
      certificationStatus: "Pending",
      blockchainVerified: false,
    },
    {
      id: 4,
      name: "Coastal Breeding Station",
      location: "Phan Thiet, Vietnam",
      capacity: 8000,
      currentStock: 6800,
      activeBatches: 6,
      totalBatches: 15,
      completedBatches: 9,
      status: "Active",
      manager: "Pham Thi D",
      phone: "+84 666 789 012",
      email: "coastal@hatchery.com",
      establishedDate: "2021-11-05",
      lastInspection: "2023-10-12",
      certificationStatus: "Valid",
      blockchainVerified: true,
      contractAddress: "0x7b91b7c1d8b9a89c8a65e06e4a4f8f0c9c6f6d4c",
    },
  ];

  const loadHatcheries = async () => {
    setIsLoading(true);
    try {
      // Simulate API call
      await new Promise((resolve) => setTimeout(resolve, 1000));
      setHatcheries(mockHatcheries);
      setFilteredHatcheries(mockHatcheries);
    } catch (error) {
      console.error("Error loading hatcheries:", error);
      Alert.alert("Error", "Failed to load hatcheries");
    } finally {
      setIsLoading(false);
    }
  };

  const refreshHatcheries = async () => {
    setIsRefreshing(true);
    try {
      // Simulate API refresh
      await new Promise((resolve) => setTimeout(resolve, 1500));
      setHatcheries(mockHatcheries);
      applyFilters(mockHatcheries, searchQuery, selectedFilter);
    } catch (error) {
      console.error("Error refreshing hatcheries:", error);
    } finally {
      setIsRefreshing(false);
    }
  };

  const applyFilters = (
    hatcheryList: Hatchery[],
    search: string,
    filter: string,
  ) => {
    let filtered = hatcheryList;

    // Apply status filter
    if (filter !== "all") {
      filtered = filtered.filter((h) => h.status.toLowerCase() === filter);
    }

    // Apply search filter
    if (search) {
      filtered = filtered.filter(
        (h) =>
          h.name.toLowerCase().includes(search.toLowerCase()) ||
          h.location.toLowerCase().includes(search.toLowerCase()) ||
          h.manager.toLowerCase().includes(search.toLowerCase()),
      );
    }

    setFilteredHatcheries(filtered);
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    applyFilters(hatcheries, query, selectedFilter);
  };

  const handleFilterChange = (filter: string) => {
    setSelectedFilter(filter);
    applyFilters(hatcheries, searchQuery, filter);
  };

  const navigateToHatchery = (hatcheryId: number) => {
    router.push(`/hatchery/${hatcheryId}`);
  };

  const createNewHatchery = () => {
    router.push("/hatchery/create");
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case "Active":
        return "bg-green-100 text-green-700";
      case "Maintenance":
        return "bg-yellow-100 text-yellow-700";
      case "Inactive":
        return "bg-red-100 text-red-700";
      default:
        return "bg-gray-100 text-gray-700";
    }
  };

  const getCertificationColor = (status: string) => {
    switch (status) {
      case "Valid":
        return "bg-green-100 text-green-700";
      case "Pending":
        return "bg-yellow-100 text-yellow-700";
      case "Expired":
        return "bg-red-100 text-red-700";
      default:
        return "bg-gray-100 text-gray-700";
    }
  };

  const calculateUtilization = (current: number, capacity: number) => {
    return Math.round((current / capacity) * 100);
  };

  useEffect(() => {
    loadHatcheries();
  }, []);

  if (isLoading) {
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
      <ScrollView
        contentContainerStyle={{ paddingBottom: 100 }}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={refreshHatcheries}
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
                My Hatcheries
              </Text>
              <Text className="text-gray-500">
                Manage your breeding facilities
              </Text>
              {userData && (
                <Text className="text-xs text-gray-400 mt-1">
                  {userData.username} • Company #{getCompanyId()}
                </Text>
              )}
            </View>
            <TouchableOpacity
              className="h-12 w-12 rounded-full bg-primary items-center justify-center"
              onPress={createNewHatchery}
            >
              <TablerIconComponent name="plus" size={24} color="white" />
            </TouchableOpacity>
          </View>

          {/* Stats Overview */}
          <View className="flex-row flex-wrap mb-6">
            <View className="w-1/2 pr-2 mb-4">
              <View className="bg-green-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="building-factory-2"
                    size={20}
                    color="#10b981"
                  />
                  <Text className="text-green-700 font-medium ml-2">
                    Total Hatcheries
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-green-800">
                  {hatcheries.length}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2 mb-4">
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
                  {hatcheries.reduce((sum, h) => sum + h.activeBatches, 0)}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pr-2">
              <View className="bg-orange-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent name="fish" size={20} color="#f97316" />
                  <Text className="text-orange-700 font-medium ml-2">
                    Total Stock
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-orange-800">
                  {hatcheries
                    .reduce((sum, h) => sum + h.currentStock, 0)
                    .toLocaleString()}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2">
              <View className="bg-purple-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="chart-line"
                    size={20}
                    color="#8b5cf6"
                  />
                  <Text className="text-purple-700 font-medium ml-2">
                    Avg Utilization
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-purple-800">
                  {Math.round(
                    hatcheries.reduce(
                      (sum, h) =>
                        sum + calculateUtilization(h.currentStock, h.capacity),
                      0,
                    ) / hatcheries.length,
                  )}
                  %
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
                placeholder="Search hatcheries, locations, or managers..."
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
          <View className="flex-row mb-6">
            <ScrollView horizontal showsHorizontalScrollIndicator={false}>
              {["all", "active", "maintenance", "inactive"].map((filter) => (
                <TouchableOpacity
                  key={filter}
                  className={`px-4 py-2 rounded-full mr-3 ${
                    selectedFilter === filter ? "bg-primary" : "bg-gray-100"
                  }`}
                  onPress={() => handleFilterChange(filter)}
                >
                  <Text
                    className={`font-medium capitalize ${
                      selectedFilter === filter ? "text-white" : "text-gray-600"
                    }`}
                  >
                    {filter}
                  </Text>
                </TouchableOpacity>
              ))}
            </ScrollView>
          </View>

          {/* Hatcheries List */}
          {filteredHatcheries.length === 0 ? (
            <View className="bg-gray-50 p-8 rounded-xl items-center">
              <TablerIconComponent
                name="building-factory-2"
                size={48}
                color="#9ca3af"
              />
              <Text className="text-gray-500 font-medium mt-4 mb-2">
                No hatcheries found
              </Text>
              <Text className="text-gray-400 text-center">
                {searchQuery || selectedFilter !== "all"
                  ? "Try adjusting your search or filters"
                  : "Create your first hatchery to get started"}
              </Text>
              {!searchQuery && selectedFilter === "all" && (
                <TouchableOpacity
                  className="bg-primary px-6 py-3 rounded-xl mt-4"
                  onPress={createNewHatchery}
                >
                  <Text className="text-white font-bold">Create Hatchery</Text>
                </TouchableOpacity>
              )}
            </View>
          ) : (
            filteredHatcheries.map((hatchery) => (
              <TouchableOpacity
                key={hatchery.id}
                className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
                onPress={() => navigateToHatchery(hatchery.id)}
              >
                {/* Header */}
                <View className="flex-row justify-between items-start mb-3">
                  <View className="flex-1">
                    <Text className="font-bold text-lg text-gray-800 mb-1">
                      {hatchery.name}
                    </Text>
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="map-pin"
                        size={14}
                        color="#9ca3af"
                      />
                      <Text className="text-gray-500 text-sm ml-1">
                        {hatchery.location}
                      </Text>
                    </View>
                  </View>
                  <View className="flex-row">
                    <View
                      className={`px-3 py-1 rounded-full mr-2 ${getStatusColor(
                        hatchery.status,
                      )}`}
                    >
                      <Text className="text-xs font-medium">
                        {hatchery.status}
                      </Text>
                    </View>
                    {hatchery.blockchainVerified && (
                      <View className="bg-indigo-100 p-1 rounded-full">
                        <TablerIconComponent
                          name="shield-check"
                          size={16}
                          color="#4338ca"
                        />
                      </View>
                    )}
                  </View>
                </View>

                {/* Utilization Bar */}
                <View className="mb-3">
                  <View className="flex-row justify-between items-center mb-1">
                    <Text className="text-gray-600 text-sm">
                      Capacity Usage
                    </Text>
                    <Text className="text-gray-800 font-medium text-sm">
                      {calculateUtilization(
                        hatchery.currentStock,
                        hatchery.capacity,
                      )}
                      %
                    </Text>
                  </View>
                  <View className="h-2 bg-gray-200 rounded-full overflow-hidden">
                    <View
                      className="h-full bg-primary rounded-full"
                      style={{
                        width: `${calculateUtilization(
                          hatchery.currentStock,
                          hatchery.capacity,
                        )}%`,
                      }}
                    />
                  </View>
                </View>

                {/* Stats Grid */}
                <View className="flex-row flex-wrap">
                  <View className="w-1/3 mb-2">
                    <Text className="text-gray-500 text-xs">
                      Active Batches
                    </Text>
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="package"
                        size={14}
                        color="#f97316"
                      />
                      <Text className="ml-1 font-medium text-sm">
                        {hatchery.activeBatches}
                      </Text>
                    </View>
                  </View>

                  <View className="w-1/3 mb-2">
                    <Text className="text-gray-500 text-xs">Current Stock</Text>
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="fish"
                        size={14}
                        color="#3b82f6"
                      />
                      <Text className="ml-1 font-medium text-sm">
                        {hatchery.currentStock.toLocaleString()}
                      </Text>
                    </View>
                  </View>

                  <View className="w-1/3 mb-2">
                    <Text className="text-gray-500 text-xs">Manager</Text>
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="user"
                        size={14}
                        color="#10b981"
                      />
                      <Text
                        className="ml-1 font-medium text-sm"
                        numberOfLines={1}
                      >
                        {hatchery.manager}
                      </Text>
                    </View>
                  </View>

                  <View className="w-1/2">
                    <Text className="text-gray-500 text-xs">Certification</Text>
                    <View
                      className={`self-start px-2 py-1 rounded ${getCertificationColor(
                        hatchery.certificationStatus,
                      )}`}
                    >
                      <Text className="text-xs font-medium">
                        {hatchery.certificationStatus}
                      </Text>
                    </View>
                  </View>

                  <View className="w-1/2">
                    <Text className="text-gray-500 text-xs">
                      Last Inspection
                    </Text>
                    <Text className="font-medium text-sm">
                      {new Date(hatchery.lastInspection).toLocaleDateString()}
                    </Text>
                  </View>
                </View>

                {/* Footer */}
                <View className="flex-row items-center justify-between mt-3 pt-3 border-t border-gray-100">
                  <View className="flex-row items-center">
                    <TablerIconComponent
                      name="calendar"
                      size={14}
                      color="#9ca3af"
                    />
                    <Text className="text-gray-500 text-xs ml-1">
                      Est. {new Date(hatchery.establishedDate).getFullYear()}
                    </Text>
                    {hatchery.blockchainVerified && (
                      <>
                        <View className="h-3 w-0.5 bg-gray-200 mx-2" />
                        <TablerIconComponent
                          name="currency-ethereum"
                          size={12}
                          color="#4338ca"
                        />
                        <Text className="text-indigo-600 text-xs ml-1">
                          Verified
                        </Text>
                      </>
                    )}
                  </View>
                  <TouchableOpacity className="flex-row items-center">
                    <Text className="font-medium text-primary mr-1 text-sm">
                      Manage
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
          )}

          {/* Quick Actions */}
          <View className="mt-6">
            <Text className="text-lg font-semibold mb-4">Quick Actions</Text>
            <View className="flex-row flex-wrap gap-3">
              <TouchableOpacity
                className="flex-1 bg-primary/10 p-4 rounded-xl items-center min-w-[45%]"
                onPress={createNewHatchery}
              >
                <TablerIconComponent
                  name="building-factory-2"
                  size={24}
                  color="#f97316"
                />
                <Text className="text-primary font-medium mt-2">
                  New Hatchery
                </Text>
              </TouchableOpacity>

              <TouchableOpacity
                className="flex-1 bg-blue-50 p-4 rounded-xl items-center min-w-[45%]"
                onPress={() => router.push("/(tabs)/(batches)")}
              >
                <TablerIconComponent name="package" size={24} color="#3b82f6" />
                <Text className="text-blue-600 font-medium mt-2">
                  View Batches
                </Text>
              </TouchableOpacity>

              <TouchableOpacity className="flex-1 bg-green-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent
                  name="file-analytics"
                  size={24}
                  color="#10b981"
                />
                <Text className="text-green-600 font-medium mt-2">
                  Analytics
                </Text>
              </TouchableOpacity>

              <TouchableOpacity className="flex-1 bg-purple-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent
                  name="settings"
                  size={24}
                  color="#8b5cf6"
                />
                <Text className="text-purple-600 font-medium mt-2">
                  Settings
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
