import { apiClient } from "./client";
import type {
  Transaction,
  TopupRequest,
  TransferRequest,
  RequestVARequest,
  VirtualAccountResponse,
  SimulatePaymentRequest,
} from "@/domain/models/transaction.types";

const TRANSFER_BASE = "/api/v1/transfers";
const PAYMENT_BASE = "/api/v1/payments/xendit";

export const transferApi = {
  /** Credits the authenticated user's wallet (mock VA top-up). */
  topup: (body: TopupRequest) =>
    apiClient.post<Transaction>(`${TRANSFER_BASE}/topup`, body),

  /** Executes a P2P transfer from the authenticated user to a target user. */
  p2pTransfer: (body: TransferRequest) =>
    apiClient.post<Transaction>(`${TRANSFER_BASE}/p2p`, body),

  /**
   * Creates or retrieves a Xendit Fixed Virtual Account for the authenticated user.
   * Subsequent calls with the same bank_code return the cached VA.
   */
  requestVA: (body: RequestVARequest) =>
    apiClient.post<VirtualAccountResponse>(`${TRANSFER_BASE}/topup/va`, body),

  /**
   * Simulates a Xendit VA payment callback for local/test-mode development.
   * Only available when APP_ENV != production on the wallet-service.
   */
  simulatePayment: (body: SimulatePaymentRequest) =>
    apiClient.post<Transaction>(`${PAYMENT_BASE}/simulate`, body),
};
