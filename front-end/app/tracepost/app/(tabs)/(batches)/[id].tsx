import React, { useState, useEffect } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  RefreshControl,
  TextInput,
  Modal,
  KeyboardAvoidingView,
  Platform,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useLocalSearchParams, useRouter } from "expo-router";
import TablerIconComponent from "@/components/icon";
import { useRole } from "@/contexts/RoleContext";
import {
  getBatchById,
  getBatchBlockchainData,
  getBatchHistory,
  getBatchEnvironment,
  createEnvironmentData,
  getBatchEvents,
  BatchData,
  BatchBlockchainData,
  BatchHistoryData,
  BatchEnvironmentRecord,
  CreateEnvironmentRequest,
  Event,
} from "@/api/batch";
import { makeAuthenticatedRequest } from "@/api/auth";
import "@/global.css";

// Add interface for batch documents
interface BatchDocument {
  id: number;
  batch_id: number;
  document_type: string;
  document_name: string;
  file_url: string;
  file_hash: string;
  blockchain_tx_id?: string;
  uploaded_by: string;
  uploaded_at: string;
  verified_on_blockchain: boolean;
  document_size?: number;
  mime_type?: string;
}

interface GetBatchDocumentsResponse {
  success: boolean;
  message: string;
  data: BatchDocument[] | null;
}

