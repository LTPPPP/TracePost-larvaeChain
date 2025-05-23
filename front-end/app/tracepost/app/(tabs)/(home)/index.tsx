import React, { useState, useEffect } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  Image,
  Dimensions,
  ActivityIndicator,
  Modal,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import TablerIconComponent from "@/components/icon";
import { LineChart } from "react-native-chart-kit";
import "@/global.css";

const screenWidth = Dimensions.get("window").width;

export default function HomeScreen() {
  const [activeTab, setActiveTab] = useState("overview");
  const [selectedPond, setSelectedPond] = useState(null);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [walletConnected, setWalletConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [blockchainSynced, setBlockchainSynced] = useState(true);
  const [dataLastUpdated, setDataLastUpdated] = useState("10 minutes ago");

  // Sample data for pond metrics
  const pondData = [
    {
      id: 1,
      name: "Pond A1",
      temperature: "28.5°C",
      oxygen: "6.7 mg/L",
      ph: "7.2",
      ammonia: "0.05 ppm",
      status: "Normal",
      lastUpdated: "5 min ago",
      deviceId: "SENSOR-0042A",
      blockchainVerified: true,
      batchId: "SH-2023-10-A1",
    },
    {
      id: 2,
      name: "Pond B2",
      temperature: "29.1°C",
      oxygen: "5.9 mg/L",
      ph: "7.4",
      ammonia: "0.08 ppm",
      status: "Warning",
      lastUpdated: "2 min ago",
      deviceId: "SENSOR-0058B",
      blockchainVerified: true,
      batchId: "SH-2023-10-B2",
    },
    {
      id: 3,
      name: "Pond C3",
      temperature: "27.8°C",
      oxygen: "6.3 mg/L",
      ph: "7.0",
      ammonia: "0.03 ppm",
      status: "Normal",
      lastUpdated: "8 min ago",
      deviceId: "SENSOR-0063C",
      blockchainVerified: true,
      batchId: "SH-2023-10-C3",
    },
  ];

  // Chart data
  const tempData = {
    labels: ["6am", "9am", "12pm", "3pm", "6pm", "9pm"],
    datasets: [
      {
        data: [27.2, 27.8, 28.5, 29.2, 28.7, 28.1],
        color: (opacity = 1) => `rgba(249, 115, 22, ${opacity})`, // orange
        strokeWidth: 2,
      },
    ],
    legend: ["Temperature (°C)"],
  };

  const oxygenData = {
    labels: ["6am", "9am", "12pm", "3pm", "6pm", "9pm"],
    datasets: [
      {
        data: [5.8, 6.2, 6.7, 6.5, 6.1, 5.9],
        color: (opacity = 1) => `rgba(67, 56, 202, ${opacity})`, // indigo
        strokeWidth: 2,
      },
    ],
    legend: ["Oxygen (mg/L)"],
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

  // Connect wallet function
  const connectWallet = () => {
    setIsConnecting(true);

    // Simulate connecting to wallet
    setTimeout(() => {
      setIsConnecting(false);
      setWalletConnected(true);
    }, 2000);
  };

  // View pond details
  const viewPondDetails = (pond) => {
    setSelectedPond(pond);
    setIsModalVisible(true);
  };

  // Refresh data function
  const refreshData = () => {
    setBlockchainSynced(false);

    // Simulate data refresh
    setTimeout(() => {
      setBlockchainSynced(true);
      setDataLastUpdated("Just now");
    }, 2000);
  };

  return (
    <SafeAreaView className="flex-1 bg-white">
      <ScrollView
        contentContainerStyle={{ paddingBottom: 100 }}
        showsVerticalScrollIndicator={false}
      >
        <View className="px-5 pt-4 pb-6">
          {/* Header */}
          <View className="flex-row items-center justify-between mb-6">
            <View>
              <Text className="text-2xl font-bold text-gray-800">
                Dashboard
              </Text>
              <Text className="text-gray-500">Monitoring pond conditions</Text>
            </View>
            <TouchableOpacity className="h-10 w-10 rounded-full bg-primary/10 items-center justify-center">
              <TablerIconComponent name="bell" size={20} color="#f97316" />
            </TouchableOpacity>
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
                  0x71C7...976F • Farm Owner
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
                  Good morning!
                </Text>
                <Text className="text-white/80 text-sm">
                  Monday, Oct 21, 2023
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
              Data last updated: {dataLastUpdated}
            </Text>
            <TouchableOpacity
              className="flex-row items-center"
              onPress={refreshData}
            >
              <TablerIconComponent
                name="refresh"
                size={16}
                color="#4b5563"
                style={{ marginRight: 4 }}
              />
              <Text className="text-gray-600 text-sm">Refresh</Text>
            </TouchableOpacity>
          </View>

          {/* Tabs */}
          <View className="flex-row bg-gray-100 rounded-xl p-1 mb-6">
            <TouchableOpacity
              className={`flex-1 py-2 rounded-lg ${activeTab === "overview" ? "bg-white shadow" : ""}`}
              onPress={() => setActiveTab("overview")}
            >
              <Text
                className={`text-center font-medium ${activeTab === "overview" ? "text-primary" : "text-gray-500"}`}
              >
                Overview
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              className={`flex-1 py-2 rounded-lg ${activeTab === "analytics" ? "bg-white shadow" : ""}`}
              onPress={() => setActiveTab("analytics")}
            >
              <Text
                className={`text-center font-medium ${activeTab === "analytics" ? "text-primary" : "text-gray-500"}`}
              >
                Analytics
              </Text>
            </TouchableOpacity>
          </View>

          {activeTab === "overview" ? (
            <>
              {/* Ponds Overview */}
              <Text className="text-lg font-semibold mb-4">Active Ponds</Text>

              {pondData.map((pond) => (
                <View
                  key={pond.id}
                  className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
                >
                  <View className="flex-row justify-between items-center mb-3">
                    <Text className="font-bold text-lg">{pond.name}</Text>
                    <View
                      className={`px-3 py-1 rounded-full ${pond.status === "Normal" ? "bg-green-100" : "bg-yellow-100"}`}
                    >
                      <Text
                        className={
                          pond.status === "Normal"
                            ? "text-green-600"
                            : "text-yellow-600"
                        }
                      >
                        {pond.status}
                      </Text>
                    </View>
                  </View>

                  <View className="flex-row flex-wrap">
                    <View className="w-1/2 mb-3">
                      <Text className="text-gray-500">Temperature</Text>
                      <View className="flex-row items-center">
                        <TablerIconComponent
                          name="temperature"
                          size={16}
                          color="#f97316"
                        />
                        <Text className="ml-1 font-medium">
                          {pond.temperature}
                        </Text>
                      </View>
                    </View>

                    <View className="w-1/2 mb-3">
                      <Text className="text-gray-500">Oxygen</Text>
                      <View className="flex-row items-center">
                        <TablerIconComponent
                          name="droplet"
                          size={16}
                          color="#4338ca"
                        />
                        <Text className="ml-1 font-medium">{pond.oxygen}</Text>
                      </View>
                    </View>

                    <View className="w-1/2 mb-1">
                      <Text className="text-gray-500">pH Level</Text>
                      <View className="flex-row items-center">
                        <TablerIconComponent
                          name="chart-bar"
                          size={16}
                          color="#10b981"
                        />
                        <Text className="ml-1 font-medium">{pond.ph}</Text>
                      </View>
                    </View>

                    <View className="w-1/2 mb-1">
                      <Text className="text-gray-500">Ammonia</Text>
                      <View className="flex-row items-center">
                        <TablerIconComponent
                          name="alert-triangle"
                          size={16}
                          color="#ef4444"
                        />
                        <Text className="ml-1 font-medium">{pond.ammonia}</Text>
                      </View>
                    </View>
                  </View>

                  <View className="flex-row items-center justify-between mt-3 pt-3 border-t border-gray-100">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name={
                          pond.blockchainVerified ? "shield-check" : "shield"
                        }
                        size={16}
                        color={pond.blockchainVerified ? "#10b981" : "#9ca3af"}
                      />
                      <Text
                        className={`text-xs ml-1 ${pond.blockchainVerified ? "text-green-600" : "text-gray-500"}`}
                      >
                        {pond.blockchainVerified
                          ? "Blockchain Verified"
                          : "Not Verified"}
                      </Text>
                      <View className="h-3 w-0.5 bg-gray-200 mx-2" />
                      <Text className="text-xs text-gray-500">
                        Updated {pond.lastUpdated}
                      </Text>
                    </View>
                    <TouchableOpacity
                      className="flex-row items-center"
                      onPress={() => viewPondDetails(pond)}
                    >
                      <Text className="font-medium text-primary mr-1">
                        Details
                      </Text>
                      <TablerIconComponent
                        name="chevron-right"
                        size={16}
                        color="#f97316"
                      />
                    </TouchableOpacity>
                  </View>
                </View>
              ))}

              {/* Latest Activity */}
              <Text className="text-lg font-semibold mb-4 mt-2">
                Latest Activity
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm">
                <View className="flex-row items-center mb-4">
                  <View className="h-10 w-10 rounded-full bg-orange-100 items-center justify-center mr-3">
                    <TablerIconComponent
                      name="alert-circle"
                      size={20}
                      color="#f97316"
                    />
                  </View>
                  <View className="flex-1">
                    <Text className="font-medium">
                      Oxygen level warning in Pond B2
                    </Text>
                    <Text className="text-gray-500 text-xs">2 hours ago</Text>
                  </View>
                  <TouchableOpacity className="bg-gray-100 px-2 py-1 rounded">
                    <Text className="text-xs text-gray-600">View</Text>
                  </TouchableOpacity>
                </View>

                <View className="flex-row items-center mb-4">
                  <View className="h-10 w-10 rounded-full bg-green-100 items-center justify-center mr-3">
                    <TablerIconComponent
                      name="check"
                      size={20}
                      color="#10b981"
                    />
                  </View>
                  <View className="flex-1">
                    <Text className="font-medium">
                      Feeding completed for Pond A1
                    </Text>
                    <Text className="text-gray-500 text-xs">4 hours ago</Text>
                  </View>
                  <TouchableOpacity className="bg-gray-100 px-2 py-1 rounded">
                    <Text className="text-xs text-gray-600">View</Text>
                  </TouchableOpacity>
                </View>

                <View className="flex-row items-center">
                  <View className="h-10 w-10 rounded-full bg-blue-100 items-center justify-center mr-3">
                    <TablerIconComponent
                      name="refresh"
                      size={20}
                      color="#3b82f6"
                    />
                  </View>
                  <View className="flex-1">
                    <Text className="font-medium">
                      Water exchange in Pond C3
                    </Text>
                    <Text className="text-gray-500 text-xs">
                      Yesterday, 4:30 PM
                    </Text>
                  </View>
                  <TouchableOpacity className="bg-gray-100 px-2 py-1 rounded">
                    <Text className="text-xs text-gray-600">View</Text>
                  </TouchableOpacity>
                </View>
              </View>

              {/* Blockchain Transparency Section */}
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
                      SH-2023-10-B2 Batch Certified
                    </Text>
                    <Text className="text-gray-500 text-xs flex-row items-center">
                      <Text>Oct 18, 2023 • </Text>
                      <Text className="text-indigo-500">0x71C7...976F</Text>
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
                      Quality Certificate Issued
                    </Text>
                    <Text className="text-gray-500 text-xs flex-row items-center">
                      <Text>Oct 17, 2023 • </Text>
                      <Text className="text-indigo-500">0x3a4e...a581</Text>
                    </Text>
                  </View>
                </View>
              </View>
            </>
          ) : (
            <>
              {/* Analytics Tab */}
              <Text className="text-lg font-semibold mb-4">
                Temperature Trend
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-2 mb-6 shadow-sm">
                <LineChart
                  data={tempData}
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

              <Text className="text-lg font-semibold mb-4">Oxygen Levels</Text>
              <View className="bg-white border border-gray-200 rounded-xl p-2 mb-6 shadow-sm">
                <LineChart
                  data={oxygenData}
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
                24-Hour Summary
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-6">
                <View className="flex-row justify-between mb-2">
                  <Text className="text-gray-500">Average Temperature</Text>
                  <Text className="font-medium">28.4°C</Text>
                </View>
                <View className="flex-row justify-between mb-2">
                  <Text className="text-gray-500">Average Oxygen</Text>
                  <Text className="font-medium">6.2 mg/L</Text>
                </View>
                <View className="flex-row justify-between mb-2">
                  <Text className="text-gray-500">pH Range</Text>
                  <Text className="font-medium">7.0 - 7.4</Text>
                </View>
                <View className="flex-row justify-between">
                  <Text className="text-gray-500">Ammonia Peak</Text>
                  <Text className="font-medium">0.08 ppm</Text>
                </View>
              </View>

              {/* Blockchain Data Verification */}
              <Text className="text-lg font-semibold mb-4">
                Blockchain Data Verification
              </Text>
              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm">
                <Text className="text-gray-700 mb-3">
                  All sensor data is securely stored on the blockchain to ensure
                  data integrity and transparency.
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
                    All sensor readings are cryptographically signed
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
                    Complete historical record of all readings
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
        </View>
      </ScrollView>

      {/* Pond Details Modal */}
      <Modal
        animationType="slide"
        transparent={true}
        visible={isModalVisible}
        onRequestClose={() => setIsModalVisible(false)}
      >
        <View className="flex-1 bg-black/50 justify-end">
          <View className="bg-white rounded-t-3xl p-5 h-[70%]">
            <View className="flex-row justify-between items-center mb-6">
              <Text className="text-2xl font-bold">
                {selectedPond?.name} Details
              </Text>
              <TouchableOpacity
                className="p-2"
                onPress={() => setIsModalVisible(false)}
              >
                <TablerIconComponent name="x" size={24} color="#000" />
              </TouchableOpacity>
            </View>

            {selectedPond && (
              <ScrollView showsVerticalScrollIndicator={false}>
                <View className="bg-gradient-to-r from-primary to-primary-dark p-4 rounded-xl mb-5">
                  <Text className="text-white/80 text-sm">Batch ID</Text>
                  <Text className="text-white font-bold text-xl mb-2">
                    {selectedPond.batchId}
                  </Text>

                  <View className="flex-row flex-wrap">
                    <View className="w-1/2 mb-2">
                      <Text className="text-white/70 text-xs">Device ID</Text>
                      <Text className="text-white">
                        {selectedPond.deviceId}
                      </Text>
                    </View>
                    <View className="w-1/2 mb-2">
                      <Text className="text-white/70 text-xs">Status</Text>
                      <Text className="text-white">{selectedPond.status}</Text>
                    </View>
                  </View>
                </View>

                <Text className="text-lg font-semibold mb-4">
                  Current Readings
                </Text>

                <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-5">
                  <View className="flex-row justify-between mb-3">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="temperature"
                        size={20}
                        color="#f97316"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">Temperature</Text>
                    </View>
                    <Text className="text-primary font-bold">
                      {selectedPond.temperature}
                    </Text>
                  </View>

                  <View className="flex-row justify-between mb-3">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="droplet"
                        size={20}
                        color="#4338ca"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">Oxygen Level</Text>
                    </View>
                    <Text className="text-indigo-600 font-bold">
                      {selectedPond.oxygen}
                    </Text>
                  </View>

                  <View className="flex-row justify-between mb-3">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="chart-bar"
                        size={20}
                        color="#10b981"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">pH Level</Text>
                    </View>
                    <Text className="text-green-600 font-bold">
                      {selectedPond.ph}
                    </Text>
                  </View>

                  <View className="flex-row justify-between">
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="alert-triangle"
                        size={20}
                        color="#ef4444"
                        style={{ marginRight: 8 }}
                      />
                      <Text className="font-medium">Ammonia</Text>
                    </View>
                    <Text className="text-red-600 font-bold">
                      {selectedPond.ammonia}
                    </Text>
                  </View>
                </View>

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
                        All readings are verified on blockchain
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
                  <TouchableOpacity className="flex-1 bg-primary py-3 rounded-xl items-center">
                    <Text className="text-white font-bold">VIEW HISTORY</Text>
                  </TouchableOpacity>

                  <TouchableOpacity className="flex-1 bg-secondary py-3 rounded-xl items-center">
                    <Text className="text-white font-bold">EXPORT DATA</Text>
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
