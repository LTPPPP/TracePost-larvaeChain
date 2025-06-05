import React, { useState } from "react";
import {
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Text,
  TextInput,
  View,
  TouchableOpacity,
  ActivityIndicator,
} from "react-native";
import { Link, useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import { login } from "@/api/auth";
import { StorageService } from "@/utils/storage";
import "@/global.css";

import { useRole } from "@/contexts/RoleContext";

export default function LoginScreen() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({
    username: "",
    password: "",
    general: "",
  });

  const router = useRouter();
  const { checkRole } = useRole();

  const validateForm = () => {
    let isValid = true;
    const newErrors = { username: "", password: "", general: "" };

    // Username validation
    if (!username) {
      newErrors.username = "Username is required";
      isValid = false;
    } else if (username.length < 3) {
      newErrors.username = "Username must be at least 3 characters";
      isValid = false;
    } else if (username.length > 20) {
      newErrors.username = "Username must be less than 20 characters";
      isValid = false;
    } else if (!/^[a-zA-Z0-9_.-]+$/.test(username)) {
      newErrors.username =
        "Username can only contain letters, numbers, dots, hyphens, and underscores";
      isValid = false;
    }

    // Password validation
    if (!password) {
      newErrors.password = "Password is required";
      isValid = false;
    } else if (password.length < 6) {
      newErrors.password = "Password must be at least 6 characters";
      isValid = false;
    }

    setErrors(newErrors);
    return isValid;
  };

  const handleLogin = async () => {
    if (!validateForm()) return;

    setIsLoading(true);
    setErrors({ username: "", password: "", general: "" });

    try {
      // Call login API
      const response = await login(username, password);

      if (response.success) {
        // Store login data in AsyncStorage
        await StorageService.storeLoginData(response);

        await checkRole();

        // Navigate to home screen
        router.replace("/(tabs)/(home)");
      } else {
        setErrors((prev) => ({
          ...prev,
          general: response.message || "Login failed",
        }));
      }
    } catch (error) {
      console.error("Login error:", error);

      // Handle different types of errors
      if (error instanceof Error) {
        if (error.message.includes("fetch")) {
          setErrors((prev) => ({
            ...prev,
            general: "Network error. Please check your connection.",
          }));
        } else {
          setErrors((prev) => ({
            ...prev,
            general: error.message,
          }));
        }
      } else {
        setErrors((prev) => ({
          ...prev,
          general: "An unexpected error occurred. Please try again.",
        }));
      }
    } finally {
      setIsLoading(false);
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
        <View className="flex-1 px-6 py-10 justify-between items-center bg-white">
          <View className="w-full items-center mt-16 mb-12">
            <Text className="font-bold text-3xl text-gray-800">
              Welcome Back
            </Text>
            <Text className="text-gray-500 mt-2 text-center">
              Sign in to continue to your account
            </Text>
          </View>

          <View className="w-full">
            {/* General Error Message */}
            {errors.general ? (
              <View className="bg-red-50 border border-red-200 rounded-xl p-4 mb-4">
                <Text className="text-red-700 text-center">
                  {errors.general}
                </Text>
              </View>
            ) : null}

            <View className="mb-5">
              <Text className="text-gray-700 font-medium mb-1 ml-1">
                Username
              </Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 ${errors.username ? "border-red-500" : "border-gray-300"}`}
                  placeholder="Enter your username"
                  value={username}
                  onChangeText={(text) => {
                    setUsername(text.trim()); // Remove spaces
                    setErrors((prev) => ({
                      ...prev,
                      username: "",
                      general: "",
                    }));
                  }}
                  autoCapitalize="none"
                  autoCorrect={false}
                  autoComplete="username"
                  textContentType="username"
                  editable={!isLoading}
                />
                {errors.username ? (
                  <Text className="text-red-500 text-xs mt-1 ml-1">
                    {errors.username}
                  </Text>
                ) : null}
              </View>
            </View>

            <View className="mb-2">
              <Text className="text-gray-700 font-medium mb-1 ml-1">
                Password
              </Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 pr-12 ${errors.password ? "border-red-500" : "border-gray-300"}`}
                  placeholder="Enter your password"
                  value={password}
                  onChangeText={(text) => {
                    setPassword(text);
                    setErrors((prev) => ({
                      ...prev,
                      password: "",
                      general: "",
                    }));
                  }}
                  secureTextEntry={!showPassword}
                  autoComplete="current-password"
                  textContentType="password"
                  editable={!isLoading}
                />
                <TouchableOpacity
                  className="absolute right-3 top-4"
                  onPress={() => setShowPassword(!showPassword)}
                  disabled={isLoading}
                >
                  <Ionicons
                    name={showPassword ? "eye-off-outline" : "eye-outline"}
                    size={24}
                    color="gray"
                  />
                </TouchableOpacity>
                {errors.password ? (
                  <Text className="text-red-500 text-xs mt-1 ml-1">
                    {errors.password}
                  </Text>
                ) : null}
              </View>
            </View>

            <View className="items-end mb-6">
              <Link href="/forgot" asChild>
                <TouchableOpacity disabled={isLoading}>
                  <Text className="text-blue-600 font-medium">
                    Forgot password?
                  </Text>
                </TouchableOpacity>
              </Link>
            </View>

            <TouchableOpacity
              className={`rounded-xl py-4 ${isLoading ? "bg-green-200" : "bg-green-500"} items-center`}
              onPress={handleLogin}
              disabled={isLoading}
            >
              {isLoading ? (
                <View className="flex-row items-center">
                  <ActivityIndicator color="white" size="small" />
                  <Text className="font-bold text-white text-lg ml-2">
                    SIGNING IN...
                  </Text>
                </View>
              ) : (
                <Text className="font-bold text-white text-lg">SIGN IN</Text>
              )}
            </TouchableOpacity>

            <View className="flex-row justify-center mt-8">
              <Text className="text-gray-600">
                Don&apos;t have an account?{" "}
              </Text>
              <Link href="/signup" asChild>
                <TouchableOpacity disabled={isLoading}>
                  <Text className="text-blue-600 font-bold">Sign up</Text>
                </TouchableOpacity>
              </Link>
            </View>

            <View className="mt-10">
              <View className="flex-row items-center my-4">
                <View className="flex-1 h-0.5 bg-gray-200" />
                <Text className="mx-4 text-gray-500">Or continue with</Text>
                <View className="flex-1 h-0.5 bg-gray-200" />
              </View>

              <View className="flex-row justify-center gap-4 mt-2">
                <TouchableOpacity
                  className="border border-gray-300 rounded-xl p-3 px-10"
                  disabled={isLoading}
                >
                  <Ionicons name="logo-google" size={24} color="#DB4437" />
                </TouchableOpacity>
                <TouchableOpacity
                  className="border border-gray-300 rounded-xl p-3 px-10"
                  disabled={isLoading}
                >
                  <Ionicons name="logo-apple" size={24} color="#000000" />
                </TouchableOpacity>
              </View>
            </View>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
