import React, { useState, useEffect } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  Dimensions,
  ActivityIndicator,
  Modal,
  RefreshControl,
  Alert,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import TablerIconComponent from "@/components/icon";
import { LineChart } from "react-native-chart-kit";
import "@/global.css";

import { logout } from "@/api/auth";
import { useRouter } from "expo-router";
import { useRole } from "@/contexts/RoleContext";
import { getHatcheries } from "@/api/hatchery";
import { getAllBatches, getBatchesByHatchery, BatchData } from "@/api/batch";

const screenWidth = Dimensions.get("window").width;

interface HatcheryWithBatches {
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
  batches: BatchData[];
  stats: {
    totalBatches: number;
    activeBatches: number;
    completedBatches: number;
    totalLarvae: number;
    averageQuantity: number;
    successRate: number;
  };
}

interface DashboardStats {
  totalHatcheries: number;
  activeHatcheries: number;
  totalBatches: number;
  activeBatches: number;
  totalLarvae: number;
  averageSuccessRate: number;
  recentActivity: ActivityItem[];
}

interface ActivityItem {
  id: string;
  type: "batch_created" | "batch_completed" | "hatchery_created";
  title: string;
  description: string;
  timestamp: string;
  hatcheryName?: string;
  batchId?: number;
  icon: string;
  color: string;
}

