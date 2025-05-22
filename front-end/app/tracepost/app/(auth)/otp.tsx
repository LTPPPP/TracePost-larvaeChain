import React, { useState, useRef, useEffect } from "react";
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
import { useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import "@/global.css";

export default function OTPScreen() {
  const [otp, setOtp] = useState(["", "", "", ""]);
  const [timer, setTimer] = useState(60);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState("");

  const router = useRouter();
  const inputRefs = useRef<(TextInput | null)[]>([]);

  useEffect(() => {
    const interval = setInterval(() => {
      setTimer((prevTimer) => (prevTimer > 0 ? prevTimer - 1 : 0));
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const handleOtpChange = (text: string, index: number) => {
    const newOtp = [...otp];
    newOtp[index] = text;
    setOtp(newOtp);

    // Move to next input if current input is filled
    if (text && index < 3) {
      inputRefs.current[index + 1]?.focus();
    }
  };

  const handleKeyPress = (e: { nativeEvent: { key: string } }, index: number) => {
    // Move to previous input on backspace if current input is empty
    if (e.nativeEvent.key === 'Backspace' && !otp[index] && index > 0) {
      inputRefs.current[index - 1]?.focus();
    }
  };

  const handleVerify = () => {
    const otpValue = otp.join("");

    if (otpValue.length !== 4) {
      setError("Please enter the complete 4-digit code");
      return;
    }

    setIsLoading(true);
    setError("");

    // Simulate API verification
    setTimeout(() => {
      setIsLoading(false);
      router.push("/(tabs)/(home)");
    }, 1500);
  };

  const handleResendCode = () => {
    if (timer === 0) {
      // Simulate resending code
      setTimer(60);
      // Show success message or toast here
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
              Verification
            </Text>
            <Text className="text-gray-500 mt-2">
              Enter the 4-digit code sent to your email
            </Text>
          </View>

          <View>
            <View className="flex-row justify-between mb-8">
              {[0, 1, 2, 3].map((index) => (
                <TextInput
                  key={index}
                  ref={(ref) => {
                    if (ref !== null) {
                      inputRefs.current[index] = ref;
                    }
                  }}
                  className="border border-gray-300 rounded-xl h-16 w-16 text-center text-xl font-bold"
                  maxLength={1}
                  keyboardType="number-pad"
                  value={otp[index]}
                  onChangeText={(text) => handleOtpChange(text, index)}
                  onKeyPress={(e) => handleKeyPress(e, index)}
                />
              ))}
            </View>

            {error ? (
              <Text className="text-red-500 text-center mb-4">{error}</Text>
            ) : null}

            <TouchableOpacity
              className={`rounded-xl py-4 ${isLoading ? "bg-blue-300" : "bg-blue-600"
                } items-center mb-6`}
              onPress={handleVerify}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="white" />
              ) : (
                <Text className="font-bold text-white text-lg">VERIFY</Text>
              )}
            </TouchableOpacity>

            <View className="flex-row justify-center items-center mt-4">
              <Text className="text-gray-600">Didn&apos;t receive the code? </Text>
              <TouchableOpacity
                onPress={handleResendCode}
                disabled={timer > 0}
              >
                <Text
                  className={`font-bold ${timer > 0 ? "text-gray-400" : "text-blue-600"
                    }`}
                >
                  {timer > 0 ? `Resend in ${timer}s` : "Resend"}
                </Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}
