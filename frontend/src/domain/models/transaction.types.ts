// ── Types mirroring the wallet-service transaction domain ──

export type TransactionType = "TOPUP" | "P2P" | "PAYMENT";
export type TransactionStatus = "PENDING" | "SUCCESS" | "FAILED";

export interface Transaction {
  id: string;
  reference_no: string;
  type: TransactionType;
  amount: number;
  status: TransactionStatus;
  source_wallet_id?: string;
  target_wallet_id?: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export interface TransactionHistoryResponse {
  transactions: Transaction[];
  total: number;
  page: number;
  page_size: number;
}

/** POST /api/v1/transfers/topup */
export interface TopupRequest {
  user_id: string;
  amount: number;
  reference_no: string;
  description?: string;
}

/** POST /api/v1/transfers/p2p */
export interface TransferRequest {
  source_user_id: string;
  target_user_id: string;
  amount: number;
  reference_no: string;
  description?: string;
  transaction_pin: string;
}
