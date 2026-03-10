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
  register: (body: RegisterRequest) =>
    apiClient.post<User>(`${BASE}/register`, body),

  login: (body: LoginRequest) =>
    apiClient.post<LoginResponse>(`${BASE}/login`, body),

  getProfile: () => apiClient.get<User>(`${BASE}/profile`),

  setPin: (body: SetPinRequest) => apiClient.put<null>(`${BASE}/pin`, body),

  updateKyc: (body: UpdateKycRequest) =>
    apiClient.put<null>(`${BASE}/kyc`, body),

  adminVerifyKyc: (userId: string) =>
    apiClient.put<null>(`${BASE}/${userId}/kyc/verify`, {}),

  adminGetUserStats: () =>
    apiClient.get<{ total_users: number; verified_users: number }>(`${BASE}/stats`),

  logout: () => apiClient.post<null>(`${BASE}/logout`, {}),
};
