import React, { useState } from "react";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  ActivityIndicator,
} from "react-native";
import { Link, useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import "@/global.css";

export default function ForgotPasswordScreen() {
  const [email, setEmail] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const router = useRouter();

  const validateEmail = () => {
    if (!email) {
      setError("Email is required");
      return false;
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      setError("Email is invalid");
      return false;
    }
    setError("");
    return true;
  };

  const handleSendCode = () => {
    if (validateEmail()) {
      setIsLoading(true);

      // Simulate API call to send password reset code
      setTimeout(() => {
        setIsLoading(false);
        setSuccess(true);

        // Wait a moment then redirect to OTP page
        setTimeout(() => {
          router.push("/login");
        }, 1500);
      }, 1500);
    }
  };

  return (
    <KeyboardAvoidingView
      style={{ flex: 1 }}
      behavior={Platform.OS === "ios" ? "padding" : "height"}
    >
      <ScrollView
        contentContainerStyle={{ flexGrow: 1 }}
        showsVerticalScrollIndicator={false}
      >
        <View className="flex-1 px-6 py-10 bg-white">
          <TouchableOpacity
            className="mt-10"
            onPress={() => router.back()}
          >
            <Ionicons name="arrow-back" size={24} color="black" />
          </TouchableOpacity>

          <View className="mt-10 mb-12">
            <Text className="font-bold text-3xl text-gray-800">
              Forgot Password
            </Text>
            <Text className="text-gray-500 mt-2">
              Enter your email address to receive a verification code
            </Text>
          </View>

          <View>
            <View className="mb-6">
              <Text className="text-gray-700 font-medium mb-1 ml-1">Email</Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 ${error ? "border-red-500" : "border-gray-300"
                    }`}
                  placeholder="Enter your email"
                  value={email}
                  onChangeText={(text) => {
                    setEmail(text);
                    setError("");
                    setSuccess(false);
                  }}
                  keyboardType="email-address"
                  autoCapitalize="none"
                  editable={!isLoading && !success}
                />
                {error ? (
                  <Text className="text-red-500 text-xs mt-1 ml-1">
                    {error}
                  </Text>
                ) : null}
              </View>
            </View>

            {success ? (
              <View className="bg-green-100 p-4 rounded-xl mb-6">
                <Text className="text-green-700">
                  A verification code has been sent to your email. Redirecting...
                </Text>
              </View>
            ) : null}

            <TouchableOpacity
              className={`rounded-xl py-4 ${isLoading || success ? "bg-blue-300" : "bg-blue-600"
                } items-center mb-6`}
              onPress={handleSendCode}
              disabled={isLoading || success}
            >
              {isLoading ? (
                <ActivityIndicator color="white" />
              ) : (
                <Text className="font-bold text-white text-lg">
                  {success ? "CODE SENT" : "SEND CODE"}
                </Text>
              )}
            </TouchableOpacity>

            <View className="flex-row justify-center mt-6">
              <Text className="text-gray-600">Remember your password? </Text>
              <Link href="/(auth)/login" asChild>
                <TouchableOpacity>
                  <Text className="text-blue-600 font-bold">Sign in</Text>
                </TouchableOpacity>
              </Link>
            </View>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
