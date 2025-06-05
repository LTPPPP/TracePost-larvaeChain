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
  Alert,
} from "react-native";
import { Link, useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import { signup } from "@/api/auth";
import "@/global.css";

export default function SignupScreen() {
  const [formData, setFormData] = useState({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
    companyId: "",
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
    companyId: "",
    general: "",
  });

  const router = useRouter();

  const updateFormData = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear related errors when user starts typing
    setErrors((prev) => ({ ...prev, [field]: "", general: "" }));
  };

  const validateForm = () => {
    let isValid = true;
    const newErrors = {
      username: "",
      email: "",
      password: "",
      confirmPassword: "",
      companyId: "",
      general: "",
    };

    // Username validation
    if (!formData.username) {
      newErrors.username = "Username is required";
      isValid = false;
    } else if (formData.username.length < 3) {
      newErrors.username = "Username must be at least 3 characters";
      isValid = false;
    } else if (formData.username.length > 20) {
      newErrors.username = "Username must be less than 20 characters";
      isValid = false;
    } else if (!/^[a-zA-Z0-9_.-]+$/.test(formData.username)) {
      newErrors.username =
        "Username can only contain letters, numbers, dots, hyphens, and underscores";
      isValid = false;
    }

    // Email validation
    if (!formData.email) {
      newErrors.email = "Email is required";
      isValid = false;
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = "Email is invalid";
      isValid = false;
    }

    // Company ID validation
    if (!formData.companyId) {
      newErrors.companyId = "Company ID is required";
      isValid = false;
    } else if (formData.companyId.length < 1) {
      newErrors.companyId = "Company ID must be at least 1 character";
      isValid = false;
    }

    // Password validation
    if (!formData.password) {
      newErrors.password = "Password is required";
      isValid = false;
    } else if (formData.password.length < 6) {
      newErrors.password = "Password must be at least 6 characters";
      isValid = false;
    }

    // Confirm Password validation
    if (!formData.confirmPassword) {
      newErrors.confirmPassword = "Please confirm your password";
      isValid = false;
    } else if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = "Passwords do not match";
      isValid = false;
    }

    setErrors(newErrors);
    return isValid;
  };

  const handleSignup = async () => {
    if (!validateForm()) return;

    setIsLoading(true);
    setErrors({
      username: "",
      email: "",
      password: "",
      confirmPassword: "",
      companyId: "",
      general: "",
    });

    try {
      const signupData = {
        company_id: formData.companyId,
        email: formData.email,
        password: formData.password,
        role: "user", // Hidden from user, set to "user" by default
        username: formData.username,
      };

      const response = await signup(signupData);

      if (response.success) {
        // Show success message
        Alert.alert(
          "Registration Successful",
          "Your account has been created successfully. Please verify your email.",
          [
            {
              text: "OK",
              onPress: () => router.push("/(auth)/otp"),
            },
          ],
        );
      } else {
        setErrors((prev) => ({
          ...prev,
          general: response.message || "Registration failed",
        }));
      }
    } catch (error) {
      console.error("Signup error:", error);

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
        <View className="flex-1 px-6 py-10 bg-white">
          <TouchableOpacity
            className="mt-10"
            onPress={() => router.back()}
            disabled={isLoading}
          >
            <Ionicons name="arrow-back" size={24} color="black" />
          </TouchableOpacity>

          <View className="w-full items-center mt-8 mb-10">
            <Text className="font-bold text-3xl text-gray-800">
              Create Account
            </Text>
            <Text className="text-gray-500 mt-2 text-center">
              Sign up to get started
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

            {/* Username Field */}
            <View className="mb-5">
              <Text className="text-gray-700 font-medium mb-1 ml-1">
                Username
              </Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 ${errors.username ? "border-red-500" : "border-gray-300"}`}
                  placeholder="Enter your username"
                  value={formData.username}
                  onChangeText={(text) =>
                    updateFormData("username", text.trim())
                  }
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

            {/* Email Field */}
            <View className="mb-5">
              <Text className="text-gray-700 font-medium mb-1 ml-1">Email</Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 ${errors.email ? "border-red-500" : "border-gray-300"}`}
                  placeholder="Enter your email"
                  value={formData.email}
                  onChangeText={(text) => updateFormData("email", text.trim())}
                  keyboardType="email-address"
                  autoCapitalize="none"
                  autoCorrect={false}
                  autoComplete="email"
                  textContentType="emailAddress"
                  editable={!isLoading}
                />
                {errors.email ? (
                  <Text className="text-red-500 text-xs mt-1 ml-1">
                    {errors.email}
                  </Text>
                ) : null}
              </View>
            </View>

            {/* Company ID Field */}
            <View className="mb-5">
              <Text className="text-gray-700 font-medium mb-1 ml-1">
                Company ID
              </Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 ${errors.companyId ? "border-red-500" : "border-gray-300"}`}
                  placeholder="Enter your company ID"
                  value={formData.companyId}
                  onChangeText={(text) =>
                    updateFormData("companyId", text.trim())
                  }
                  autoCapitalize="none"
                  autoCorrect={false}
                  editable={!isLoading}
                />
                {errors.companyId ? (
                  <Text className="text-red-500 text-xs mt-1 ml-1">
                    {errors.companyId}
                  </Text>
                ) : null}
              </View>
            </View>

            {/* Password Field */}
            <View className="mb-5">
              <Text className="text-gray-700 font-medium mb-1 ml-1">
                Password
              </Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 pr-12 ${errors.password ? "border-red-500" : "border-gray-300"}`}
                  placeholder="Create a password"
                  value={formData.password}
                  onChangeText={(text) => updateFormData("password", text)}
                  secureTextEntry={!showPassword}
                  autoComplete="new-password"
                  textContentType="newPassword"
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

            {/* Confirm Password Field */}
            <View className="mb-8">
              <Text className="text-gray-700 font-medium mb-1 ml-1">
                Confirm Password
              </Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 pr-12 ${errors.confirmPassword ? "border-red-500" : "border-gray-300"}`}
                  placeholder="Confirm your password"
                  value={formData.confirmPassword}
                  onChangeText={(text) =>
                    updateFormData("confirmPassword", text)
                  }
                  secureTextEntry={!showConfirmPassword}
                  autoComplete="new-password"
                  textContentType="newPassword"
                  editable={!isLoading}
                />
                <TouchableOpacity
                  className="absolute right-3 top-4"
                  onPress={() => setShowConfirmPassword(!showConfirmPassword)}
                  disabled={isLoading}
                >
                  <Ionicons
                    name={
                      showConfirmPassword ? "eye-off-outline" : "eye-outline"
                    }
                    size={24}
                    color="gray"
                  />
                </TouchableOpacity>
                {errors.confirmPassword ? (
                  <Text className="text-red-500 text-xs mt-1 ml-1">
                    {errors.confirmPassword}
                  </Text>
                ) : null}
              </View>
            </View>

            <TouchableOpacity
              className={`rounded-xl py-4 ${isLoading ? "bg-green-200" : "bg-green-500"} items-center mb-6`}
              onPress={handleSignup}
              disabled={isLoading}
            >
              {isLoading ? (
                <View className="flex-row items-center">
                  <ActivityIndicator color="white" size="small" />
                  <Text className="font-bold text-white text-lg ml-2">
                    CREATING ACCOUNT...
                  </Text>
                </View>
              ) : (
                <Text className="font-bold text-white text-lg">SIGN UP</Text>
              )}
            </TouchableOpacity>

            <View className="flex-row justify-center mt-6">
              <Text className="text-gray-600">Already have an account? </Text>
              <Link href="/(auth)/login" asChild>
                <TouchableOpacity disabled={isLoading}>
                  <Text className="text-blue-600 font-bold">Sign in</Text>
                </TouchableOpacity>
              </Link>
            </View>

            <View className="mt-8">
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
