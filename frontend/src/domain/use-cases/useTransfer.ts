"use client";

import { useState } from "react";
import { v4 as uuidv4 } from "uuid";
import { transferApi } from "@/infrastructure/api/transfer.api";
import { ApiError } from "@/infrastructure/api/client";
import type { Transaction, VirtualAccountResponse } from "@/domain/models/transaction.types";
import { useAuthStore } from "@/store/auth.store";

/**
 * Custom hook encapsulating top-up and P2P transfer logic.
 * Generates reference_no automatically using UUID v4 to ensure idempotency.
 */
export function useTransfer() {
  const { user } = useAuthStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastTransaction, setLastTransaction] = useState<Transaction | null>(null);
  const [lastVA, setLastVA] = useState<VirtualAccountResponse | null>(null);

  const clearError = () => setError(null);

  const topup = async (amount: number) => {
    if (!user) return null;
    setIsLoading(true);
    setError(null);
    try {
      const res = await transferApi.topup({
        user_id: user.id,
        amount,
        reference_no: uuidv4(),
        description: "Top-up via Virtual Account",
      });
      if (res.data) {
        setLastTransaction(res.data);
        return res.data;
      }
      return null;
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Top-up failed");
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Creates or retrieves a Xendit Fixed Virtual Account for the given bank.
   * The name is taken from the authenticated user's profile.
   */
  const requestVA = async (bankCode: string) => {
    if (!user) return null;
    setIsLoading(true);
    setError(null);
    try {
      const res = await transferApi.requestVA({
        bank_code: bankCode,
        name: user.name ?? user.email,
      });
      if (res.data) {
        setLastVA(res.data);
        return res.data;
      }
      return null;
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Gagal membuat Virtual Account");
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  /**
   * Simulates a Xendit VA payment — triggers the same wallet-credit
   * pipeline as a real Xendit callback. Only works in non-production.
   */
  const simulatePayment = async (va: VirtualAccountResponse, amount: number) => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await transferApi.simulatePayment({
        external_id: va.external_id,
        bank_code: va.bank_code,
        payment_id: `sim-${uuidv4()}`,
        account_number: va.account_number,
        amount,
        transaction_timestamp: new Date().toISOString(),
      });
      if (res.data) {
        setLastTransaction(res.data);
        return res.data;
      }
      return null;
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Simulasi pembayaran gagal");
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  const transfer = async (
    targetEmail: string,
    amount: number,
    pin: string,
    description?: string,
  ) => {
    if (!user) return null;
    setIsLoading(true);
    setError(null);
    try {
      const res = await transferApi.p2pTransfer({
        source_user_id: user.id,
        target_email: targetEmail,
        amount,
        reference_no: uuidv4(),
        description: description ?? "P2P Transfer",
        transaction_pin: pin,
      });
      if (res.data) {
        setLastTransaction(res.data);
        return res.data;
      }
      return null;
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Transfer failed");
      return null;
    } finally {
      setIsLoading(false);
    }
  };

  return { isLoading, error, lastTransaction, lastVA, topup, requestVA, simulatePayment, transfer, clearError };
}
