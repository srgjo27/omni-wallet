// ── Request / Response types mirroring the user-service API ──

export type KycStatus = "UNVERIFIED" | "PENDING" | "VERIFIED";

export interface User {
  id: string;
  name: string;
  email: string;
  kyc_status: KycStatus;
  created_at: string;
  updated_at: string;
}

/** POST /api/v1/users/register */
export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

/** POST /api/v1/users/login */
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  token_type: string;
  expires_in: number;
  user: User;
}

/** PUT /api/v1/users/pin */
export interface SetPinRequest {
  pin: string;
  confirm_pin: string;
}

/** PUT /api/v1/users/kyc */
export interface UpdateKycRequest {
  id_number: string;
  full_name: string;
}
