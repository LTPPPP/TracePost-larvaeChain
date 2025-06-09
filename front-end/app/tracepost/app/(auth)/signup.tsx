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
  Dimensions,
} from "react-native";
import { Link, useRouter } from "expo-router";
import { LinearGradient } from "expo-linear-gradient";
import TablerIconComponent from "@/components/icon";
import { signup } from "@/api/auth";
import "@/global.css";

const { height } = Dimensions.get("window");

export default function SignupScreen() {
  const [currentStep, setCurrentStep] = useState(1);
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

  const validateStep = (step: number) => {
    const newErrors = { ...errors };
    let isValid = true;

    if (step === 1) {
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
      } else {
        newErrors.username = "";
      }

      // Email validation
      if (!formData.email) {
        newErrors.email = "Email is required";
        isValid = false;
      } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
        newErrors.email = "Email is invalid";
        isValid = false;
      } else {
        newErrors.email = "";
      }

      // Company ID validation
      if (!formData.companyId) {
        newErrors.companyId = "Company ID is required";
        isValid = false;
      } else if (formData.companyId.length < 1) {
        newErrors.companyId = "Company ID must be at least 1 character";
        isValid = false;
      } else {
        newErrors.companyId = "";
      }
    }

    if (step === 2) {
      // Password validation
      if (!formData.password) {
        newErrors.password = "Password is required";
        isValid = false;
      } else if (formData.password.length < 6) {
        newErrors.password = "Password must be at least 6 characters";
        isValid = false;
      } else {
        newErrors.password = "";
      }

      // Confirm Password validation
      if (!formData.confirmPassword) {
        newErrors.confirmPassword = "Please confirm your password";
        isValid = false;
      } else if (formData.password !== formData.confirmPassword) {
        newErrors.confirmPassword = "Passwords do not match";
        isValid = false;
      } else {
        newErrors.confirmPassword = "";
      }
    }

    setErrors(newErrors);
    return isValid;
  };

  const handleNext = () => {
    if (validateStep(currentStep)) {
      setCurrentStep(2);
    }
  };

  const handleBack = () => {
    setCurrentStep(1);
  };

  const handleSignup = async () => {
    if (!validateStep(2)) return;

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
        role: "user",
        username: formData.username,
      };

      const response = await signup(signupData);

      if (response.success) {
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

  const renderStep1 = () => (
    <View className="space-y-6">
      <Text className="text-gray-800 text-2xl font-bold text-center mb-2">
        Create Your Account
      </Text>
      <Text className="text-gray-600 text-center mb-8">
        Join the future of aquaculture management
      </Text>

      {/* Username Field */}
      <View className="mb-6">
        <Text className="text-gray-700 font-semibold mb-3 text-base">
          Username
        </Text>
        <View
          className={`flex-row items-center bg-gray-50 rounded-xl border-2 ${errors.username ? "border-red-300" : "border-gray-200"}`}
        >
          <View className="p-4">
            <TablerIconComponent name="user" size={20} color="#6b7280" />
          </View>
          <TextInput
            className="flex-1 p-4 text-gray-800 text-base"
            placeholder="Choose a username"
            placeholderTextColor="#9ca3af"
            value={formData.username}
            onChangeText={(text) => updateFormData("username", text.trim())}
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

      {/* Email Field */}
      <View className="mb-6">
        <Text className="text-gray-700 font-semibold mb-3 text-base">
          Email Address
        </Text>
        <View
          className={`flex-row items-center bg-gray-50 rounded-xl border-2 ${errors.email ? "border-red-300" : "border-gray-200"}`}
        >
          <View className="p-4">
            <TablerIconComponent name="mail" size={20} color="#6b7280" />
          </View>
          <TextInput
            className="flex-1 p-4 text-gray-800 text-base"
            placeholder="Enter your email"
            placeholderTextColor="#9ca3af"
            value={formData.email}
            onChangeText={(text) => updateFormData("email", text.trim())}
            keyboardType="email-address"
            autoCapitalize="none"
            autoCorrect={false}
            autoComplete="email"
            textContentType="emailAddress"
            editable={!isLoading}
          />
        </View>
        {errors.email ? (
          <Text className="text-red-500 text-sm mt-2 ml-2">{errors.email}</Text>
        ) : null}
      </View>

      {/* Company ID Field */}
      <View className="mb-6">
        <Text className="text-gray-700 font-semibold mb-3 text-base">
          Company ID
        </Text>
        <View
          className={`flex-row items-center bg-gray-50 rounded-xl border-2 ${errors.companyId ? "border-red-300" : "border-gray-200"}`}
        >
          <View className="p-4">
            <TablerIconComponent name="building" size={20} color="#6b7280" />
          </View>
          <TextInput
            className="flex-1 p-4 text-gray-800 text-base"
            placeholder="Enter your company ID"
            placeholderTextColor="#9ca3af"
            value={formData.companyId}
            onChangeText={(text) => updateFormData("companyId", text.trim())}
            autoCapitalize="none"
            autoCorrect={false}
            editable={!isLoading}
          />
        </View>
        {errors.companyId ? (
          <Text className="text-red-500 text-sm mt-2 ml-2">
            {errors.companyId}
          </Text>
        ) : null}
        <Text className="text-gray-500 text-sm mt-2 ml-2">
          Contact your administrator for your company ID
        </Text>
      </View>
    </View>
  );

  const renderStep2 = () => (
    <View className="space-y-6">
      <Text className="text-gray-800 text-2xl font-bold text-center mb-2">
        Secure Your Account
      </Text>
      <Text className="text-gray-600 text-center mb-8">
        Create a strong password to protect your data
      </Text>

      {/* Password Field */}
      <View className="mb-6">
        <Text className="text-gray-700 font-semibold mb-3 text-base">
          Password
        </Text>
        <View
          className={`flex-row items-center bg-gray-50 rounded-xl border-2 ${errors.password ? "border-red-300" : "border-gray-200"}`}
        >
          <View className="p-4">
            <TablerIconComponent name="lock" size={20} color="#6b7280" />
          </View>
          <TextInput
            className="flex-1 p-4 text-gray-800 text-base"
            placeholder="Create a password"
            placeholderTextColor="#9ca3af"
            value={formData.password}
            onChangeText={(text) => updateFormData("password", text)}
            secureTextEntry={!showPassword}
            autoComplete="new-password"
            textContentType="newPassword"
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

      {/* Confirm Password Field */}
      <View className="mb-6">
        <Text className="text-gray-700 font-semibold mb-3 text-base">
          Confirm Password
        </Text>
        <View
          className={`flex-row items-center bg-gray-50 rounded-xl border-2 ${errors.confirmPassword ? "border-red-300" : "border-gray-200"}`}
        >
          <View className="p-4">
            <TablerIconComponent name="lock-check" size={20} color="#6b7280" />
          </View>
          <TextInput
            className="flex-1 p-4 text-gray-800 text-base"
            placeholder="Confirm your password"
            placeholderTextColor="#9ca3af"
            value={formData.confirmPassword}
            onChangeText={(text) => updateFormData("confirmPassword", text)}
            secureTextEntry={!showConfirmPassword}
            autoComplete="new-password"
            textContentType="newPassword"
            editable={!isLoading}
          />
          <TouchableOpacity
            className="p-4"
            onPress={() => setShowConfirmPassword(!showConfirmPassword)}
            disabled={isLoading}
          >
            <TablerIconComponent
              name={showConfirmPassword ? "eye-off" : "eye"}
              size={20}
              color="#6b7280"
            />
          </TouchableOpacity>
        </View>
        {errors.confirmPassword ? (
          <Text className="text-red-500 text-sm mt-2 ml-2">
            {errors.confirmPassword}
          </Text>
        ) : null}
      </View>

      {/* Password Requirements */}
      <View className="bg-blue-50 p-4 rounded-xl">
        <Text className="text-blue-800 font-medium mb-2">
          Password Requirements:
        </Text>
        <View className="space-y-1">
          <View className="flex-row items-center">
            <TablerIconComponent
              name={formData.password.length >= 6 ? "check" : "x"}
              size={16}
              color={formData.password.length >= 6 ? "#059669" : "#dc2626"}
            />
            <Text
              className={`ml-2 text-sm ${formData.password.length >= 6 ? "text-green-600" : "text-red-600"}`}
            >
              At least 6 characters long
            </Text>
          </View>
          <View className="flex-row items-center">
            <TablerIconComponent
              name={
                formData.password === formData.confirmPassword &&
                formData.password
                  ? "check"
                  : "x"
              }
              size={16}
              color={
                formData.password === formData.confirmPassword &&
                formData.password
                  ? "#059669"
                  : "#dc2626"
              }
            />
            <Text
              className={`ml-2 text-sm ${formData.password === formData.confirmPassword && formData.password ? "text-green-600" : "text-red-600"}`}
            >
              Passwords match
            </Text>
          </View>
        </View>
      </View>
    </View>
  );

  return (
    <KeyboardAvoidingView
      style={{ flex: 1 }}
      behavior={Platform.OS === "ios" ? "padding" : "height"}
    >
      <LinearGradient
        colors={["#059669", "#10b981", "#34d399"]}
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
            {/* Back Button */}
            <TouchableOpacity
              className="absolute top-16 left-6 z-10"
              onPress={() => router.back()}
              disabled={isLoading}
            >
              <View className="h-10 w-10 rounded-full bg-white/20 items-center justify-center">
                <TablerIconComponent
                  name="arrow-left"
                  size={20}
                  color="white"
                />
              </View>
            </TouchableOpacity>

            <View className="items-center mb-8 mt-12">
              {/* Logo/Brand Icon */}
              <View className="h-20 w-20 rounded-full bg-white/20 items-center justify-center mb-4">
                <TablerIconComponent name="user-plus" size={40} color="white" />
              </View>

              <Text className="text-white text-3xl font-bold text-center mb-2">
                Join TracePost
              </Text>
              <Text className="text-white/80 text-lg text-center">
                Aquaculture Management Platform
              </Text>
            </View>

            {/* Progress Indicator */}
            <View className="flex-row justify-center mb-8">
              <View
                className={`h-2 w-16 mx-1 rounded-full ${currentStep >= 1 ? "bg-white" : "bg-white/30"}`}
              />
              <View
                className={`h-2 w-16 mx-1 rounded-full ${currentStep >= 2 ? "bg-white" : "bg-white/30"}`}
              />
            </View>

            {/* Signup Card */}
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

              {/* Form Steps */}
              {currentStep === 1 ? renderStep1() : renderStep2()}

              {/* Navigation Buttons */}
              <View className="flex-row gap-3 mt-8">
                {currentStep > 1 && (
                  <TouchableOpacity
                    className="flex-1 bg-gray-100 py-4 rounded-xl items-center"
                    onPress={handleBack}
                    disabled={isLoading}
                  >
                    <View className="flex-row items-center">
                      <TablerIconComponent
                        name="arrow-left"
                        size={20}
                        color="#4b5563"
                      />
                      <Text className="font-bold text-gray-700 ml-2">Back</Text>
                    </View>
                  </TouchableOpacity>
                )}

                <TouchableOpacity
                  className={`flex-1 py-4 rounded-xl items-center ${
                    isLoading ? "bg-green-300" : "bg-green-600"
                  }`}
                  onPress={currentStep === 2 ? handleSignup : handleNext}
                  disabled={isLoading}
                >
                  {isLoading ? (
                    <View className="flex-row items-center">
                      <ActivityIndicator color="white" size="small" />
                      <Text className="font-bold text-white ml-2">
                        Creating Account...
                      </Text>
                    </View>
                  ) : (
                    <View className="flex-row items-center">
                      <Text className="font-bold text-white text-lg">
                        {currentStep === 2 ? "Create Account" : "Next"}
                      </Text>
                      <TablerIconComponent
                        name={currentStep === 2 ? "check" : "arrow-right"}
                        size={20}
                        color="white"
                        style={{ marginLeft: 8 }}
                      />
                    </View>
                  )}
                </TouchableOpacity>
              </View>

              {/* Sign In Link */}
              <View className="flex-row justify-center mt-6">
                <Text className="text-gray-600 text-base">
                  Already have an account?{" "}
                </Text>
                <Link href="/(auth)/login" asChild>
                  <TouchableOpacity disabled={isLoading}>
                    <Text className="text-green-600 font-bold text-base">
                      Sign in
                    </Text>
                  </TouchableOpacity>
                </Link>
              </View>
            </View>

            {/* Social Signup Options */}
            {currentStep === 1 && (
              <View className="mb-8">
                <View className="flex-row items-center mb-6">
                  <View className="flex-1 h-px bg-white/30" />
                  <Text className="mx-4 text-white/80 font-medium">
                    Or sign up with
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
            )}
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
                Your data is protected by enterprise-grade security
              </Text>
            </View>

            <View className="flex-row justify-center space-x-6">
              <View className="items-center">
                <TablerIconComponent
                  name="currency-ethereum"
                  size={20}
                  color="white"
                />
                <Text className="text-white/60 text-xs mt-1">Blockchain</Text>
              </View>
              <View className="items-center">
                <TablerIconComponent name="cloud" size={20} color="white" />
                <Text className="text-white/60 text-xs mt-1">Cloud Sync</Text>
              </View>
              <View className="items-center">
                <TablerIconComponent
                  name="shield-lock"
                  size={20}
                  color="white"
                />
                <Text className="text-white/60 text-xs mt-1">Encrypted</Text>
              </View>
            </View>
          </View>
        </ScrollView>
      </LinearGradient>
    </KeyboardAvoidingView>
  );
}
