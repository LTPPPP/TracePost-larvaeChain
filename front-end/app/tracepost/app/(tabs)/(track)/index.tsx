import React, { useState, useEffect } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  Image,
  Linking,
  StyleSheet,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import TablerIconComponent from "@/components/icon";
import { CameraView, Camera } from "expo-camera";
import "@/global.css";

export default function TrackScreen() {
  const [hasPermission, setHasPermission] = useState<boolean | null>(null);
  const [scanning, setScanning] = useState(false);
  const [scannedData, setScannedData] = useState<null | {
    batchId: string;
    origin: string;
    harvestDate: string;
    certificationDate: string;
    nftId: string;
    blockchain: string;
  }>(null);
  const [timeline, setTimeline] = useState<any[]>([]);
  const [verificationStatus, setVerificationStatus] = useState<
    "verified" | "pending" | "invalid" | null
  >(null);

  useEffect(() => {
    // Request camera permissions when component loads
    const requestPermissions = async () => {
      const { status } = await Camera.requestCameraPermissionsAsync();
      setHasPermission(status === "granted");
    };

    requestPermissions();
  }, []);

  const startScan = () => {
    if (hasPermission) {
      setScanning(true);
    } else {
      alert("Camera permission is required to scan QR codes");
    }
  };

  const handleBarCodeScanned = ({
    type,
    data,
  }: {
    type: string;
    data: string;
  }) => {
    setScanning(false);
    setVerificationStatus("pending");

    // Simulate blockchain verification process
    setTimeout(() => {
      // Mock data - in real app, this would be from the QR code or an API call
      const mockBatchData = {
        batchId: "SH-2023-10-B42",
        origin: "Mekong Delta Hatchery",
        harvestDate: "2023-10-15",
        certificationDate: "2023-10-18",
        nftId: "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
        blockchain: "Ethereum",
      };

      const mockTimeline = [
        {
          id: 1,
          date: "2023-09-01",
          event: "Breeding",
          details: "Initial breeding of parent shrimp",
          icon: "egg",
          color: "#f97316",
          txHash:
            "0x3a4e813ea3bf9913613ee7a1bea26e02e85f9ea9a5929ec4bcb76a4c8693a581",
        },
        {
          id: 2,
          date: "2023-09-15",
          event: "Larvae Stage",
          details: "Transition to larvae growing stage",
          icon: "fish",
          color: "#3b82f6",
          txHash:
            "0x7b91b7c1d8b9a89c8a65e06e4a4f8f0c9c6f6d4c9a8b7c6d5e4f3a2b1c0d9e8f",
        },
        {
          id: 3,
          date: "2023-09-30",
          event: "Feeding Phase",
          details: "Regular feeding and monitoring",
          icon: "bucket",
          color: "#10b981",
          txHash:
            "0xa1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
        },
        {
          id: 4,
          date: "2023-10-12",
          event: "Quality Check",
          details: "Pre-harvest quality assurance test",
          icon: "clipboard-check",
          color: "#8b5cf6",
          txHash:
            "0xd9e8f7c6b5a4d3c2b1a0f9e8d7c6b5a4d3c2b1a0f9e8d7c6b5a4d3c2b1a0f9e8",
        },
        {
          id: 5,
          date: "2023-10-15",
          event: "Harvesting",
          details: "Batch harvested and prepared for shipping",
          icon: "scale",
          color: "#f97316",
          txHash:
            "0x5e4d3c2b1a0f9e8d7c6b5a4d3c2b1a0f9e8d7c6b5a4d3c2b1a0f9e8d7c6b5a4",
        },
        {
          id: 6,
          date: "2023-10-18",
          event: "Certification",
          details: "Blockchain certification issued",
          icon: "certificate",
          color: "#4338ca",
          txHash:
            "0x1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2",
        },
      ];

      setScannedData(mockBatchData);
      setTimeline(mockTimeline);
      setVerificationStatus("verified");
    }, 2000);
  };

  const viewOnBlockchain = (txHash?: string) => {
    const url = txHash
      ? `https://etherscan.io/tx/${txHash}`
      : `https://etherscan.io/address/${scannedData?.nftId}`;

    Linking.openURL(url);
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
                Traceability
              </Text>
              <Text className="text-gray-500">
                Track shrimp batches on blockchain
              </Text>
            </View>
            <TouchableOpacity className="h-10 w-10 rounded-full bg-secondary/10 items-center justify-center">
              <TablerIconComponent name="search" size={20} color="#4338ca" />
            </TouchableOpacity>
          </View>

          {/* Web3 Badge */}
          <View className="bg-indigo-50 px-4 py-3 rounded-xl mb-6 flex-row items-center">
            <TablerIconComponent
              name="currency-ethereum"
              size={20}
              color="#4338ca"
            />
            <Text className="ml-2 text-indigo-700 flex-1">
              Verify authenticity with blockchain-backed traceability
            </Text>
          </View>

          {/* Scanner Section */}
          {scanning ? (
            <View className="h-80 rounded-xl overflow-hidden mb-6">
              <CameraView
                onBarcodeScanned={handleBarCodeScanned}
                barcodeScannerSettings={{
                  barcodeTypes: ["qr"],
                }}
                style={StyleSheet.absoluteFill}
              />
              <TouchableOpacity
                className="absolute top-4 right-4 bg-white/80 p-2 rounded-full"
                onPress={() => setScanning(false)}
              >
                <TablerIconComponent name="x" size={24} color="#000" />
              </TouchableOpacity>
            </View>
          ) : (
            <View className="bg-white border border-gray-200 rounded-xl p-6 shadow-sm mb-6 items-center">
              <View className="h-20 w-20 rounded-full bg-secondary/10 items-center justify-center mb-4">
                <TablerIconComponent name="qrcode" size={40} color="#4338ca" />
              </View>
              <Text className="text-lg font-semibold text-center mb-2">
                Scan QR Code
              </Text>
              <Text className="text-gray-500 text-center mb-6">
                Scan a batch QR code to verify its origin and complete lifecycle
                on the blockchain
              </Text>
              <TouchableOpacity
                className="bg-secondary py-3 px-8 rounded-xl"
                onPress={startScan}
              >
                <Text className="text-white font-bold">START SCANNING</Text>
              </TouchableOpacity>
            </View>
          )}

          {/* Verification Status Indicator */}
          {verificationStatus === "pending" && (
            <View className="bg-yellow-50 rounded-xl p-4 mb-6 flex-row items-center">
              <View className="h-12 w-12 rounded-full bg-yellow-100 items-center justify-center mr-4">
                <TablerIconComponent name="loader" size={24} color="#eab308" />
              </View>
              <View className="flex-1">
                <Text className="font-semibold text-yellow-800">
                  Verifying on Blockchain
                </Text>
                <Text className="text-yellow-700 text-sm">
                  Please wait while we verify this batch's authenticity on the
                  blockchain...
                </Text>
              </View>
            </View>
          )}

          {/* Batch Information Section */}
          {scannedData && verificationStatus === "verified" && (
            <>
              <View className="bg-gradient-to-r from-primary to-primary-dark p-5 rounded-xl mb-6">
                <View className="flex-row justify-between items-start mb-4">
                  <View>
                    <Text className="text-white/80 text-sm">Batch ID</Text>
                    <Text className="text-white font-bold text-xl">
                      {scannedData.batchId}
                    </Text>
                  </View>
                  <View className="bg-white/20 p-2 rounded-lg">
                    <TablerIconComponent
                      name="certificate"
                      size={24}
                      color="white"
                    />
                  </View>
                </View>

                <View className="flex-row flex-wrap mb-4">
                  <View className="w-1/2 mb-3">
                    <Text className="text-white/70 text-xs">Origin</Text>
                    <Text className="text-white">{scannedData.origin}</Text>
                  </View>
                  <View className="w-1/2 mb-3">
                    <Text className="text-white/70 text-xs">Harvest Date</Text>
                    <Text className="text-white">
                      {scannedData.harvestDate}
                    </Text>
                  </View>
                  <View className="w-1/2 mb-3">
                    <Text className="text-white/70 text-xs">Certified On</Text>
                    <Text className="text-white">
                      {scannedData.certificationDate}
                    </Text>
                  </View>
                  <View className="w-1/2 mb-3">
                    <Text className="text-white/70 text-xs">Blockchain</Text>
                    <Text className="text-white">{scannedData.blockchain}</Text>
                  </View>
                </View>

                <TouchableOpacity
                  className="bg-white/20 p-3 rounded-lg items-center flex-row justify-center"
                  onPress={() => viewOnBlockchain()}
                >
                  <TablerIconComponent
                    name="currency-ethereum"
                    size={18}
                    color="white"
                  />
                  <Text className="text-white ml-2 font-medium">
                    View NFT Certificate
                  </Text>
                </TouchableOpacity>
              </View>

              {/* Timeline Section */}
              <Text className="text-lg font-semibold mb-4">
                Blockchain Timeline
              </Text>

              <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm mb-6">
                {timeline.map((item, index) => (
                  <View key={item.id} className="mb-4 last:mb-0">
                    <View className="flex-row">
                      {/* Timeline line */}
                      <View className="items-center mr-4">
                        <View
                          style={{ backgroundColor: item.color }}
                          className="h-10 w-10 rounded-full items-center justify-center z-10"
                        >
                          <TablerIconComponent
                            name={item.icon}
                            size={20}
                            color="white"
                          />
                        </View>
                        {index < timeline.length - 1 && (
                          <View className="h-full w-0.5 bg-gray-200 absolute top-10 bottom-0 left-5" />
                        )}
                      </View>

                      {/* Content */}
                      <View className="flex-1">
                        <Text className="font-semibold">{item.event}</Text>
                        <Text className="text-gray-500 text-xs mb-1">
                          {item.date}
                        </Text>
                        <Text className="text-gray-600 mb-2">
                          {item.details}
                        </Text>
                        <TouchableOpacity
                          className="flex-row items-center"
                          onPress={() => viewOnBlockchain(item.txHash)}
                        >
                          <Text className="text-indigo-600 text-xs">
                            View on blockchain
                          </Text>
                          <TablerIconComponent
                            name="external-link"
                            size={12}
                            color="#4338ca"
                            style={{ marginLeft: 4 }}
                          />
                        </TouchableOpacity>
                      </View>
                    </View>
                  </View>
                ))}
              </View>

              {/* Verification Badge */}
              <View className="bg-green-50 rounded-xl p-4 mb-6 flex-row items-center">
                <View className="h-12 w-12 rounded-full bg-green-100 items-center justify-center mr-4">
                  <TablerIconComponent
                    name="shield-check"
                    size={24}
                    color="#10b981"
                  />
                </View>
                <View className="flex-1">
                  <Text className="font-semibold text-green-800">
                    Blockchain Verified
                  </Text>
                  <Text className="text-green-700 text-sm">
                    This batch has been verified on the blockchain and has a
                    valid NFT certificate
                  </Text>
                </View>
              </View>
            </>
          )}

          {/* Recent Scans Section */}
          <Text className="text-lg font-semibold mt-2 mb-4">Recent Scans</Text>

          <View className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm">
            <View className="flex-row items-center">
              <View className="h-10 w-10 rounded-full bg-blue-100 items-center justify-center mr-3">
                <TablerIconComponent name="qrcode" size={20} color="#3b82f6" />
              </View>
              <View className="flex-1">
                <Text className="font-medium">SH-2023-09-A31</Text>
                <Text className="text-gray-500 text-xs">
                  Scanned today, 2:15 PM
                </Text>
              </View>
              <TouchableOpacity>
                <TablerIconComponent
                  name="chevron-right"
                  size={20}
                  color="#9ca3af"
                />
              </TouchableOpacity>
            </View>
          </View>

          <View className="bg-white border border-gray-200 rounded-xl p-4 shadow-sm">
            <View className="flex-row items-center">
              <View className="h-10 w-10 rounded-full bg-blue-100 items-center justify-center mr-3">
                <TablerIconComponent name="qrcode" size={20} color="#3b82f6" />
              </View>
              <View className="flex-1">
                <Text className="font-medium">SH-2023-08-B17</Text>
                <Text className="text-gray-500 text-xs">
                  Scanned yesterday, 10:30 AM
                </Text>
              </View>
              <TouchableOpacity>
                <TablerIconComponent
                  name="chevron-right"
                  size={20}
                  color="#9ca3af"
                />
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
