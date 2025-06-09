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
  Dimensions,
} from "react-native";
import { Link, useRouter } from "expo-router";
import { LinearGradient } from "expo-linear-gradient";
import TablerIconComponent from "@/components/icon";
import { login } from "@/api/auth";
import { StorageService } from "@/utils/storage";
import "@/global.css";

import { useRole } from "@/contexts/RoleContext";

const { height } = Dimensions.get("window");

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
      <LinearGradient
        colors={["#1e40af", "#3b82f6", "#60a5fa"]}
        className="flex-1"
        start={{ x: 0, y: 0 }}
        end={{ x: 1, y: 1 }}
      >
        <ScrollView
          contentContainerStyle={{ flexGrow: 1 }}
          showsVerticalScrollIndicator={false}
        >
          {/* Top Section with Branding */}
          <View className="flex-1 pt-16 px-6">
            <View className="items-center mb-12">
              {/* Logo/Brand Icon */}
              <View className="h-20 w-20 rounded-full bg-white/20 items-center justify-center mb-4">
                <TablerIconComponent name="fish" size={40} color="white" />
              </View>

              <Text className="text-white text-3xl font-bold text-center mb-2">
                Welcome Back
              </Text>
              <Text className="text-white/80 text-lg text-center">
                TracePost Aquaculture
              </Text>
              <Text className="text-white/60 text-center mt-2">
                Sign in to manage your breeding operations
              </Text>
            </View>

            {/* Login Card */}
            <View className="bg-white rounded-3xl p-6 shadow-lg mb-6">
              {/* General Error Message */}
              {errors.general ? (
                <View className="bg-red-50 border border-red-200 rounded-xl p-4 mb-6">
                  <View className="flex-row items-center">
                    <TablerIconComponent
                      name="alert-circle"
                      size={20}
                      color="#dc2626"
                    />
                    <Text className="text-red-700 ml-2 flex-1">
                      {errors.general}
                    </Text>
                  </View>
                </View>
              ) : null}

              {/* Username Field */}
              <View className="mb-6">
                <Text className="text-gray-700 font-semibold mb-3 text-base">
                  Username
                </Text>
                <View
                  className={`flex-row items-center bg-gray-50 rounded-xl border-2 ${errors.username ? "border-red-300" : "border-gray-200"}`}
                >
                  <View className="p-4">
                    <TablerIconComponent
                      name="user"
                      size={20}
                      color="#6b7280"
                    />
                  </View>
                  <TextInput
                    className="flex-1 p-4 text-gray-800 text-base"
                    placeholder="Enter your username"
                    placeholderTextColor="#9ca3af"
                    value={username}
                    onChangeText={(text) => {
                      setUsername(text.trim());
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
                </View>
                {errors.username ? (
                  <Text className="text-red-500 text-sm mt-2 ml-2">
                    {errors.username}
                  </Text>
                ) : null}
              </View>

              {/* Password Field */}
              <View className="mb-6">
                <Text className="text-gray-700 font-semibold mb-3 text-base">
                  Password
                </Text>
                <View
                  className={`flex-row items-center bg-gray-50 rounded-xl border-2 ${errors.password ? "border-red-300" : "border-gray-200"}`}
                >
                  <View className="p-4">
                    <TablerIconComponent
                      name="lock"
                      size={20}
                      color="#6b7280"
                    />
                  </View>
                  <TextInput
                    className="flex-1 p-4 text-gray-800 text-base"
                    placeholder="Enter your password"
                    placeholderTextColor="#9ca3af"
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
                    className="p-4"
                    onPress={() => setShowPassword(!showPassword)}
                    disabled={isLoading}
                  >
                    <TablerIconComponent
                      name={showPassword ? "eye-off" : "eye"}
                      size={20}
                      color="#6b7280"
                    />
                  </TouchableOpacity>
                </View>
                {errors.password ? (
                  <Text className="text-red-500 text-sm mt-2 ml-2">
                    {errors.password}
                  </Text>
                ) : null}
              </View>

              {/* Forgot Password Link */}
              <View className="items-end mb-8">
                <Link href="/forgot" asChild>
                  <TouchableOpacity disabled={isLoading}>
                    <Text className="text-blue-600 font-medium text-base">
                      Forgot password?
                    </Text>
                  </TouchableOpacity>
                </Link>
              </View>

              {/* Login Button */}
              <TouchableOpacity
                className={`rounded-xl py-4 items-center mb-6 ${
                  isLoading ? "bg-blue-300" : "bg-blue-600"
                }`}
                onPress={handleLogin}
                disabled={isLoading}
              >
                {isLoading ? (
                  <View className="flex-row items-center">
                    <ActivityIndicator color="white" size="small" />
                    <Text className="font-bold text-white text-lg ml-3">
                      Signing In...
                    </Text>
                  </View>
                ) : (
                  <View className="flex-row items-center">
                    <TablerIconComponent name="login" size={20} color="white" />
                    <Text className="font-bold text-white text-lg ml-3">
                      Sign In
                    </Text>
                  </View>
                )}
              </TouchableOpacity>

              {/* Sign Up Link */}
              <View className="flex-row justify-center">
                <Text className="text-gray-600 text-base">
                  Don't have an account?{" "}
                </Text>
                <Link href="/signup" asChild>
                  <TouchableOpacity disabled={isLoading}>
                    <Text className="text-blue-600 font-bold text-base">
                      Sign up
                    </Text>
                  </TouchableOpacity>
                </Link>
              </View>
            </View>

            {/* Social Login Options */}
            <View className="mb-8">
              <View className="flex-row items-center mb-6">
                <View className="flex-1 h-px bg-white/30" />
                <Text className="mx-4 text-white/80 font-medium">
                  Or continue with
                </Text>
                <View className="flex-1 h-px bg-white/30" />
              </View>

              <View className="flex-row justify-center gap-4">
                <TouchableOpacity
                  className="bg-white/20 rounded-xl p-4 flex-1 items-center flex-row justify-center"
                  disabled={isLoading}
                >
                  <TablerIconComponent
                    name="currency-google"
                    size={24}
                    color="white"
                  />
                  <Text className="text-white font-medium ml-2">Google</Text>
                </TouchableOpacity>

                <TouchableOpacity
                  className="bg-white/20 rounded-xl p-4 flex-1 items-center flex-row justify-center"
                  disabled={isLoading}
                >
                  <TablerIconComponent
                    name="currency-apple"
                    size={24}
                    color="white"
                  />
                  <Text className="text-white font-medium ml-2">Apple</Text>
                </TouchableOpacity>
              </View>
            </View>
          </View>

          {/* Bottom Features */}
          <View className="px-6 pb-8">
            <View className="flex-row items-center justify-center mb-4">
              <TablerIconComponent
                name="shield-check"
                size={16}
                color="white"
              />
              <Text className="text-white/80 text-sm ml-2">
                Secured by blockchain technology
              </Text>
            </View>

            <View className="flex-row justify-center space-x-6">
              <View className="items-center">
                <TablerIconComponent
                  name="currency-ethereum"
                  size={20}
                  color="white"
                />
                <Text className="text-white/60 text-xs mt-1">Web3 Ready</Text>
              </View>
              <View className="items-center">
                <TablerIconComponent name="fish" size={20} color="white" />
                <Text className="text-white/60 text-xs mt-1">Aquaculture</Text>
              </View>
              <View className="items-center">
                <TablerIconComponent
                  name="chart-line"
                  size={20}
                  color="white"
                />
                <Text className="text-white/60 text-xs mt-1">Analytics</Text>
              </View>
            </View>
          </View>
        </ScrollView>
      </LinearGradient>
    </KeyboardAvoidingView>
  );
}
