import React, { useState } from "react";
import {
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Text,
  TextInput,
  View,
  TouchableOpacity,
  ActivityIndicator
} from "react-native";
import { Link, useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import "@/global.css";

export default function LoginScreen() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState({ email: "", password: "" });

  const router = useRouter();

  const validateForm = () => {
    let isValid = true;
    const newErrors = { email: "", password: "" };

    // Email validation
    if (!email) {
      newErrors.email = "Email is required";
      isValid = false;
    } else if (!/\S+@\S+\.\S+/.test(email)) {
      newErrors.email = "Email is invalid";
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

  const handleLogin = () => {
    if (validateForm()) {
      setIsLoading(true);

      // Simulate API call
      setTimeout(() => {
        setIsLoading(false);
        router.push("/(tabs)/(home)");
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
            <View className="mb-5">
              <Text className="text-gray-700 font-medium mb-1 ml-1">Email</Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 ${errors.email ? "border-red-500" : "border-gray-300"
                    }`}
                  placeholder="Enter your email"
                  value={email}
                  onChangeText={setEmail}
                  keyboardType="email-address"
                  autoCapitalize="none"
                />
                {errors.email ? (
                  <Text className="text-red-500 text-xs mt-1 ml-1">
                    {errors.email}
                  </Text>
                ) : null}
              </View>
            </View>

            <View className="mb-2">
              <Text className="text-gray-700 font-medium mb-1 ml-1">Password</Text>
              <View className="relative">
                <TextInput
                  className={`border rounded-xl p-4 w-full bg-gray-50 ${errors.password ? "border-red-500" : "border-gray-300"
                    }`}
                  placeholder="Enter your password"
                  value={password}
                  onChangeText={setPassword}
                  secureTextEntry={!showPassword}
                />
                <TouchableOpacity
                  className="absolute right-3 top-4"
                  onPress={() => setShowPassword(!showPassword)}
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
                <TouchableOpacity>
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
                <ActivityIndicator color="white" />
              ) : (
                <Text className="font-bold text-white text-lg">SIGN IN</Text>
              )}
            </TouchableOpacity>

            <View className="flex-row justify-center mt-8">
              <Text className="text-gray-600">Don&apos;t have an account? </Text>
              <Link href="/signup" asChild>
                <TouchableOpacity>
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
                <TouchableOpacity className="border border-gray-300 rounded-xl p-3 px-10">
                  <Ionicons name="logo-google" size={24} color="#DB4437" />
                </TouchableOpacity>
                <TouchableOpacity className="border border-gray-300 rounded-xl p-3 px-10">
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
