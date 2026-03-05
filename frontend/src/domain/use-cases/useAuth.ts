"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { authApi } from "@/infrastructure/api/auth.api";
import { useAuthStore } from "@/store/auth.store";
import { ApiError } from "@/infrastructure/api/client";
import type { LoginRequest, RegisterRequest, SetPinRequest } from "@/domain/models/auth.types";

/**
 * Custom hook encapsulating authentication logic.
 * Components stay thin — they only call actions and read state.
 */
export function useAuth() {
  const { user, isAuthenticated, setAuth, clearAuth } = useAuthStore();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const clearError = () => setError(null);

  const login = async (credentials: LoginRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await authApi.login(credentials);
      if (res.data) {
        setAuth(res.data.access_token, res.data.user);
        // Route based on role — admin email convention used as a simple guard
        if (res.data.user.email.endsWith("@admin.omniwallet")) {
          router.push("/admin");
        } else {
          router.push("/dashboard");
        }
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Login failed");
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (data: RegisterRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      await authApi.register(data);
      router.push("/login?registered=1");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Registration failed");
    } finally {
      setIsLoading(false);
    }
  };

  const setPin = async (data: SetPinRequest) => {
    setIsLoading(true);
    setError(null);
    try {
      await authApi.setPin(data);
      return true;
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to set PIN");
      return false;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    try {
      await authApi.logout();
    } finally {
      clearAuth();
      router.push("/login");
    }
  };

  return { user, isAuthenticated, isLoading, error, login, register, setPin, logout, clearError };
}
