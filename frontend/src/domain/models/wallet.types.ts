// ── Types mirroring the wallet-service wallet domain ──

export type WalletStatus = "ACTIVE" | "INACTIVE" | "FROZEN";

export interface Wallet {
  id: string;
  user_id: string;
  balance: number;
  status: WalletStatus;
  created_at: string;
  updated_at: string;
}

export interface BalanceResponse {
  wallet_id: string;
  balance: number;
  status: WalletStatus;
}

export type MutationDirection = "CREDIT" | "DEBIT";

export interface WalletMutation {
  id: string;
  wallet_id: string;
  transaction_id: string;
  direction: MutationDirection;
  amount: number;
  balance_after: number;
  description: string;
  created_at: string;
}

export interface MutationListResponse {
  mutations: WalletMutation[];
  total: number;
  page: number;
  page_size: number;
}
