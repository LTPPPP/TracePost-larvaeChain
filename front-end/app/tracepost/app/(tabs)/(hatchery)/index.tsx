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
import { getHatcheries } from "@/api/hatchery";
import "@/global.css";

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

export default function HatcheryScreen() {
  const [hatcheries, setHatcheries] = useState<Hatchery[]>([]);
  const [filteredHatcheries, setFilteredHatcheries] = useState<Hatchery[]>([]);
  const [searchQuery, setSearchQuery] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const router = useRouter();
  const { userData, getCompanyId } = useRole();

  const loadHatcheries = async () => {
    setIsLoading(true);
    try {
      const response = await getHatcheries();

      if (response.success) {
        setHatcheries(response.data);
        setFilteredHatcheries(response.data);
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
      setIsLoading(false);
    }
  };

  const refreshHatcheries = async () => {
    setIsRefreshing(true);
    try {
      const response = await getHatcheries();

      if (response.success) {
        setHatcheries(response.data);
        applyFilters(response.data, searchQuery);
      } else {
        throw new Error(response.message);
      }
    } catch (error) {
      console.error("Error refreshing hatcheries:", error);
      Alert.alert(
        "Error",
        error instanceof Error ? error.message : "Failed to refresh hatcheries",
      );
    } finally {
      setIsRefreshing(false);
    }
  };

  const applyFilters = (hatcheryList: Hatchery[], search: string) => {
    let filtered = hatcheryList;

    // Apply search filter
    if (search) {
      filtered = filtered.filter(
        (h) =>
          h.name.toLowerCase().includes(search.toLowerCase()) ||
          h.company.name.toLowerCase().includes(search.toLowerCase()) ||
          h.company.location.toLowerCase().includes(search.toLowerCase()),
      );
    }

    setFilteredHatcheries(filtered);
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    applyFilters(hatcheries, query);
  };

  const navigateToHatchery = (hatcheryId: number) => {
    router.push(`/(tabs)/(hatchery)/${hatcheryId}`);
  };

  const createNewHatchery = () => {
    router.push("/(tabs)/(hatchery)/create");
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
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
                  <TablerIconComponent name="check" size={20} color="#3b82f6" />
                  <Text className="text-blue-700 font-medium ml-2">Active</Text>
                </View>
                <Text className="text-2xl font-bold text-blue-800">
                  {hatcheries.filter((h) => h.is_active).length}
                </Text>
              </View>
            </View>
          </View>

          {/* Search Bar */}
          <View className="mb-6">
            <View className="flex-row bg-gray-100 rounded-xl p-3 items-center">
              <TablerIconComponent name="search" size={20} color="#9ca3af" />
              <TextInput
                className="flex-1 ml-3 text-gray-700"
                placeholder="Search hatcheries..."
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
                {searchQuery
                  ? "Try adjusting your search"
                  : "Create your first hatchery to get started"}
              </Text>
              {!searchQuery && (
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
                        name="building"
                        size={14}
                        color="#9ca3af"
                      />
                      <Text className="text-gray-500 text-sm ml-1">
                        {hatchery.company.name}
                      </Text>
                    </View>
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

                {/* Company Info */}
                <View className="bg-gray-50 p-3 rounded-lg mb-3">
                  <View className="flex-row justify-between items-center mb-1">
                    <Text className="text-gray-600 text-sm">Company Type</Text>
                    <Text className="font-medium text-sm">
                      {hatchery.company.type}
                    </Text>
                  </View>
                  <View className="flex-row justify-between items-center mb-1">
                    <Text className="text-gray-600 text-sm">Location</Text>
                    <Text className="font-medium text-sm">
                      {hatchery.company.location}
                    </Text>
                  </View>
                  <View className="flex-row justify-between items-center">
                    <Text className="text-gray-600 text-sm">Contact</Text>
                    <Text className="font-medium text-sm" numberOfLines={1}>
                      {hatchery.company.contact_info}
                    </Text>
                  </View>
                </View>

                {/* Footer */}
                <View className="flex-row items-center justify-between pt-3 border-t border-gray-100">
                  <View className="flex-row items-center">
                    <TablerIconComponent
                      name="calendar"
                      size={14}
                      color="#9ca3af"
                    />
                    <Text className="text-gray-500 text-xs ml-1">
                      Created {formatDate(hatchery.created_at)}
                    </Text>
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
                <TablerIconComponent name="refresh" size={24} color="#10b981" />
                <Text className="text-green-600 font-medium mt-2">Refresh</Text>
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
