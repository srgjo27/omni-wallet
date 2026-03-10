import { apiClient } from "./client";
import type {
  BalanceResponse,
  MutationListResponse,
} from "@/domain/models/wallet.types";
import type { TransactionHistoryResponse } from "@/domain/models/transaction.types";

const WALLET_BASE = "/api/v1/wallets";

export const walletApi = {
  getBalance: () => apiClient.get<BalanceResponse>(`${WALLET_BASE}/balance`),

  getMutations: (page = 1, pageSize = 10) =>
    apiClient.get<MutationListResponse>(
      `${WALLET_BASE}/mutations?page=${page}&page_size=${pageSize}`,
    ),

  getTransactionHistory: (page = 1, pageSize = 10) =>
    apiClient.get<TransactionHistoryResponse>(
      `${WALLET_BASE}/transactions?page=${page}&page_size=${pageSize}`,
    ),

  adminGetWalletStats: () =>
    apiClient.get<{ total_transactions: number; total_volume: number }>(`${WALLET_BASE}/stats`),
};