export default function HomeScreen() {
  const [activeTab, setActiveTab] = useState("overview");
  const [selectedItem, setSelectedItem] = useState<HatcheryWithBatches | null>(
    null,
  );
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [walletConnected, setWalletConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [blockchainSynced, setBlockchainSynced] = useState(true);
  const [dataLastUpdated, setDataLastUpdated] = useState<string>("");

  // Real data state
  const [hatcheriesWithBatches, setHatcheriesWithBatches] = useState<
    HatcheryWithBatches[]
  >([]);
  const [dashboardStats, setDashboardStats] = useState<DashboardStats>({
    totalHatcheries: 0,
    activeHatcheries: 0,
    totalBatches: 0,
    activeBatches: 0,
    totalLarvae: 0,
    averageSuccessRate: 0,
    recentActivity: [],
  });
  // Add batches state for user dashboard
  const [batches, setBatches] = useState<BatchData[]>([]);

  const router = useRouter();
  const { currentRole, userData, isHatchery, isUser } = useRole();

  // Load real data
  const loadDashboardData = async () => {
    try {
      setIsLoading(true);

      // Load hatcheries
      const hatcheriesResponse = await getHatcheries();
      if (!hatcheriesResponse.success) {
        throw new Error(hatcheriesResponse.message);
      }

      // Load all batches
      const batchesResponse = await getAllBatches();
      let allBatches: BatchData[] = [];
      if (batchesResponse.success && batchesResponse.data) {
        allBatches = batchesResponse.data;
        setBatches(allBatches); // Set batches for user view
      }

      // Combine hatcheries with their batches and calculate stats
      const hatcheriesWithStats: HatcheryWithBatches[] =
        hatcheriesResponse.data.map((hatchery) => {
          const hatcheryBatches = allBatches.filter(
            (batch) => batch.hatchery_id === hatchery.id,
          );

          const totalBatches = hatcheryBatches.length;
          const activeBatches = hatcheryBatches.filter(
            (b) => b.is_active,
          ).length;
          const completedBatches = hatcheryBatches.filter(
            (b) => b.status.toLowerCase() === "completed",
          ).length;
          const totalLarvae = hatcheryBatches.reduce(
            (sum, b) => sum + b.quantity,
            0,
          );
          const averageQuantity =
            totalBatches > 0 ? Math.round(totalLarvae / totalBatches) : 0;
          const successRate =
            totalBatches > 0
              ? Math.round((completedBatches / totalBatches) * 100)
              : 0;

          return {
            ...hatchery,
            batches: hatcheryBatches,
            stats: {
              totalBatches,
              activeBatches,
              completedBatches,
              totalLarvae,
              averageQuantity,
              successRate,
            },
          };
        });

      setHatcheriesWithBatches(hatcheriesWithStats);

      // Calculate dashboard statistics
      const stats = calculateDashboardStats(hatcheriesWithStats, allBatches);
      setDashboardStats(stats);

      // Update last updated time
      setDataLastUpdated(new Date().toLocaleTimeString());
    } catch (error) {
      console.error("Error loading dashboard data:", error);
      Alert.alert("Error", "Failed to load dashboard data. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  // Calculate dashboard statistics
  const calculateDashboardStats = (
    hatcheries: HatcheryWithBatches[],
    batches: BatchData[],
  ): DashboardStats => {
    const totalHatcheries = hatcheries.length;
    const activeHatcheries = hatcheries.filter((h) => h.is_active).length;
    const totalBatches = batches.length;
    const activeBatches = batches.filter((b) => b.is_active).length;
    const totalLarvae = batches.reduce((sum: number, b) => sum + b.quantity, 0);

    // Calculate average success rate across all hatcheries
    const hatcheriesWithBatches = hatcheries.filter(
      (h) => h.stats.totalBatches > 0,
    );
    const averageSuccessRate =
      hatcheriesWithBatches.length > 0
        ? Math.round(
            hatcheriesWithBatches.reduce(
              (sum, h) => sum + h.stats.successRate,
              0,
            ) / hatcheriesWithBatches.length,
          )
        : 0;

    // Generate recent activity
    const recentActivity = generateRecentActivity(hatcheries, batches);

    return {
      totalHatcheries,
      activeHatcheries,
      totalBatches,
      activeBatches,
      totalLarvae,
      averageSuccessRate,
      recentActivity,
    };
  };

  // Generate recent activity from real data
  const generateRecentActivity = (
    hatcheries: HatcheryWithBatches[],
    batches: BatchData[],
  ): ActivityItem[] => {
    const activities: ActivityItem[] = [];

    // Add recent batches (last 10 days)
    const recentBatches = batches
      .filter((batch) => {
        const batchDate = new Date(batch.created_at);
        const tenDaysAgo = new Date();
        tenDaysAgo.setDate(tenDaysAgo.getDate() - 10);
        return batchDate > tenDaysAgo;
      })
      .sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
      )
      .slice(0, 5);

    recentBatches.forEach((batch) => {
      const hatchery = hatcheries.find((h) => h.id === batch.hatchery_id);
      if (batch.status.toLowerCase() === "completed") {
        activities.push({
          id: `batch_completed_${batch.id}`,
          type: "batch_completed",
          title: `Batch #${batch.id} completed`,
          description: `${batch.quantity.toLocaleString()} ${batch.species}`,
          timestamp: batch.updated_at,
          hatcheryName: hatchery?.name,
          batchId: batch.id,
          icon: "check-circle",
          color: "#10b981",
        });
      } else {
        activities.push({
          id: `batch_created_${batch.id}`,
          type: "batch_created",
          title: `New batch #${batch.id} created`,
          description: `${batch.quantity.toLocaleString()} ${batch.species}`,
          timestamp: batch.created_at,
          hatcheryName: hatchery?.name,
          batchId: batch.id,
          icon: "package",
          color: "#f97316",
        });
      }
    });

    // Add recent hatcheries (last 30 days)
    const recentHatcheries = hatcheries
      .filter((hatchery) => {
        const hatcheryDate = new Date(hatchery.created_at);
        const thirtyDaysAgo = new Date();
        thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
        return hatcheryDate > thirtyDaysAgo;
      })
      .sort(
        (a, b) =>
          new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
      )
      .slice(0, 2);

    recentHatcheries.forEach((hatchery) => {
      activities.push({
        id: `hatchery_created_${hatchery.id}`,
        type: "hatchery_created",
        title: `New hatchery "${hatchery.name}" established`,
        description: `${hatchery.company.location} • ${hatchery.company.type}`,
        timestamp: hatchery.created_at,
        hatcheryName: hatchery.name,
        icon: "building-factory-2",
        color: "#3b82f6",
      });
    });

    // Sort all activities by timestamp (most recent first)
    return activities
      .sort(
        (a, b) =>
          new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
      )
      .slice(0, 5);
  };

  // Generate chart data from real batches
  const getProductionChartData = () => {
    if (hatcheriesWithBatches.length === 0) {
      return {
        labels: ["No Data"],
        datasets: [{ data: [0] }],
        legend: ["Production Rate (%)"],
      };
    }

    // Get last 6 months of data
    const months = [];
    const productionData = [];

    for (let i = 5; i >= 0; i--) {
      const date = new Date();
      date.setMonth(date.getMonth() - i);
      const monthKey = date.toISOString().slice(0, 7); // YYYY-MM format
      const monthLabel = date.toLocaleDateString("en-US", { month: "short" });

      months.push(monthLabel);

      // Calculate production rate for this month (completed batches / total batches * 100)
      const monthBatches = hatcheriesWithBatches
        .flatMap((h) => h.batches)
        .filter((batch) => {
          return batch.created_at.startsWith(monthKey);
        });

      const completedBatches = monthBatches.filter(
        (b) => b.status.toLowerCase() === "completed",
      ).length;
      const productionRate =
        monthBatches.length > 0
          ? Math.round((completedBatches / monthBatches.length) * 100)
          : 0;

      productionData.push(productionRate);
    }

    return {
      labels: months,
      datasets: [
        {
          data: productionData.length > 0 ? productionData : [0],
          color: (opacity = 1) => `rgba(67, 56, 202, ${opacity})`,
          strokeWidth: 2,
        },
      ],
      legend: ["Production Rate (%)"],
    };
  };

  // Generate temperature trend data (simulated based on real batches count)
  const getTemperatureChartData = () => {
    const activeBatchesCount = dashboardStats.activeBatches;
    const baseTemp = 28.5;

    // Simulate temperature variations based on number of active batches
    const tempData = Array.from({ length: 6 }, (_, i) => {
      const variation =
        Math.sin(i) * 0.8 + (activeBatchesCount > 10 ? 0.5 : -0.5);
      return Number((baseTemp + variation).toFixed(1));
    });

    return {
      labels: ["6am", "9am", "12pm", "3pm", "6pm", "9pm"],
      datasets: [
        {
          data: tempData,
          color: (opacity = 1) => `rgba(249, 115, 22, ${opacity})`,
          strokeWidth: 2,
        },
      ],
      legend: ["Avg Temperature (°C)"],
    };
  };

  const handleLogout = async () => {
    try {
      await logout();
      router.replace("/(auth)/login");
    } catch (error) {
      console.error("Logout error:", error);
    }
  };

  const connectWallet = () => {
    setIsConnecting(true);
    setTimeout(() => {
      setIsConnecting(false);
      setWalletConnected(true);
    }, 2000);
  };

  const refreshData = async () => {
    setIsRefreshing(true);
    setBlockchainSynced(false);
    await loadDashboardData();
    setTimeout(() => {
      setBlockchainSynced(true);
      setIsRefreshing(false);
    }, 1000);
  };

  const viewItemDetails = (item: HatcheryWithBatches) => {
    setSelectedItem(item);
    setIsModalVisible(true);
  };

  const navigateToHatchery = (hatcheryId: number) => {
    router.push(`/(tabs)/(hatchery)/${hatcheryId}`);
  };

  const navigateToBatch = (batchId: number) => {
    router.push(`/(tabs)/(batches)/${batchId}`);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const chartConfig = {
    backgroundGradientFrom: "#fff",
    backgroundGradientTo: "#fff",
    decimalPlaces: 1,
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

  // Load data on component mount
  useEffect(() => {
    if (isHatchery || isUser) {
      loadDashboardData();
    }
  }, [isHatchery, isUser]);

  // For user role, show the existing user dashboard
  if (isUser) {
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
                  Farm Dashboard
                </Text>
                <Text className="text-gray-500">
                  Monitoring pond conditions
                </Text>
                {userData && (
                  <Text className="text-xs text-gray-400 mt-1">
                    {userData.username} • {currentRole}
                  </Text>
                )}
              </View>
              <View className="flex-row">
                <TouchableOpacity
                  className="h-10 w-10 rounded-full bg-primary/10 items-center justify-center mr-2"
                  onPress={() => {}}
                >
                  <TablerIconComponent name="bell" size={20} color="#3b82f6" />
                </TouchableOpacity>
                <TouchableOpacity
                  className="h-10 w-10 rounded-full bg-red-100 items-center justify-center"
                  onPress={handleLogout}
                >
                  <TablerIconComponent
                    name="logout"
                    size={20}
                    color="#ef4444"
                  />
                </TouchableOpacity>
              </View>
            </View>

            {/* Stats Overview */}
            <View className="flex-row flex-wrap mb-6">
              <View className="w-1/2 pr-2 mb-4">
                <View className="bg-blue-50 p-4 rounded-xl">
                  <View className="flex-row items-center mb-2">
                    <TablerIconComponent
                      name="fish"
                      size={20}
                      color="#3b82f6"
                    />
                    <Text className="text-blue-700 font-medium ml-2">
                      Current Batches
                    </Text>
                  </View>
                  <Text className="text-2xl font-bold text-blue-800">
                    {batches.length}
                  </Text>
                </View>
              </View>

              <View className="w-1/2 pl-2 mb-4">
                <View className="bg-green-50 p-4 rounded-xl">
                  <View className="flex-row items-center mb-2">
                    <TablerIconComponent
                      name="check"
                      size={20}
                      color="#10b981"
                    />
                    <Text className="text-green-700 font-medium ml-2">
                      Active
                    </Text>
                  </View>
                  <Text className="text-2xl font-bold text-green-800">
                    {batches.filter((b: BatchData) => b.is_active).length}
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
                      Total Larvae
                    </Text>
                  </View>
                  <Text className="text-2xl font-bold text-orange-800">
                    {batches
                      .reduce(
                        (sum: number, b: BatchData) => sum + b.quantity,
                        0,
                      )
                      .toLocaleString()}
                  </Text>
                </View>
              </View>

              <View className="w-1/2 pl-2">
                <View className="bg-indigo-50 p-4 rounded-xl">
                  <View className="flex-row items-center mb-2">
                    <TablerIconComponent
                      name="calendar"
                      size={20}
                      color="#4338ca"
                    />
                    <Text className="text-indigo-700 font-medium ml-2">
                      Latest Batch
                    </Text>
                  </View>
                  <Text className="text-2xl font-bold text-indigo-800">
                    {batches.length > 0
                      ? new Date(
                          Math.max(
                            ...batches.map((b: BatchData) =>
                              new Date(b.created_at).getTime(),
                            ),
                          ),
                        ).toLocaleDateString("en-US", {
                          month: "short",
                          day: "numeric",
                        })
                      : "None"}
                  </Text>
                </View>
              </View>
            </View>

            {/* Environment Overview */}
            <View className="bg-blue-300 p-5 rounded-xl mb-6">
              <View className="flex-row justify-between items-start mb-4">
                <View>
                  <Text className="text-white/80 text-sm">
                    Today&apos;s Conditions
                  </Text>
                  <Text className="text-white font-bold text-xl">Optimal</Text>
                </View>
                <View className="bg-white/20 p-2 rounded-lg">
                  <TablerIconComponent
                    name="temperature"
                    size={24}
                    color="white"
                  />
                </View>
              </View>

              <View className="flex-row flex-wrap">
                <View className="w-1/3 mb-3">
                  <Text className="text-white/70 text-xs">Temperature</Text>
                  <Text className="text-white">28.5°C</Text>
                </View>
                <View className="w-1/3 mb-3">
                  <Text className="text-white/70 text-xs">pH Level</Text>
                  <Text className="text-white">7.8</Text>
                </View>
                <View className="w-1/3 mb-3">
                  <Text className="text-white/70 text-xs">Salinity</Text>
                  <Text className="text-white">15 ppt</Text>
                </View>
              </View>

              <View className="flex-row items-center bg-white/20 p-3 rounded-lg mt-3">
                <TablerIconComponent
                  name="alert-circle"
                  size={18}
                  color="white"
                />
                <Text className="text-white ml-2">
                  All parameters within acceptable ranges
                </Text>
              </View>
            </View>

            {/* My Batches */}
            <Text className="text-lg font-semibold mb-4">My Batches</Text>

            {isLoading ? (
              <View className="items-center py-10">
                <ActivityIndicator size="large" color="#3b82f6" />
                <Text className="text-gray-500 mt-3">
                  Loading your batches...
                </Text>
              </View>
            ) : batches.length > 0 ? (
              batches.map((batch: BatchData) => (
                <TouchableOpacity
                  key={batch.id}
                  className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
                  onPress={() => router.push(`/(tabs)/(batches)/${batch.id}`)}
                >
                  <View className="flex-row justify-between items-start mb-3">
                    <View className="flex-1">
                      <Text className="font-bold text-lg text-gray-800 mb-1">
                        Batch #{batch.id}
                      </Text>
                      <Text className="text-gray-500 text-sm">
                        {batch.hatchery.name}
                      </Text>
                    </View>
                    <View
                      className={`px-3 py-1 rounded-full ${
                        batch.status.toLowerCase() === "active"
                          ? "bg-green-100 text-green-700"
                          : batch.status.toLowerCase() === "completed"
                            ? "bg-blue-100 text-blue-700"
                            : "bg-gray-100 text-gray-700"
                      }`}
                    >
                      <Text className="text-xs font-medium capitalize">
                        {batch.status}
                      </Text>
                    </View>
                  </View>

                  <View className="flex-row flex-wrap">
                    <View className="w-1/2 mb-2">
                      <Text className="text-gray-500 text-xs">Species</Text>
                      <Text className="font-medium" numberOfLines={1}>
                        {batch.species}
                      </Text>
                    </View>

                    <View className="w-1/2 mb-2">
                      <Text className="text-gray-500 text-xs">Quantity</Text>
                      <View className="flex-row items-center">
                        <TablerIconComponent
                          name="fish"
                          size={14}
                          color="#3b82f6"
                        />
                        <Text className="ml-1 font-medium">
                          {batch.quantity.toLocaleString()}
                        </Text>
                      </View>
                    </View>
                  </View>

                  <View className="flex-row items-center justify-between mt-3 pt-3 border-t border-gray-100">
                    <Text className="text-xs text-gray-500">
                      Created {new Date(batch.created_at).toLocaleDateString()}
                    </Text>
                    <View className="flex-row items-center">
                      <Text className="font-medium text-blue-600 mr-1 text-sm">
                        View Details
                      </Text>
                      <TablerIconComponent
                        name="chevron-right"
                        size={16}
                        color="#3b82f6"
                      />
                    </View>
                  </View>
                </TouchableOpacity>
              ))
            ) : (
              <View className="bg-gray-50 p-8 rounded-xl items-center mb-6">
                <TablerIconComponent
                  name="fish-off"
                  size={48}
                  color="#9ca3af"
                />
                <Text className="text-gray-500 font-medium mt-4 mb-2">
                  No batches found
                </Text>
                <Text className="text-gray-400 text-center mb-4">
                  You don&apos;t have any assigned batches yet
                </Text>
              </View>
            )}

            {/* Quick Actions */}
            <View className="mt-6">
              <Text className="text-lg font-semibold mb-4">Quick Actions</Text>
              <View className="flex-row flex-wrap gap-3">
                <TouchableOpacity
                  className="flex-1 bg-blue-50 p-4 rounded-xl items-center min-w-[45%]"
                  onPress={() => router.push("/(tabs)/(report)")}
                >
                  <TablerIconComponent
                    name="report"
                    size={24}
                    color="#3b82f6"
                  />
                  <Text className="text-blue-600 font-medium mt-2">
                    New Report
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  className="flex-1 bg-green-50 p-4 rounded-xl items-center min-w-[45%]"
                  onPress={() => router.push("/(tabs)/(track)")}
                >
                  <TablerIconComponent
                    name="qrcode"
                    size={24}
                    color="#10b981"
                  />
                  <Text className="text-green-600 font-medium mt-2">
                    Scan QR Code
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity
                  className="flex-1 bg-orange-50 p-4 rounded-xl items-center min-w-[45%]"
                  onPress={refreshData}
                >
                  <TablerIconComponent
                    name="refresh"
                    size={24}
                    color="#f97316"
                  />
                  <Text className="text-orange-600 font-medium mt-2">
                    Refresh
                  </Text>
                </TouchableOpacity>

                <TouchableOpacity className="flex-1 bg-indigo-50 p-4 rounded-xl items-center min-w-[45%]">
                  <TablerIconComponent
                    name="chart-dots"
                    size={24}
                    color="#4338ca"
                  />
                  <Text className="text-indigo-600 font-medium mt-2">
                    Analytics
                  </Text>
                </TouchableOpacity>
              </View>
            </View>
          </View>
        </ScrollView>
      </SafeAreaView>
    );
  }

  // Loading state
  if (isLoading) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <View className="flex-1 justify-center items-center">
          <ActivityIndicator size="large" color="#f97316" />
          <Text className="text-gray-500 mt-4">Loading dashboard...</Text>
        </View>
      </SafeAreaView>
    );
  }

  // Hatchery Dashboard
  return (
    <SafeAreaView className="flex-1 bg-white">
      <ScrollView
        contentContainerStyle={{ paddingBottom: 100 }}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={refreshData}
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
                Hatchery Dashboard
              </Text>
              <Text className="text-gray-500">
                Manage your breeding operations
              </Text>
              {userData && (
                <Text className="text-xs text-gray-400 mt-1">
                  {userData.username} • {currentRole}
                </Text>
              )}
            </View>
            <View className="flex-row">
              <TouchableOpacity
                className="h-10 w-10 rounded-full bg-primary/10 items-center justify-center mr-2"
                onPress={() => {}}
              >
                <TablerIconComponent name="bell" size={20} color="#f97316" />
              </TouchableOpacity>
              <TouchableOpacity
                className="h-10 w-10 rounded-full bg-red-100 items-center justify-center"
                onPress={handleLogout}
              >
                <TablerIconComponent name="logout" size={20} color="#ef4444" />
              </TouchableOpacity>
            </View>
          </View>

          {/* Wallet Connection Status */}
          {!walletConnected ? (
            <TouchableOpacity
              className="bg-indigo-100 p-4 rounded-xl mb-6 flex-row items-center"
              onPress={connectWallet}
              disabled={isConnecting}
            >
              <View className="h-10 w-10 rounded-full bg-indigo-200 items-center justify-center mr-3">
                <TablerIconComponent name="wallet" size={20} color="#4338ca" />
              </View>
              <View className="flex-1">
                <Text className="font-semibold text-indigo-700">
                  Connect Your Wallet
                </Text>
                <Text className="text-indigo-600 text-sm">
                  Connect your Web3 wallet to enable blockchain features
                </Text>
              </View>
              {isConnecting ? (
                <ActivityIndicator color="#4338ca" />
              ) : (
                <TablerIconComponent
                  name="chevron-right"
                  size={20}
                  color="#4338ca"
                />
              )}
            </TouchableOpacity>
          ) : (
            <View className="bg-green-50 p-4 rounded-xl mb-6 flex-row items-center">
              <View className="h-10 w-10 rounded-full bg-green-100 items-center justify-center mr-3">
                <TablerIconComponent name="wallet" size={20} color="#10b981" />
              </View>
              <View className="flex-1">
                <Text className="font-semibold text-green-700">
                  Wallet Connected
                </Text>
                <Text className="text-green-600 text-sm">
                  0x71C7...976F • Hatchery Owner
                </Text>
              </View>
              <TouchableOpacity
                className="bg-green-200 px-3 py-1 rounded-lg"
                onPress={refreshData}
              >
                <Text className="text-green-700 text-sm font-medium">
                  {blockchainSynced ? "Synced" : "Syncing..."}
                </Text>
              </TouchableOpacity>
            </View>
          )}

          {/* Weather & Date Card */}
          <View className="bg-green-300 p-4 rounded-2xl mb-6">
            <View className="flex-row justify-between items-center">
              <View>
                <Text className="text-white text-lg font-medium">
                  Good{" "}
                  {new Date().getHours() < 12
                    ? "morning"
                    : new Date().getHours() < 18
                      ? "afternoon"
                      : "evening"}
                  !
                </Text>
                <Text className="text-white/80 text-sm">
                  {new Date().toLocaleDateString("en-US", {
                    weekday: "long",
                    year: "numeric",
                    month: "long",
                    day: "numeric",
                  })}
                </Text>
                <View className="flex-row items-center mt-2">
                  <TablerIconComponent
                    name="temperature"
                    size={18}
                    color="white"
                  />
                  <Text className="text-white ml-1">32°C</Text>
                  <View className="h-4 w-0.5 bg-white/30 mx-2" />
                  <TablerIconComponent name="droplet" size={18} color="white" />
                  <Text className="text-white ml-1">65%</Text>
                </View>
              </View>
              <View className="items-center">
                <TablerIconComponent name="sun" size={36} color="white" />
                <Text className="text-white mt-1">Sunny</Text>
              </View>
            </View>
          </View>

          {/* Data Last Updated */}
          <View className="flex-row items-center justify-between mb-6">
            <Text className="text-gray-500 text-sm">
              Data last updated: {dataLastUpdated || "Loading..."}
            </Text>
            <TouchableOpacity
              className="flex-row items-center"
              onPress={refreshData}
              disabled={isRefreshing}
            >
              <TablerIconComponent
                name="refresh"
                size={16}
                color="#4b5563"
                style={{ marginRight: 4 }}
              />
              <Text className="text-gray-600 text-sm">
                {isRefreshing ? "Refreshing..." : "Refresh"}
              </Text>
            </TouchableOpacity>
          </View>

          {/* Statistics Cards */}
          <View className="flex-row flex-wrap mb-6">
            <View className="w-1/2 pr-2 mb-4">
              <View className="bg-blue-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="building-factory-2"
                    size={20}
                    color="#3b82f6"
                  />
                  <Text className="text-blue-700 font-medium ml-2">
                    Hatcheries
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-blue-800">
                  {dashboardStats.totalHatcheries}
                </Text>
                <Text className="text-blue-600 text-xs">
                  {dashboardStats.activeHatcheries} active
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2 mb-4">
              <View className="bg-green-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="package"
                    size={20}
                    color="#10b981"
                  />
                  <Text className="text-green-700 font-medium ml-2">
                    Batches
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-green-800">
                  {dashboardStats.totalBatches}
                </Text>
                <Text className="text-green-600 text-xs">
                  {dashboardStats.activeBatches} active
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
                  {dashboardStats.totalLarvae.toLocaleString()}
                </Text>
              </View>
            </View>

            <View className="w-1/2 pl-2">
              <View className="bg-indigo-50 p-4 rounded-xl">
                <View className="flex-row items-center mb-2">
                  <TablerIconComponent
                    name="percentage"
                    size={20}
                    color="#4338ca"
                  />
                  <Text className="text-indigo-700 font-medium ml-2">
                    Success Rate
                  </Text>
                </View>
                <Text className="text-2xl font-bold text-indigo-800">
                  {dashboardStats.averageSuccessRate}%
                </Text>
              </View>
            </View>
          </View>

          {/* Tabs */}
          <View className="mb-4">
            <Text className="text-sm font-medium text-gray-700 mb-2">View</Text>
            <View className="flex-row">
              {["overview", "analytics"].map((tab) => (
                <TouchableOpacity
                  key={tab}
                  className={`px-6 py-3 rounded-full mr-3 ${
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

          {activeTab === "overview" ? (
            <>
              {/* Your Hatcheries */}
              <Text className="text-lg font-semibold mb-4">
                Your Hatcheries
              </Text>
              {hatcheriesWithBatches.length > 0 ? (
                hatcheriesWithBatches.map((hatchery) => (
                  <TouchableOpacity
                    key={hatchery.id}
                    className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
                    onPress={() => navigateToHatchery(hatchery.id)}
                  >
                    <View className="flex-row justify-between items-center mb-3">
                      <Text className="font-bold text-lg">{hatchery.name}</Text>
                      <View
                        className={`px-3 py-1 rounded-full ${
                          hatchery.is_active ? "bg-green-100" : "bg-yellow-100"
                        }`}
                      >
                        <Text
                          className={
                            hatchery.is_active
                              ? "text-green-600"
                              : "text-yellow-600"
                          }
                        >
                          {hatchery.is_active ? "Active" : "Inactive"}
                        </Text>
                      </View>
                    </View>

                    <View className="flex-row flex-wrap">
                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-500">Active Batches</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="package"
                            size={16}
                            color="#f97316"
                          />
                          <Text className="ml-1 font-medium">
                            {hatchery.stats.activeBatches}
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-500">Total Batches</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="clipboard-list"
                            size={16}
                            color="#4338ca"
                          />
                          <Text className="ml-1 font-medium">
                            {hatchery.stats.totalBatches}
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-1">
                        <Text className="text-gray-500">Total Larvae</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="fish"
                            size={16}
                            color="#10b981"
                          />
                          <Text className="ml-1 font-medium">
                            {hatchery.stats.totalLarvae.toLocaleString()}
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-1">
                        <Text className="text-gray-500">Success Rate</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="percentage"
                            size={16}
                            color="#ef4444"
                          />
                          <Text className="ml-1 font-medium">
                            {hatchery.stats.successRate}%
                          </Text>
                        </View>
                      </View>
                    </View>

                    <View className="flex-row items-center justify-between mt-3 pt-3 border-t border-gray-100">
                      <Text className="text-xs text-gray-500">
                        {hatchery.company.location} • {hatchery.company.type}
                      </Text>
                      <TouchableOpacity
                        className="flex-row items-center"
                        onPress={() => viewItemDetails(hatchery)}
                      >
                        <Text className="font-medium text-primary mr-1">
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
              ) : (
                <View className="bg-gray-50 p-8 rounded-xl items-center mb-6">
                  <TablerIconComponent
                    name="building-factory-2"
                    size={48}
                    color="#9ca3af"
                  />
                  <Text className="text-gray-500 font-medium mt-4 mb-2">
                    No hatcheries yet
                  </Text>
                  <Text className="text-gray-400 text-center mb-4">
                    Create your first hatchery to start managing batches
                  </Text>
                  <TouchableOpacity
                    className="bg-primary px-6 py-3 rounded-xl"
                    onPress={() => router.push("/(tabs)/(hatchery)/create")}
                  >
                    <Text className="text-white font-bold">
                      Create Hatchery
                    </Text>
                  </TouchableOpacity>
                </View>
              )}

              {/* Latest Activity */}
              <Text className="text-lg font-semibold mb-4 mt-6">
                Latest Activity
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm">
                {dashboardStats.recentActivity.length > 0 ? (
                  dashboardStats.recentActivity.map((activity) => (
                    <View key={activity.id}>
                      <TouchableOpacity
                        className="flex-row items-center mb-4"
                        onPress={() => {
                          if (activity.batchId) {
                            navigateToBatch(activity.batchId);
                          }
                        }}
                      >
                        <View
                          style={{ backgroundColor: activity.color }}
                          className="h-10 w-10 rounded-full items-center justify-center mr-3"
                        >
                          <TablerIconComponent
                            name={activity.icon}
                            size={20}
                            color="white"
                          />
                        </View>
                        <View className="flex-1">
                          <Text className="font-medium">{activity.title}</Text>
                          <Text className="text-gray-500 text-xs">
                            {activity.hatcheryName &&
                              `${activity.hatcheryName} • `}
                            {formatDate(activity.timestamp)}
                          </Text>
                        </View>
                        <TouchableOpacity className="bg-gray-100 px-2 py-1 rounded">
                          <Text className="text-xs text-gray-600">View</Text>
                        </TouchableOpacity>
                      </TouchableOpacity>
                      <View className="h-px bg-gray-100 mb-4" />
                    </View>
                  ))
                ) : (
                  <View className="items-center py-8">
                    <TablerIconComponent
                      name="clock"
                      size={48}
                      color="#9ca3af"
                    />
                    <Text className="text-gray-500 mt-2">
                      No recent activity
                    </Text>
                    <Text className="text-gray-400 text-sm">
                      Create some batches to see activity here
                    </Text>
                  </View>
                )}
              </View>

              {/* Blockchain Transparency Section */}
              {walletConnected && (
                <>
                  <Text className="text-lg font-semibold mb-4 mt-6">
                    Blockchain Transparency
                  </Text>

                  <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-4">
                    <View className="flex-row items-center mb-4">
                      <View className="h-10 w-10 rounded-full bg-indigo-100 items-center justify-center mr-3">
                        <TablerIconComponent
                          name="currency-ethereum"
                          size={20}
                          color="#4338ca"
                        />
                      </View>
                      <View className="flex-1">
                        <Text className="font-medium">
                          {dashboardStats.totalBatches} Batches Registered
                        </Text>
                        <Text className="text-gray-500 text-xs">
                          All batches recorded on blockchain
                        </Text>
                      </View>
                    </View>

                    <View className="flex-row items-center">
                      <View className="h-10 w-10 rounded-full bg-indigo-100 items-center justify-center mr-3">
                        <TablerIconComponent
                          name="file-certificate"
                          size={20}
                          color="#4338ca"
                        />
                      </View>
                      <View className="flex-1">
                        <Text className="font-medium">
                          Breeding Certificates Issued
                        </Text>
                        <Text className="text-gray-500 text-xs">
                          Verified by smart contracts
                        </Text>
                      </View>
                    </View>
                  </View>
                </>
              )}
            </>
          ) : (
            <>
              {/* Analytics Tab */}
              <Text className="text-lg font-semibold mb-4">
                Production Rate Trend
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-2 mb-6 shadow-sm">
                <LineChart
                  data={getProductionChartData()}
                  width={screenWidth - 40}
                  height={220}
                  chartConfig={{
                    ...chartConfig,
                    color: (opacity = 1) => `rgba(67, 56, 202, ${opacity})`,
                  }}
                  bezier
                  style={{
                    marginVertical: 8,
                    borderRadius: 16,
                  }}
                />
              </View>

              <Text className="text-lg font-semibold mb-4">
                Average Temperature
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-2 mb-6 shadow-sm">
                <LineChart
                  data={getTemperatureChartData()}
                  width={screenWidth - 40}
                  height={220}
                  chartConfig={chartConfig}
                  bezier
                  style={{
                    marginVertical: 8,
                    borderRadius: 16,
                  }}
                />
              </View>

              <Text className="text-lg font-semibold mb-4">
                Performance Summary
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-6">
                <View className="flex-row justify-between mb-2">
                  <Text className="text-gray-500">Total Production</Text>
                  <Text className="font-medium">
                    {dashboardStats.totalLarvae.toLocaleString()}
                  </Text>
                </View>
                <View className="flex-row justify-between mb-2">
                  <Text className="text-gray-500">Average Success Rate</Text>
                  <Text className="font-medium">
                    {dashboardStats.averageSuccessRate}%
                  </Text>
                </View>
                <View className="flex-row justify-between mb-2">
                  <Text className="text-gray-500">Active Operations</Text>
                  <Text className="font-medium">
                    {dashboardStats.activeBatches} batches
                  </Text>
                </View>
                <View className="flex-row justify-between">
                  <Text className="text-gray-500">Operational Facilities</Text>
                  <Text className="font-medium">
                    {dashboardStats.activeHatcheries} hatcheries
                  </Text>
                </View>
              </View>

              {/* Blockchain Data Verification */}
              <Text className="text-lg font-semibold mb-4">
                Blockchain Data Verification
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm">
                <Text className="text-gray-700 mb-3">
                  All breeding and batch data is recorded on the blockchain for
                  complete traceability and authenticity.
                </Text>

                <View className="bg-indigo-50 p-3 rounded-lg mb-4">
                  <View className="flex-row items-center">
                    <TablerIconComponent
                      name="shield-check"
                      size={16}
                      color="#4338ca"
                      style={{ marginRight: 6 }}
                    />
                    <Text className="text-indigo-700 text-sm font-medium">
                      Data Integrity
                    </Text>
                  </View>
                  <Text className="text-indigo-600 text-xs ml-6 mt-1">
                    All breeding records are cryptographically signed
                  </Text>
                </View>

                <View className="bg-indigo-50 p-3 rounded-lg mb-4">
                  <View className="flex-row items-center">
                    <TablerIconComponent
                      name="clock"
                      size={16}
                      color="#4338ca"
                      style={{ marginRight: 6 }}
                    />
                    <Text className="text-indigo-700 text-sm font-medium">
                      Immutable History
                    </Text>
                  </View>
                  <Text className="text-indigo-600 text-xs ml-6 mt-1">
                    Complete historical record of all batches
                  </Text>
                </View>

                <View className="bg-indigo-50 p-3 rounded-lg">
                  <View className="flex-row items-center">
                    <TablerIconComponent
                      name="eye"
                      size={16}
                      color="#4338ca"
                      style={{ marginRight: 6 }}
                    />
                    <Text className="text-indigo-700 text-sm font-medium">
                      Transparent Access
                    </Text>
                  </View>
                  <Text className="text-indigo-600 text-xs ml-6 mt-1">
                    Anyone can verify data authenticity
                  </Text>
                </View>
              </View>
            </>
          )}

          {/* Quick Actions */}
          <View className="mt-6">
            <Text className="text-lg font-semibold mb-4">Quick Actions</Text>
            <View className="flex-row flex-wrap gap-3">
              <TouchableOpacity
                className="flex-1 bg-primary/10 p-4 rounded-xl items-center min-w-[45%]"
                onPress={() => router.push("/(tabs)/(hatchery)/create")}
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
                className="flex-1 bg-green-50 p-4 rounded-xl items-center min-w-[45%]"
                onPress={() => router.push("/(tabs)/(batches)/create")}
              >
                <TablerIconComponent name="package" size={24} color="#10b981" />
                <Text className="text-green-600 font-medium mt-2">
                  New Batch
                </Text>
              </TouchableOpacity>

              <TouchableOpacity
                className="flex-1 bg-blue-50 p-4 rounded-xl items-center min-w-[45%]"
                onPress={() => router.push("/(tabs)/(batches)")}
              >
                <TablerIconComponent
                  name="clipboard-list"
                  size={24}
                  color="#3b82f6"
                />
                <Text className="text-blue-600 font-medium mt-2">
                  View Batches
                </Text>
              </TouchableOpacity>

              <TouchableOpacity
                className="flex-1 bg-indigo-50 p-4 rounded-xl items-center min-w-[45%]"
                onPress={() => router.push("/(tabs)/(hatchery)")}
              >
                <TablerIconComponent
                  name="chart-bar"
                  size={24}
                  color="#4338ca"
                />
                <Text className="text-indigo-600 font-medium mt-2">
                  Analytics
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>

      {/* Hatchery Detail Modal */}
      <Modal
        animationType="slide"
        transparent={true}
        visible={isModalVisible}
        onRequestClose={() => setIsModalVisible(false)}
      >
        <View className="flex-1 bg-black/50 justify-end">
          <View className="bg-white rounded-t-3xl p-5 h-[70%]">
            <View className="flex-row justify-between items-center mb-6">
              <Text className="text-xl font-bold">
                {selectedItem?.name} Overview
              </Text>
              <TouchableOpacity
                className="p-2"
                onPress={() => setIsModalVisible(false)}
              >
                <TablerIconComponent name="x" size={24} color="#000" />
              </TouchableOpacity>
            </View>

            {selectedItem && (
              <ScrollView showsVerticalScrollIndicator={false}>
                <View className="bg-gradient-to-r from-primary to-primary-dark p-4 rounded-xl mb-5">
                  <Text className="text-white/80 text-sm">Hatchery</Text>
                  <Text className="text-white font-bold text-xl mb-2">
                    {selectedItem.name}
                  </Text>

                  <View className="flex-row flex-wrap">
                    <View className="w-1/2 mb-2">
                      <Text className="text-white/70 text-xs">Status</Text>
                      <Text className="text-white">
                        {selectedItem.is_active ? "Active" : "Inactive"}
                      </Text>
                    </View>
                    <View className="w-1/2 mb-2">
                      <Text className="text-white/70 text-xs">Location</Text>
                      <Text className="text-white">
                        {selectedItem.company.location}
                      </Text>
                    </View>
                  </View>
                </View>

                <Text className="text-lg font-semibold mb-4">
                  Hatchery Statistics
                </Text>

                <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-5">
                  <View className="flex-row justify-between mb-3">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="package"
                        size={20}
                        color="#f97316"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">Active Batches</Text>
                    </View>
                    <Text className="text-primary font-bold">
                      {selectedItem.stats.activeBatches}
                    </Text>
                  </View>

                  <View className="flex-row justify-between mb-3">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="clipboard-list"
                        size={20}
                        color="#4338ca"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">Total Batches</Text>
                    </View>
                    <Text className="text-indigo-600 font-bold">
                      {selectedItem.stats.totalBatches}
                    </Text>
                  </View>

                  <View className="flex-row justify-between mb-3">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="fish"
                        size={20}
                        color="#10b981"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">Total Larvae</Text>
                    </View>
                    <Text className="text-green-600 font-bold">
                      {selectedItem.stats.totalLarvae.toLocaleString()}
                    </Text>
                  </View>

                  <View className="flex-row justify-between">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="percentage"
                        size={20}
                        color="#ef4444"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">Success Rate</Text>
                    </View>
                    <Text className="text-red-600 font-bold">
                      {selectedItem.stats.successRate}%
                    </Text>
                  </View>
                </View>

                {/* Recent Batches */}
                <Text className="text-lg font-semibold mb-4">
                  Recent Batches
                </Text>
                <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-5">
                  {selectedItem.batches.slice(0, 3).map((batch, index) => (
                    <TouchableOpacity
                      key={batch.id}
                      className="flex-row items-center mb-3 last:mb-0"
                      onPress={() => {
                        setIsModalVisible(false);
                        navigateToBatch(batch.id);
                      }}
                    >
                      <View className="h-8 w-8 rounded-full bg-primary/10 items-center justify-center mr-3">
                        <TablerIconComponent
                          name="package"
                          size={16}
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
                      <View
                        className={`px-2 py-1 rounded ${
                          batch.status === "active"
                            ? "bg-green-100"
                            : "bg-blue-100"
                        }`}
                      >
                        <Text
                          className={`text-xs ${
                            batch.status === "active"
                              ? "text-green-600"
                              : "text-blue-600"
                          }`}
                        >
                          {batch.status}
                        </Text>
                      </View>
                    </TouchableOpacity>
                  ))}
                </View>

                {/* Blockchain verification section */}
                <Text className="text-lg font-semibold mb-4">
                  Blockchain Verification
                </Text>

                <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-5">
                  <View className="flex-row items-center mb-3">
                    <View className="h-10 w-10 rounded-full bg-green-100 items-center justify-center mr-3">
                      <TablerIconComponent
                        name="shield-check"
                        size={20}
                        color="#10b981"
                      />
                    </View>
                    <View>
                      <Text className="font-medium">Data Verification</Text>
                      <Text className="text-gray-500 text-xs">
                        All batch data is verified on blockchain
                      </Text>
                    </View>
                    <View className="ml-auto">
                      <Text className="text-green-600 font-medium">
                        Verified
                      </Text>
                    </View>
                  </View>

                  <View className="flex-row items-center">
                    <View className="h-10 w-10 rounded-full bg-indigo-100 items-center justify-center mr-3">
                      <TablerIconComponent
                        name="currency-ethereum"
                        size={20}
                        color="#4338ca"
                      />
                    </View>
                    <View>
                      <Text className="font-medium">Smart Contract</Text>
                      <Text className="text-gray-500 text-xs">
                        0x3a4e...a581
                      </Text>
                    </View>
                    <TouchableOpacity className="ml-auto bg-indigo-100 px-3 py-1 rounded">
                      <Text className="text-indigo-600 text-xs">View</Text>
                    </TouchableOpacity>
                  </View>
                </View>

                <View className="flex-row gap-3 mb-10">
                  <TouchableOpacity
                    className="flex-1 bg-primary py-3 rounded-xl items-center"
                    onPress={() => {
                      setIsModalVisible(false);
                      navigateToHatchery(selectedItem.id);
                    }}
                  >
                    <Text className="text-white font-bold">
                      MANAGE HATCHERY
                    </Text>
                  </TouchableOpacity>

                  <TouchableOpacity
                    className="flex-1 bg-secondary py-3 rounded-xl items-center"
                    onPress={() => {
                      setIsModalVisible(false);
                      router.push("/(tabs)/(batches)/create");
                    }}
                  >
                    <Text className="text-white font-bold">NEW BATCH</Text>
                  </TouchableOpacity>
                </View>
              </ScrollView>
            )}
          </View>
        </View>
      </Modal>
    </SafeAreaView>
  );
}
