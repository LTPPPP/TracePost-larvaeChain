import React, { useState } from "react";
import {
  ScrollView,
  Text,
  View,
  TouchableOpacity,
  Image,
  TextInput,
  ActivityIndicator,
  Alert,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import TablerIconComponent from "@/components/icon";
import * as ImagePicker from "expo-image-picker";
import "@/global.css";

export default function ReportScreen() {
  const [activeTab, setActiveTab] = useState("feeding");
  const [images, setImages] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSigning, setIsSigning] = useState(false);
  const [formData, setFormData] = useState({
    pond: "",
    feedType: "",
    amount: "",
    notes: "",
    diseaseType: "",
    severity: "",
    treatmentApplied: "",
    harvestAmount: "",
    harvestQuality: "",
  });

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

  const signWithWallet = () => {
    setIsSigning(true);

    // Simulate wallet connection and signing
    setTimeout(() => {
      setIsSigning(false);
      handleSubmit();
    }, 2000);
  };

  const handleSubmit = () => {
    setIsSubmitting(true);

    // Simulate API call and blockchain transaction
    setTimeout(() => {
      setIsSubmitting(false);

      // Reset form
      setFormData({
        pond: "",
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

      // Show success message with transaction hash
      Alert.alert(
        "Report Submitted",
        "Your report has been recorded on the blockchain. Transaction Hash: 0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
        [
          {
            text: "View on Etherscan",
            onPress: () => {
              // Open Etherscan in browser
            },
          },
          { text: "OK" },
        ],
      );
    }, 3000);
  };

  const isFormValid = () => {
    if (activeTab === "feeding") {
      return formData.pond && formData.feedType && formData.amount;
    } else if (activeTab === "disease") {
      return formData.pond && formData.diseaseType && formData.severity;
    } else if (activeTab === "harvest") {
      return formData.pond && formData.harvestAmount;
    }
    return false;
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
                Report Events
              </Text>
              <Text className="text-gray-500">
                Record farm activities & issues
              </Text>
            </View>
            <TouchableOpacity className="h-10 w-10 rounded-full bg-secondary/10 items-center justify-center">
              <TablerIconComponent name="history" size={20} color="#4338ca" />
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
          <View className="flex-row bg-gray-100 rounded-xl p-1 mb-6">
            <TouchableOpacity
              className={`flex-1 py-2 rounded-lg ${activeTab === "feeding" ? "bg-white shadow" : ""}`}
              onPress={() => setActiveTab("feeding")}
            >
              <Text
                className={`text-center font-medium ${activeTab === "feeding" ? "text-primary" : "text-gray-500"}`}
              >
                Feeding
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              className={`flex-1 py-2 rounded-lg ${activeTab === "disease" ? "bg-white shadow" : ""}`}
              onPress={() => setActiveTab("disease")}
            >
              <Text
                className={`text-center font-medium ${activeTab === "disease" ? "text-primary" : "text-gray-500"}`}
              >
                Disease
              </Text>
            </TouchableOpacity>
            <TouchableOpacity
              className={`flex-1 py-2 rounded-lg ${activeTab === "harvest" ? "bg-white shadow" : ""}`}
              onPress={() => setActiveTab("harvest")}
            >
              <Text
                className={`text-center font-medium ${activeTab === "harvest" ? "text-primary" : "text-gray-500"}`}
              >
                Harvest
              </Text>
            </TouchableOpacity>
          </View>

          {/* Form Section */}
          <View className="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
            {activeTab === "feeding" && (
              <>
                <Text className="text-lg font-semibold mb-4">
                  Feeding Report
                </Text>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Pond ID
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="fish"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter pond ID"
                      value={formData.pond}
                      onChangeText={(text) =>
                        setFormData({ ...formData, pond: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Feed Type
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
                    Amount (kg)
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
                    Pond ID
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="fish"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter pond ID"
                      value={formData.pond}
                      onChangeText={(text) =>
                        setFormData({ ...formData, pond: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Disease Type
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
                    Severity
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
                    Pond ID
                  </Text>
                  <View className="flex-row border border-gray-300 rounded-lg overflow-hidden">
                    <View className="bg-gray-100 p-3">
                      <TablerIconComponent
                        name="fish"
                        size={20}
                        color="#4b5563"
                      />
                    </View>
                    <TextInput
                      className="flex-1 p-3 bg-white"
                      placeholder="Enter pond ID"
                      value={formData.pond}
                      onChangeText={(text) =>
                        setFormData({ ...formData, pond: text })
                      }
                    />
                  </View>
                </View>

                <View className="mb-4">
                  <Text className="font-medium text-gray-700 mb-1">
                    Harvest Amount (kg)
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
          <Text className="text-lg font-semibold mt-8 mb-4">
            Recent Reports
          </Text>

          <View className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm">
            <View className="flex-row items-center mb-2">
              <View className="h-10 w-10 rounded-full bg-primary-light/30 items-center justify-center mr-3">
                <TablerIconComponent name="bucket" size={20} color="#f97316" />
              </View>
              <View>
                <Text className="font-medium">Feeding - Pond A1</Text>
                <Text className="text-gray-500 text-xs">Today, 10:30 AM</Text>
              </View>
              <View className="ml-auto bg-gray-100 px-2 py-1 rounded">
                <Text className="text-xs text-gray-600">12kg</Text>
              </View>
            </View>
            <View className="flex-row items-center mt-2 pt-2 border-t border-gray-100">
              <TablerIconComponent
                name="currency-ethereum"
                size={12}
                color="#9ca3af"
              />
              <Text className="text-xs text-gray-500 ml-1">0x3a4e...a581</Text>
            </View>
          </View>

          <View className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm">
            <View className="flex-row items-center mb-2">
              <View className="h-10 w-10 rounded-full bg-red-100 items-center justify-center mr-3">
                <TablerIconComponent name="virus" size={20} color="#ef4444" />
              </View>
              <View>
                <Text className="font-medium">Disease Report - Pond B2</Text>
                <Text className="text-gray-500 text-xs">
                  Yesterday, 4:15 PM
                </Text>
              </View>
              <View className="ml-auto bg-red-100 px-2 py-1 rounded">
                <Text className="text-xs text-red-600">Medium</Text>
              </View>
            </View>
            <View className="flex-row items-center mt-2 pt-2 border-t border-gray-100">
              <TablerIconComponent
                name="currency-ethereum"
                size={12}
                color="#9ca3af"
              />
              <Text className="text-xs text-gray-500 ml-1">0x7b91b...e8f</Text>
            </View>
          </View>

          <View className="bg-white border border-gray-200 rounded-xl p-4 mb-4 shadow-sm">
            <View className="flex-row items-center mb-2">
              <View className="h-10 w-10 rounded-full bg-green-100 items-center justify-center mr-3">
                <TablerIconComponent name="scale" size={20} color="#10b981" />
              </View>
              <View>
                <Text className="font-medium">Harvest - Pond C3</Text>
                <Text className="text-gray-500 text-xs">Oct 19, 2023</Text>
              </View>
              <View className="ml-auto bg-green-100 px-2 py-1 rounded">
                <Text className="text-xs text-green-600">450kg</Text>
              </View>
            </View>
            <View className="flex-row items-center mt-2 pt-2 border-t border-gray-100">
              <TablerIconComponent
                name="currency-ethereum"
                size={12}
                color="#9ca3af"
              />
              <Text className="text-xs text-gray-500 ml-1">0x5e4d3...a4</Text>
            </View>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
