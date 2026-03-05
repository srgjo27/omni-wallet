import { apiClient } from "./client";
import type {
  BalanceResponse,
  MutationListResponse,
} from "@/domain/models/wallet.types";
import type { TransactionHistoryResponse } from "@/domain/models/transaction.types";

const WALLET_BASE = "/api/v1/wallets";

export const walletApi = {
  /** Returns the authenticated user's current wallet balance. */
  getBalance: () => apiClient.get<BalanceResponse>(`${WALLET_BASE}/balance`),

  /** Returns a paginated list of wallet mutations (ledger entries). */
  getMutations: (page = 1, pageSize = 10) =>
    apiClient.get<MutationListResponse>(
      `${WALLET_BASE}/mutations?page=${page}&page_size=${pageSize}`,
    ),

  /** Returns a paginated list of transactions involving this wallet. */
  getTransactionHistory: (page = 1, pageSize = 10) =>
    apiClient.get<TransactionHistoryResponse>(
      `${WALLET_BASE}/transactions?page=${page}&page_size=${pageSize}`,
    ),
};