export default function BatchDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const router = useRouter();
  const { isHatchery, userData } = useRole();

  // State management
  const [batch, setBatch] = useState<BatchData | null>(null);
  const [blockchainData, setBlockchainData] =
    useState<BatchBlockchainData | null>(null);
  const [historyData, setHistoryData] = useState<BatchHistoryData | null>(null);
  const [environmentData, setEnvironmentData] = useState<
    BatchEnvironmentRecord[]
  >([]);
  const [documentsData, setDocumentsData] = useState<BatchDocument[]>([]);
  const [batchEvents, setBatchEvents] = useState<Event[]>([]);
  const [showMoreModal, setShowMoreModal] = useState(false);

  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const allTabs = [
    "overview",
    "blockchain",
    "environment",
    "documents",
    "events",
  ] as const;
  type TabKey = (typeof allTabs)[number];
  const [activeTab, setActiveTab] = useState<TabKey>("overview");

  // Environment modal state
  const [isEnvironmentModalVisible, setIsEnvironmentModalVisible] =
    useState(false);
  const [isSubmittingEnvironment, setIsSubmittingEnvironment] = useState(false);
  const [environmentForm, setEnvironmentForm] =
    useState<CreateEnvironmentRequest>({
      age: 0,
      batch_id: parseInt(id || "0"),
      density: 0,
      ph: 0,
      salinity: 0,
      temperature: 0,
    });

  // Function to fetch batch documents
  const getBatchDocuments = async (
    batchId: number,
  ): Promise<GetBatchDocumentsResponse> => {
    try {
      const response = await makeAuthenticatedRequest(
        `${process.env.EXPO_PUBLIC_API_URL}/batches/${batchId}/documents`,
        {
          method: "GET",
        },
      );

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || "Failed to fetch documents");
      }

      const data: GetBatchDocumentsResponse = await response.json();
      return data;
    } catch (error) {
      console.error("Get Batch Documents API error:", error);
      throw error;
    }
  };

  // Load all batch data
  const loadBatchData = async () => {
    if (!id) return;

    try {
      setIsLoading(true);
      const batchId = parseInt(id);

      // Load batch basic info
      const batchResponse = await getBatchById(batchId);
      if (batchResponse.success) {
        setBatch(batchResponse.data);
      }

      // Load blockchain data
      try {
        const blockchainResponse = await getBatchBlockchainData(batchId);
        if (blockchainResponse.success) {
          setBlockchainData(blockchainResponse.data);
        }
      } catch (error) {
        console.log("Blockchain data not available:", error);
        setBlockchainData(null);
      }

      // Load history data
      try {
        const historyResponse = await getBatchHistory(batchId);
        if (historyResponse.success) {
          setHistoryData(historyResponse.data);
        }
      } catch (error) {
        console.log("History data not available:", error);
        setHistoryData(null);
      }

      // Load environment data - handle null case
      try {
        const environmentResponse = await getBatchEnvironment(batchId);
        if (environmentResponse.success) {
          // Handle case where data is null or empty array
          if (
            environmentResponse.data === null ||
            environmentResponse.data === undefined
          ) {
            setEnvironmentData([]);
          } else if (Array.isArray(environmentResponse.data)) {
            setEnvironmentData(environmentResponse.data);
          } else {
            // Handle case where data is a single object instead of array
            setEnvironmentData([environmentResponse.data]);
          }
        } else {
          setEnvironmentData([]);
        }
      } catch (error) {
        console.log("Environment data not available:", error);
        setEnvironmentData([]);
      }

      // Load documents data - handle null case
      try {
        const documentsResponse = await getBatchDocuments(batchId);
        if (documentsResponse.success) {
          if (
            documentsResponse.data === null ||
            documentsResponse.data === undefined
          ) {
            setDocumentsData([]);
          } else if (Array.isArray(documentsResponse.data)) {
            setDocumentsData(documentsResponse.data);
          } else {
            setDocumentsData([]);
          }
        } else {
          setDocumentsData([]);
        }
      } catch (error) {
        console.log("Documents data not available:", error);
        setDocumentsData([]);
      }
      try {
        const evResp = await getBatchEvents(batchId);
        if (evResp.success) setBatchEvents(evResp.data);
      } catch {
        /* ignore */
      }
    } catch (error) {
      console.error("Error loading batch data:", error);
      Alert.alert(
        "Error",
        error instanceof Error ? error.message : "Failed to load batch data",
      );
    } finally {
      setIsLoading(false);
    }
  };

  // Refresh data
  const handleRefresh = async () => {
    setIsRefreshing(true);
    await loadBatchData();
    setIsRefreshing(false);
  };

  // Submit environment data
  const handleSubmitEnvironment = async () => {
    if (!batch) return;

    setIsSubmittingEnvironment(true);
    try {
      const response = await createEnvironmentData(environmentForm);

      if (response.success) {
        Alert.alert("Success", "Environment data recorded successfully");
        setIsEnvironmentModalVisible(false);
        // Reset form
        setEnvironmentForm({
          age: 0,
          batch_id: batch.id,
          density: 0,
          ph: 0,
          salinity: 0,
          temperature: 0,
        });
        // Refresh environment data
        await handleRefresh();
      }
    } catch (error) {
      console.error("Error submitting environment data:", error);
      Alert.alert(
        "Error",
        error instanceof Error
          ? error.message
          : "Failed to record environment data",
      );
    } finally {
      setIsSubmittingEnvironment(false);
    }
  };

  useEffect(() => {
    loadBatchData();
  }, [id]);

  const inlineTabs = allTabs.slice(0, 3);
  const overflowTabs = allTabs.slice(3);

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
      case "in_transit":
        return "bg-yellow-100 text-yellow-700";
      case "delivered":
        return "bg-green-100 text-green-700";
      case "active":
        return "bg-green-100 text-green-700";
      case "completed":
        return "bg-gray-100 text-gray-700";
      default:
        return "bg-gray-100 text-gray-700";
    }
  };

  const getTransactionTypeIcon = (type: string) => {
    switch (type) {
      case "CREATE_BATCH":
        return "package";
      case "UPDATE_BATCH_STATUS":
        return "refresh";
      default:
        return "circle";
    }
  };

  const getDocumentTypeIcon = (documentType: string) => {
    switch (documentType.toLowerCase()) {
      case "transport":
        return "truck";
      case "certificate":
        return "certificate";
      case "invoice":
        return "receipt";
      case "health":
        return "health-recognition";
      case "quality":
        return "badge-check";
      default:
        return "file-text";
    }
  };

  const getDocumentTypeColor = (documentType: string) => {
    switch (documentType.toLowerCase()) {
      case "transport":
        return "#f97316"; // orange
      case "certificate":
        return "#10b981"; // green
      case "invoice":
        return "#3b82f6"; // blue
      case "health":
        return "#ef4444"; // red
      case "quality":
        return "#8b5cf6"; // purple
      default:
        return "#6b7280"; // gray
    }
  };

  const formatFileSize = (bytes?: number) => {
    if (!bytes) return "Unknown size";
    const sizes = ["Bytes", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round((bytes / Math.pow(1024, i)) * 100) / 100 + " " + sizes[i];
  };

  if (isLoading) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <View className="flex-1 justify-center items-center">
          <ActivityIndicator size="large" color="#f97316" />
          <Text className="text-gray-500 mt-4">Loading batch details...</Text>
        </View>
      </SafeAreaView>
    );
  }

  if (!batch) {
    return (
      <SafeAreaView className="flex-1 bg-white">
        <View className="flex-1 justify-center items-center px-6">
          <TablerIconComponent name="package-off" size={64} color="#9ca3af" />
          <Text className="text-gray-500 font-medium mt-4 mb-2 text-center">
            Batch not found
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
              <Text className="text-xl font-bold text-gray-800 text-center">
                Batch #{batch.id}
              </Text>
              <Text className="text-gray-500 text-center text-sm">
                {batch.species}
              </Text>
            </View>
            <TouchableOpacity
              className="h-10 w-10 rounded-full bg-primary/10 items-center justify-center"
              onPress={handleRefresh}
            >
              <TablerIconComponent name="refresh" size={20} color="#f97316" />
            </TouchableOpacity>
          </View>

          {/* Batch Overview Card */}
          <View className="bg-sky-300 p-5 rounded-xl mb-6">
            <View className="flex-row justify-between items-start mb-4">
              <View>
                <Text className="text-white/80 text-sm">Batch Status</Text>
                <Text className="text-white font-bold text-xl capitalize">
                  {blockchainData?.state.status || batch.status}
                </Text>
              </View>
              <View
                className={`px-3 py-1 rounded-full ${getStatusColor(
                  blockchainData?.state.status || batch.status,
                )}`}
              >
                <Text className="text-xs font-medium capitalize">
                  {blockchainData?.state.status || batch.status}
                </Text>
              </View>
            </View>

            <View className="flex-row flex-wrap">
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Quantity</Text>
                <Text className="text-white">
                  {batch.quantity.toLocaleString()} {""}
                  larvaes
                </Text>
              </View>
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Species</Text>
                <Text className="text-white">
                  {blockchainData?.state.species || batch.species}
                </Text>
              </View>
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Hatchery</Text>
                <Text className="text-white">{batch.hatchery.name}</Text>
              </View>
              <View className="w-1/2 mb-3">
                <Text className="text-white/70 text-xs">Created</Text>
                <Text className="text-white">
                  {formatDate(batch.created_at)}
                </Text>
              </View>
            </View>

            {blockchainData && (
              <View className="bg-white/20 p-3 rounded-lg mt-4">
                <View className="flex-row items-center">
                  <TablerIconComponent
                    name="currency-ethereum"
                    size={18}
                    color="white"
                  />
                  <Text className="text-white ml-2 font-medium">
                    Blockchain Verified • {blockchainData.tx_count} transactions
                  </Text>
                </View>
              </View>
            )}
          </View>

          {/* Tabs */}
          <View className="mb-4 flex-row items-center">
            {inlineTabs.map((tab) => (
              <TouchableOpacity
                key={tab}
                className={`px-4 py-2 rounded-full mr-2 ${
                  activeTab === tab ? "bg-primary" : "bg-gray-100"
                }`}
                onPress={() => setActiveTab(tab)}
              >
                <Text
                  className={`capitalize ${
                    activeTab === tab ? "text-white" : "text-gray-600"
                  }`}
                >
                  {tab}
                </Text>
              </TouchableOpacity>
            ))}

            {/* only show if there’s overflow */}
            {overflowTabs.length > 0 && (
              <TouchableOpacity
                className="px-3 py-2 rounded-full bg-gray-200"
                onPress={() => setShowMoreModal(true)}
              >
                <Text className="text-gray-600">⋯ More</Text>
              </TouchableOpacity>
            )}
          </View>

          {/* Tab Content */}
          {activeTab === "overview" && (
            <View>
              {/* Hatchery Information */}
              <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                <Text className="text-lg font-semibold mb-4">
                  Hatchery Information
                </Text>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Name</Text>
                  <Text className="font-medium">{batch.hatchery.name}</Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Company</Text>
                  <Text className="font-medium">
                    {batch.hatchery.company.name}
                  </Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Location</Text>
                  <Text className="font-medium">
                    {batch.hatchery.company.location}
                  </Text>
                </View>

                <View className="mb-3">
                  <Text className="text-gray-600 text-sm">Contact</Text>
                  <Text className="font-medium">
                    {batch.hatchery.company.contact_info}
                  </Text>
                </View>

                <View className="flex-row items-center justify-between pt-3 border-t border-gray-100">
                  <Text className="text-gray-500 text-xs">
                    Company Type: {batch.hatchery.company.type}
                  </Text>
                  <View
                    className={`px-2 py-1 rounded ${
                      batch.hatchery.is_active ? "bg-green-100" : "bg-red-100"
                    }`}
                  >
                    <Text
                      className={`text-xs ${
                        batch.hatchery.is_active
                          ? "text-green-600"
                          : "text-red-600"
                      }`}
                    >
                      {batch.hatchery.is_active ? "Active" : "Inactive"}
                    </Text>
                  </View>
                </View>
              </View>

              {/* Batch Statistics */}
              <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                <Text className="text-lg font-semibold mb-4">
                  Batch Statistics
                </Text>

                <View className="flex-row flex-wrap">
                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">
                      Initial Quantity
                    </Text>
                    <Text className="font-medium text-lg">
                      {batch.quantity.toLocaleString()}
                    </Text>
                  </View>

                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">
                      Current Status
                    </Text>
                    <Text className="font-medium text-lg capitalize">
                      {batch.status}
                    </Text>
                  </View>

                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">Days Active</Text>
                    <Text className="font-medium text-lg">
                      {Math.floor(
                        (new Date().getTime() -
                          new Date(batch.created_at).getTime()) /
                          (1000 * 60 * 60 * 24),
                      )}
                    </Text>
                  </View>

                  <View className="w-1/2 mb-3">
                    <Text className="text-gray-600 text-sm">Last Updated</Text>
                    <Text className="font-medium text-lg">
                      {formatDate(batch.updated_at)}
                    </Text>
                  </View>
                </View>
              </View>
            </View>
          )}

          {activeTab === "blockchain" && (
            <View>
              {blockchainData ? (
                <>
                  {/* Blockchain Summary */}
                  <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                    <Text className="text-lg font-semibold mb-4">
                      Blockchain Summary
                    </Text>

                    <View className="flex-row flex-wrap">
                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">
                          Total Transactions
                        </Text>
                        <Text className="font-medium text-lg">
                          {blockchainData.tx_count}
                        </Text>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">
                          First Transaction
                        </Text>
                        <Text className="font-medium text-sm">
                          {formatDate(blockchainData.first_tx)}
                        </Text>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">
                          Latest Transaction
                        </Text>
                        <Text className="font-medium text-sm">
                          {formatDate(blockchainData.latest_tx)}
                        </Text>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">Batch ID</Text>
                        <Text className="font-medium text-lg">
                          {blockchainData.batch_id}
                        </Text>
                      </View>
                    </View>
                  </View>

                  {/* Transaction Timeline */}
                  <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                    <Text className="text-lg font-semibold mb-4">
                      Transaction Timeline
                    </Text>

                    {blockchainData.txs.map((tx, index) => (
                      <View key={tx.tx_id} className="mb-4 last:mb-0">
                        <View className="flex-row">
                          <View className="items-center mr-4">
                            <View className="h-10 w-10 rounded-full bg-indigo-100 items-center justify-center z-10">
                              <TablerIconComponent
                                name={getTransactionTypeIcon(tx.type)}
                                size={20}
                                color="#4338ca"
                              />
                            </View>
                            {index < blockchainData.txs.length - 1 && (
                              <View className="h-full w-0.5 bg-gray-200 absolute top-10 bottom-0 left-5" />
                            )}
                          </View>

                          <View className="flex-1">
                            <Text className="font-semibold">
                              {tx.type
                                .replace("_", " ")
                                .toLowerCase()
                                .replace(/\b\w/g, (l) => l.toUpperCase())}
                            </Text>
                            <Text className="text-gray-500 text-xs mb-1">
                              {formatDate(tx.timestamp)}
                            </Text>
                            <Text className="text-gray-600 mb-2">
                              Transaction ID: {tx.tx_id}
                            </Text>
                            <View className="bg-gray-50 p-2 rounded">
                              <Text className="text-xs text-gray-600">
                                Validated: {formatDate(tx.validated_at)}
                              </Text>
                            </View>
                          </View>
                        </View>
                      </View>
                    ))}
                  </View>

                  {/* History Data */}
                  {historyData && (
                    <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                      <Text className="text-lg font-semibold mb-4">
                        History Verification
                      </Text>

                      <View className="flex-row items-center mb-4">
                        <TablerIconComponent
                          name={
                            historyData.verifiable_history
                              ? "shield-check"
                              : "shield-x"
                          }
                          size={20}
                          color={
                            historyData.verifiable_history
                              ? "#10b981"
                              : "#ef4444"
                          }
                        />
                        <Text
                          className={`ml-2 font-medium ${historyData.verifiable_history ? "text-green-600" : "text-red-600"}`}
                        >
                          {historyData.verifiable_history
                            ? "History Verified"
                            : "History Unverified"}
                        </Text>
                      </View>

                      <View className="mb-3">
                        <Text className="text-gray-600 text-sm">
                          Batch Events
                        </Text>
                        <Text className="font-medium">
                          {historyData.batch_events.length}
                        </Text>
                      </View>

                      <View className="mb-3">
                        <Text className="text-gray-600 text-sm">
                          Database Records
                        </Text>
                        <Text className="font-medium">
                          {historyData.db_records.length}
                        </Text>
                      </View>
                    </View>
                  )}
                </>
              ) : (
                <View className="bg-gray-50 p-8 rounded-xl items-center">
                  <TablerIconComponent
                    name="database-off"
                    size={48}
                    color="#9ca3af"
                  />
                  <Text className="text-gray-500 font-medium mt-4 mb-2">
                    No blockchain data available
                  </Text>
                  <Text className="text-gray-400 text-center">
                    This batch may not have been recorded on the blockchain yet
                  </Text>
                </View>
              )}
            </View>
          )}

          {activeTab === "environment" && (
            <View>
              {/* Environment Controls - Only for hatchery */}
              {isHatchery && (
                <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                  <View className="flex-row items-center justify-between mb-4">
                    <Text className="text-lg font-semibold">
                      Environment Control
                    </Text>
                    <TouchableOpacity
                      className="bg-primary px-4 py-2 rounded-lg flex-row items-center"
                      onPress={() => setIsEnvironmentModalVisible(true)}
                    >
                      <TablerIconComponent
                        name="plus"
                        size={16}
                        color="white"
                      />
                      <Text className="text-white ml-1 font-medium">
                        Add Data
                      </Text>
                    </TouchableOpacity>
                  </View>

                  <Text className="text-gray-600 text-sm">
                    Record environmental parameters for this batch to track
                    optimal growing conditions.
                  </Text>
                </View>
              )}

              {/* Environment Data */}
              {environmentData && environmentData.length > 0 ? (
                environmentData.map((envRecord) => (
                  <View
                    key={envRecord.id}
                    className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
                  >
                    <View className="flex-row justify-between items-start mb-4">
                      <Text className="text-lg font-semibold">
                        Environment Record #{envRecord.id}
                      </Text>
                      <Text className="text-gray-500 text-xs">
                        {formatDate(envRecord.environment_data.timestamp)}
                      </Text>
                    </View>

                    <View className="flex-row flex-wrap">
                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">
                          Temperature
                        </Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="temperature"
                            size={16}
                            color="#f97316"
                          />
                          <Text className="ml-1 font-medium">
                            {envRecord.environment_data.temperature}°C
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">pH Level</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="chart-bar"
                            size={16}
                            color="#10b981"
                          />
                          <Text className="ml-1 font-medium">
                            {envRecord.environment_data.ph}
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">Salinity</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="droplet"
                            size={16}
                            color="#3b82f6"
                          />
                          <Text className="ml-1 font-medium">
                            {envRecord.environment_data.salinity} ppt
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">Density</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="fish"
                            size={16}
                            color="#8b5cf6"
                          />
                          <Text className="ml-1 font-medium">
                            {envRecord.environment_data.density}/m²
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">Age</Text>
                        <View className="flex-row items-center">
                          <TablerIconComponent
                            name="calendar"
                            size={16}
                            color="#f59e0b"
                          />
                          <Text className="ml-1 font-medium">
                            {envRecord.environment_data.age} days
                          </Text>
                        </View>
                      </View>

                      <View className="w-1/2 mb-3">
                        <Text className="text-gray-600 text-sm">Status</Text>
                        <View
                          className={`px-2 py-1 rounded flex-row items-center self-start ${
                            envRecord.environment_data.is_active
                              ? "bg-green-100"
                              : "bg-red-100"
                          }`}
                        >
                          <Text
                            className={`text-xs ${
                              envRecord.environment_data.is_active
                                ? "text-green-600"
                                : "text-red-600"
                            }`}
                          >
                            {envRecord.environment_data.is_active
                              ? "Active"
                              : "Inactive"}
                          </Text>
                        </View>
                      </View>
                    </View>

                    {/* Facility Information */}
                    <View className="mt-4 pt-4 border-t border-gray-100">
                      <Text className="text-gray-600 text-sm mb-2">
                        Facility Information
                      </Text>
                      <View className="flex-row justify-between items-center">
                        <Text className="text-xs text-gray-500">
                          {envRecord.facility_info.hatchery_name} •{" "}
                          {envRecord.facility_info.company_name}
                        </Text>
                        <Text className="text-xs text-gray-500">
                          {envRecord.facility_info.company_location}
                        </Text>
                      </View>
                    </View>
                  </View>
                ))
              ) : (
                <View className="bg-gray-50 p-8 rounded-xl items-center">
                  <TablerIconComponent
                    name="chart-dots"
                    size={48}
                    color="#9ca3af"
                  />
                  <Text className="text-gray-500 font-medium mt-4 mb-2">
                    No environment data available
                  </Text>
                  <Text className="text-gray-400 text-center mb-4">
                    {isHatchery
                      ? "Start monitoring by adding the first environment reading for this batch"
                      : "Environment data will appear here when the hatchery starts recording measurements"}
                  </Text>

                  {/* Show different actions based on role */}
                  {isHatchery ? (
                    <TouchableOpacity
                      className="bg-primary px-6 py-3 rounded-xl flex-row items-center"
                      onPress={() => setIsEnvironmentModalVisible(true)}
                    >
                      <TablerIconComponent
                        name="plus"
                        size={18}
                        color="white"
                      />
                      <Text className="text-white font-bold ml-2">
                        Add First Reading
                      </Text>
                    </TouchableOpacity>
                  ) : (
                    <View className="bg-blue-50 p-4 rounded-lg">
                      <View className="flex-row items-center justify-center">
                        <TablerIconComponent
                          name="info-circle"
                          size={16}
                          color="#3b82f6"
                        />
                        <Text className="text-blue-700 text-sm ml-2">
                          Contact the hatchery for environment monitoring data
                        </Text>
                      </View>
                    </View>
                  )}
                </View>
              )}

              {/* Environment Data Summary - Only show if data exists */}
              {environmentData && environmentData.length > 0 && (
                <View className="bg-white border border-gray-200 rounded-xl p-4 mt-4 shadow-sm">
                  <Text className="text-lg font-semibold mb-4">
                    Environment Summary
                  </Text>

                  <View className="flex-row flex-wrap">
                    <View className="w-1/3 mb-3">
                      <Text className="text-gray-600 text-xs">
                        Total Records
                      </Text>
                      <Text className="font-bold text-lg text-primary">
                        {environmentData.length}
                      </Text>
                    </View>

                    <View className="w-1/3 mb-3">
                      <Text className="text-gray-600 text-xs">
                        Latest Record
                      </Text>
                      <Text className="font-medium text-sm">
                        {formatDate(
                          environmentData[environmentData.length - 1]
                            ?.environment_data.timestamp || "",
                        )}
                      </Text>
                    </View>

                    <View className="w-1/3 mb-3">
                      <Text className="text-gray-600 text-xs">
                        Active Status
                      </Text>
                      <View
                        className={`px-2 py-1 rounded self-start ${
                          environmentData[environmentData.length - 1]
                            ?.environment_data.is_active
                            ? "bg-green-100"
                            : "bg-red-100"
                        }`}
                      >
                        <Text
                          className={`text-xs font-medium ${
                            environmentData[environmentData.length - 1]
                              ?.environment_data.is_active
                              ? "text-green-600"
                              : "text-red-600"
                          }`}
                        >
                          {environmentData[environmentData.length - 1]
                            ?.environment_data.is_active
                            ? "Active"
                            : "Inactive"}
                        </Text>
                      </View>
                    </View>
                  </View>

                  {/* Latest values */}
                  {environmentData.length > 0 && (
                    <View className="mt-4 pt-4 border-t border-gray-100">
                      <Text className="text-gray-600 text-sm mb-3">
                        Latest Measurements
                      </Text>
                      <View className="flex-row flex-wrap">
                        <View className="w-1/4 mb-2">
                          <Text className="text-xs text-gray-500">Temp</Text>
                          <Text className="font-medium text-sm">
                            {
                              environmentData[environmentData.length - 1]
                                ?.environment_data.temperature
                            }
                            °C
                          </Text>
                        </View>
                        <View className="w-1/4 mb-2">
                          <Text className="text-xs text-gray-500">pH</Text>
                          <Text className="font-medium text-sm">
                            {
                              environmentData[environmentData.length - 1]
                                ?.environment_data.ph
                            }
                          </Text>
                        </View>
                        <View className="w-1/4 mb-2">
                          <Text className="text-xs text-gray-500">
                            Salinity
                          </Text>
                          <Text className="font-medium text-sm">
                            {
                              environmentData[environmentData.length - 1]
                                ?.environment_data.salinity
                            }{" "}
                            ppt
                          </Text>
                        </View>
                        <View className="w-1/4 mb-2">
                          <Text className="text-xs text-gray-500">Age</Text>
                          <Text className="font-medium text-sm">
                            {
                              environmentData[environmentData.length - 1]
                                ?.environment_data.age
                            }
                            d
                          </Text>
                        </View>
                      </View>
                    </View>
                  )}
                </View>
              )}
            </View>
          )}

          {activeTab === "documents" && (
            <View>
              {/* Documents Header */}
              <View className="bg-white border border-gray-200 rounded-xl p-4 mb-6 shadow-sm">
                <View className="flex-row items-center justify-between mb-4">
                  <Text className="text-lg font-semibold">Batch Documents</Text>
                  {isHatchery && (
                    <TouchableOpacity className="bg-primary px-4 py-2 rounded-lg flex-row items-center">
                      <TablerIconComponent
                        name="upload"
                        size={16}
                        color="white"
                      />
                      <Text className="text-white ml-1 font-medium">
                        Upload
                      </Text>
                    </TouchableOpacity>
                  )}
                </View>
                <Text className="text-gray-600 text-sm">
                  Transport documents and certificates are stored on the
                  blockchain network for immutable verification.
                </Text>
              </View>

              {/* Documents List */}
              {documentsData && documentsData.length > 0 ? (
                documentsData.map((document) => (
                  <View
                    key={document.id}
                    className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm"
                  >
                    <View className="flex-row items-start mb-3">
                      <View
                        className="h-12 w-12 rounded-xl items-center justify-center mr-3"
                        style={{
                          backgroundColor: `${getDocumentTypeColor(document.document_type)}15`,
                        }}
                      >
                        <TablerIconComponent
                          name={getDocumentTypeIcon(document.document_type)}
                          size={24}
                          color={getDocumentTypeColor(document.document_type)}
                        />
                      </View>

                      <View className="flex-1">
                        <View className="flex-row items-center justify-between mb-1">
                          <Text className="font-semibold text-lg">
                            {document.document_name}
                          </Text>
                          <View
                            className={`px-2 py-1 rounded-full ${
                              document.verified_on_blockchain
                                ? "bg-green-100"
                                : "bg-yellow-100"
                            }`}
                          >
                            <Text
                              className={`text-xs font-medium ${
                                document.verified_on_blockchain
                                  ? "text-green-700"
                                  : "text-yellow-700"
                              }`}
                            >
                              {document.verified_on_blockchain
                                ? "Verified"
                                : "Pending"}
                            </Text>
                          </View>
                        </View>

                        <View className="mb-2">
                          <Text className="text-gray-600 text-sm capitalize">
                            {document.document_type.replace("_", " ")} Document
                          </Text>
                          <Text className="text-gray-500 text-xs">
                            Uploaded by {document.uploaded_by} •{" "}
                            {formatDate(document.uploaded_at)}
                          </Text>
                        </View>

                        <View className="flex-row flex-wrap mb-3">
                          <View className="w-1/2 mb-2">
                            <Text className="text-gray-500 text-xs">
                              File Size
                            </Text>
                            <Text className="font-medium text-sm">
                              {formatFileSize(document.document_size)}
                            </Text>
                          </View>
                          <View className="w-1/2 mb-2">
                            <Text className="text-gray-500 text-xs">
                              File Type
                            </Text>
                            <Text className="font-medium text-sm">
                              {document.mime_type || "Unknown"}
                            </Text>
                          </View>
                        </View>

                        {/* Blockchain Information */}
                        {document.verified_on_blockchain &&
                          document.blockchain_tx_id && (
                            <View className="bg-indigo-50 p-3 rounded-lg mb-3">
                              <View className="flex-row items-center mb-1">
                                <TablerIconComponent
                                  name="currency-ethereum"
                                  size={14}
                                  color="#4338ca"
                                />
                                <Text className="text-indigo-700 text-xs font-medium ml-1">
                                  Blockchain Verified
                                </Text>
                              </View>
                              <Text className="text-indigo-600 text-xs">
                                TX: {document.blockchain_tx_id}
                              </Text>
                              <Text className="text-indigo-600 text-xs">
                                Hash: {document.file_hash}
                              </Text>
                            </View>
                          )}

                        {/* Action Buttons */}
                        <View className="flex-row gap-2">
                          <TouchableOpacity className="flex-1 bg-gray-100 py-2 rounded-lg items-center">
                            <View className="flex-row items-center">
                              <TablerIconComponent
                                name="download"
                                size={16}
                                color="#6b7280"
                              />
                              <Text className="text-gray-700 text-sm font-medium ml-1">
                                Download
                              </Text>
                            </View>
                          </TouchableOpacity>

                          {document.verified_on_blockchain && (
                            <TouchableOpacity className="flex-1 bg-indigo-100 py-2 rounded-lg items-center">
                              <View className="flex-row items-center">
                                <TablerIconComponent
                                  name="external-link"
                                  size={16}
                                  color="#4338ca"
                                />
                                <Text className="text-indigo-700 text-sm font-medium ml-1">
                                  View on Chain
                                </Text>
                              </View>
                            </TouchableOpacity>
                          )}
                        </View>
                      </View>
                    </View>
                  </View>
                ))
              ) : (
                <View className="bg-gray-50 p-8 rounded-xl items-center">
                  <TablerIconComponent
                    name="file-off"
                    size={48}
                    color="#9ca3af"
                  />
                  <Text className="text-gray-500 font-medium mt-4 mb-2">
                    No documents available
                  </Text>
                  <Text className="text-gray-400 text-center mb-4">
                    {isHatchery
                      ? "Upload transport documents and certificates to track them on blockchain"
                      : "Documents will appear here when the hatchery uploads transport and certification documents"}
                  </Text>

                  {isHatchery && (
                    <TouchableOpacity className="bg-primary px-6 py-3 rounded-xl flex-row items-center">
                      <TablerIconComponent
                        name="upload"
                        size={18}
                        color="white"
                      />
                      <Text className="text-white font-bold ml-2">
                        Upload First Document
                      </Text>
                    </TouchableOpacity>
                  )}
                </View>
              )}

              {/* Documents Summary - Only show if documents exist */}
              {documentsData && documentsData.length > 0 && (
                <View className="bg-white border border-gray-200 rounded-xl p-4 mt-4 shadow-sm">
                  <Text className="text-lg font-semibold mb-4">
                    Documents Summary
                  </Text>

                  <View className="flex-row flex-wrap">
                    <View className="w-1/3 mb-3">
                      <Text className="text-gray-600 text-xs">
                        Total Documents
                      </Text>
                      <Text className="font-bold text-lg text-primary">
                        {documentsData.length}
                      </Text>
                    </View>

                    <View className="w-1/3 mb-3">
                      <Text className="text-gray-600 text-xs">
                        Verified on Chain
                      </Text>
                      <Text className="font-bold text-lg text-green-600">
                        {
                          documentsData.filter(
                            (doc) => doc.verified_on_blockchain,
                          ).length
                        }
                      </Text>
                    </View>

                    <View className="w-1/3 mb-3">
                      <Text className="text-gray-600 text-xs">
                        Pending Verification
                      </Text>
                      <Text className="font-bold text-lg text-yellow-600">
                        {
                          documentsData.filter(
                            (doc) => !doc.verified_on_blockchain,
                          ).length
                        }
                      </Text>
                    </View>
                  </View>

                  {/* Document Types Breakdown */}
                  <View className="mt-4 pt-4 border-t border-gray-100">
                    <Text className="text-gray-600 text-sm mb-3">
                      Document Types
                    </Text>
                    <View className="flex-row flex-wrap">
                      {Object.entries(
                        documentsData.reduce(
                          (acc, doc) => {
                            acc[doc.document_type] =
                              (acc[doc.document_type] || 0) + 1;
                            return acc;
                          },
                          {} as Record<string, number>,
                        ),
                      ).map(([type, count]) => (
                        <View key={type} className="w-1/2 mb-2">
                          <View className="flex-row items-center">
                            <TablerIconComponent
                              name={getDocumentTypeIcon(type)}
                              size={14}
                              color={getDocumentTypeColor(type)}
                            />
                            <Text className="text-xs text-gray-500 ml-1 capitalize">
                              {type.replace("_", " ")}
                            </Text>
                          </View>
                          <Text className="font-medium text-sm">{count}</Text>
                        </View>
                      ))}
                    </View>
                  </View>
                </View>
              )}
            </View>
          )}

          {activeTab === "events" && (
            <View>
              <Text className="text-lg font-semibold mb-4">Batch Events</Text>
              {batchEvents.length === 0 ? (
                <View className="items-center p-8 bg-gray-50 rounded-xl">
                  <Text className="text-gray-500">No events recorded.</Text>
                </View>
              ) : (
                batchEvents.map((evt) => (
                  <View
                    key={evt.id}
                    className="bg-white border border-gray-200 rounded-xl p-4 mb-3 shadow-sm"
                  >
                    <View className="flex-row justify-between items-center mb-2">
                      <Text className="font-medium capitalize">
                        {evt.event_type.replace("_", " ")}
                      </Text>
                      <Text className="text-gray-400 text-xs">
                        {new Date(evt.timestamp).toLocaleString()}
                      </Text>
                    </View>
                    <Text className="text-gray-600 text-sm mb-1">
                      Location: {evt.location}
                    </Text>
                    <Text className="text-gray-600 text-sm mb-1">
                      Status: {evt.batch_info.status}
                    </Text>
                    <Text className="text-gray-600 text-sm mb-1">
                      Quantity: {evt.batch_info.quantity}
                    </Text>
                    {/* any metadata */}
                    {evt.metadata &&
                      Object.entries(evt.metadata).map(([k, v]) => (
                        <Text key={k} className="text-xs text-gray-500">
                          {k}: {String(v)}
                        </Text>
                      ))}
                  </View>
                ))
              )}
            </View>
          )}
          {/* Quick Actions */}
          <View className="mt-6">
            <Text className="text-lg font-semibold mb-4">Quick Actions</Text>
            <View className="flex-row flex-wrap gap-3">
              <TouchableOpacity
                className="flex-1 bg-primary/10 p-4 rounded-xl items-center min-w-[45%]"
                onPress={handleRefresh}
              >
                <TablerIconComponent name="refresh" size={24} color="#f97316" />
                <Text className="text-primary font-medium mt-2">
                  Refresh Data
                </Text>
              </TouchableOpacity>

              <TouchableOpacity className="flex-1 bg-blue-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent name="qrcode" size={24} color="#3b82f6" />
                <Text className="text-blue-600 font-medium mt-2">
                  Generate QR
                </Text>
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

              <TouchableOpacity className="flex-1 bg-indigo-50 p-4 rounded-xl items-center min-w-[45%]">
                <TablerIconComponent name="share" size={24} color="#4338ca" />
                <Text className="text-indigo-600 font-medium mt-2">Share</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>

      {/* Environment Data Modal */}
      <Modal
        animationType="slide"
        transparent={true}
        visible={isEnvironmentModalVisible}
        onRequestClose={() => setIsEnvironmentModalVisible(false)}
      >
        <KeyboardAvoidingView
          className="flex-1 bg-black/50 justify-end"
          behavior={Platform.OS === "ios" ? "padding" : "height"}
        >
          <View className="bg-white rounded-t-3xl p-5 max-h-[80%]">
            <View className="flex-row justify-between items-center mb-6">
              <Text className="text-xl font-bold">Add Environment Data</Text>
              <TouchableOpacity
                className="p-2"
                onPress={() => setIsEnvironmentModalVisible(false)}
              >
                <TablerIconComponent name="x" size={24} color="#000" />
              </TouchableOpacity>
            </View>

            <ScrollView showsVerticalScrollIndicator={false}>
              <View className="mb-4">
                <Text className="font-medium text-gray-700 mb-1">
                  Temperature (°C)
                </Text>
                <TextInput
                  className="p-3 border border-gray-300 rounded-xl bg-white"
                  placeholder="Enter temperature"
                  keyboardType="numeric"
                  value={environmentForm.temperature.toString()}
                  onChangeText={(text) =>
                    setEnvironmentForm({
                      ...environmentForm,
                      temperature: parseFloat(text) || 0,
                    })
                  }
                />
              </View>

              <View className="mb-4">
                <Text className="font-medium text-gray-700 mb-1">pH Level</Text>
                <TextInput
                  className="p-3 border border-gray-300 rounded-xl bg-white"
                  placeholder="Enter pH level"
                  keyboardType="numeric"
                  value={environmentForm.ph.toString()}
                  onChangeText={(text) =>
                    setEnvironmentForm({
                      ...environmentForm,
                      ph: parseFloat(text) || 0,
                    })
                  }
                />
              </View>

              <View className="mb-4">
                <Text className="font-medium text-gray-700 mb-1">
                  Salinity (ppt)
                </Text>
                <TextInput
                  className="p-3 border border-gray-300 rounded-xl bg-white"
                  placeholder="Enter salinity"
                  keyboardType="numeric"
                  value={environmentForm.salinity.toString()}
                  onChangeText={(text) =>
                    setEnvironmentForm({
                      ...environmentForm,
                      salinity: parseFloat(text) || 0,
                    })
                  }
                />
              </View>

              <View className="mb-4">
                <Text className="font-medium text-gray-700 mb-1">
                  Density (per m²)
                </Text>
                <TextInput
                  className="p-3 border border-gray-300 rounded-xl bg-white"
                  placeholder="Enter density"
                  keyboardType="numeric"
                  value={environmentForm.density.toString()}
                  onChangeText={(text) =>
                    setEnvironmentForm({
                      ...environmentForm,
                      density: parseFloat(text) || 0,
                    })
                  }
                />
              </View>

              <View className="mb-6">
                <Text className="font-medium text-gray-700 mb-1">
                  Age (days)
                </Text>
                <TextInput
                  className="p-3 border border-gray-300 rounded-xl bg-white"
                  placeholder="Enter age in days"
                  keyboardType="numeric"
                  value={environmentForm.age.toString()}
                  onChangeText={(text) =>
                    setEnvironmentForm({
                      ...environmentForm,
                      age: parseInt(text) || 0,
                    })
                  }
                />
              </View>

              <TouchableOpacity
                className={`py-4 rounded-xl items-center ${
                  isSubmittingEnvironment ? "bg-primary/60" : "bg-primary"
                }`}
                onPress={handleSubmitEnvironment}
                disabled={isSubmittingEnvironment}
              >
                {isSubmittingEnvironment ? (
                  <View className="flex-row items-center">
                    <ActivityIndicator color="white" size="small" />
                    <Text className="font-bold text-white ml-2">
                      Recording...
                    </Text>
                  </View>
                ) : (
                  <Text className="font-bold text-white text-lg">
                    Record Environment Data
                  </Text>
                )}
              </TouchableOpacity>
            </ScrollView>
          </View>
        </KeyboardAvoidingView>
      </Modal>

      <Modal
        transparent
        visible={showMoreModal}
        animationType="none"
        onRequestClose={() => setShowMoreModal(false)}
      >
        <View className="flex-1 bg-black/50 justify-end">
          <View className="bg-white rounded-t-2xl p-4">
            {overflowTabs.map((tab) => (
              <TouchableOpacity
                key={tab}
                className="py-3 border-b border-gray-200"
                onPress={() => {
                  setActiveTab(tab);
                  setShowMoreModal(false);
                }}
              >
                <Text className="text-lg capitalize">{tab}</Text>
              </TouchableOpacity>
            ))}
            <TouchableOpacity
              className="mt-4 py-3 items-center"
              onPress={() => setShowMoreModal(false)}
            >
              <Text className="text-red-500">Cancel</Text>
            </TouchableOpacity>
          </View>
        </View>
      </Modal>
    </SafeAreaView>
  );
}
