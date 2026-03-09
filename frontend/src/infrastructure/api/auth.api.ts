import { apiClient } from "./client";
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  SetPinRequest,
  UpdateKycRequest,
  User,
} from "@/domain/models/auth.types";

const BASE = "/api/v1/users";

export const authApi = {
  /** Registers a new user. Returns the created user entity. */
  register: (body: RegisterRequest) =>
    apiClient.post<User>(`${BASE}/register`, body),

  /** Authenticates a user and returns a JWT token + user profile. */
  login: (body: LoginRequest) =>
    apiClient.post<LoginResponse>(`${BASE}/login`, body),

  /** Returns the authenticated user's profile. */
  getProfile: () => apiClient.get<User>(`${BASE}/profile`),

  /** Sets or updates the user's 6-digit transaction PIN. */
  setPin: (body: SetPinRequest) => apiClient.put<null>(`${BASE}/pin`, body),

  /** Submits KYC details; transitions status to PENDING. */
  updateKyc: (body: UpdateKycRequest) =>
    apiClient.put<null>(`${BASE}/kyc`, body),

  /** Admin: sets a user's KYC status to VERIFIED. */
  adminVerifyKyc: (userId: string) =>
    apiClient.put<null>(`${BASE}/${userId}/kyc/verify`, {}),

  /** Admin: returns total user count and KYC-verified count. */
  adminGetUserStats: () =>
    apiClient.get<{ total_users: number; verified_users: number }>(`${BASE}/stats`),

  /** Invalidates the current session token. */
  logout: () => apiClient.post<null>(`${BASE}/logout`, {}),
};
