"use client";

import useSWR from "swr";
import { walletApi } from "@/infrastructure/api/wallet.api";
import { useAuthStore } from "@/store/auth.store";
import type { BalanceResponse, MutationListResponse } from "@/domain/models/wallet.types";
import type { TransactionHistoryResponse } from "@/domain/models/transaction.types";

/**
 * Custom hook for wallet data fetching.
 * SWR handles caching, revalidation, and loading states automatically.
 * Conditional fetching (null key) prevents requests when not authenticated —
 * this avoids 401s that happen when SWR fires before the token is ready.
 */
export function useWallet() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const {
    data: balanceData,
    isLoading: isBalanceLoading,
    error: balanceError,
    mutate: refreshBalance,
  } = useSWR<{ data?: BalanceResponse }>(
    // null key = SWR skips the fetch until user is authenticated
    isAuthenticated ? "wallet/balance" : null,
    () => walletApi.getBalance(),
  );

  return {
    balance: balanceData?.data ?? null,
    isBalanceLoading,
    balanceError,
    refreshBalance,
  };
}

export function useMutations(page = 1, pageSize = 10) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const { data, isLoading, error, mutate } = useSWR<{ data?: MutationListResponse }>(
    isAuthenticated ? `wallet/mutations?page=${page}&page_size=${pageSize}` : null,
    () => walletApi.getMutations(page, pageSize),
  );

  return {
    mutations: data?.data?.mutations ?? [],
    total: data?.data?.total ?? 0,
    isLoading,
    error,
    refresh: mutate,
  };
}

export function useTransactionHistory(page = 1, pageSize = 10) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const { data, isLoading, error, mutate } = useSWR<{ data?: TransactionHistoryResponse }>(
    isAuthenticated ? `wallet/transactions?page=${page}&page_size=${pageSize}` : null,
    () => walletApi.getTransactionHistory(page, pageSize),
  );

  return {
    transactions: data?.data?.transactions ?? [],
    total: data?.data?.total ?? 0,
    isLoading,
    error,
    refresh: mutate,
  };
}
