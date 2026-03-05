import { apiClient } from "./client";
import type { Transaction, TopupRequest, TransferRequest } from "@/domain/models/transaction.types";

const TRANSFER_BASE = "/api/v1/transfers";

export const transferApi = {
  /** Credits the authenticated user's wallet (mock VA top-up). */
  topup: (body: TopupRequest) =>
    apiClient.post<Transaction>(`${TRANSFER_BASE}/topup`, body),

  /** Executes a P2P transfer from the authenticated user to a target user. */
  p2pTransfer: (body: TransferRequest) =>
    apiClient.post<Transaction>(`${TRANSFER_BASE}/p2p`, body),
};
