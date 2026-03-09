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
  target_email: string;
  amount: number;
  reference_no: string;
  description?: string;
  transaction_pin: string;
}

// ── Xendit Virtual Account types ──

/** POST /api/v1/transfers/topup/va — request body */
export interface RequestVARequest {
  bank_code: string;
  name: string;
}

/** POST /api/v1/transfers/topup/va — response data */
export interface VirtualAccountResponse {
  id: string;
  external_id: string;
  bank_code: string;
  name: string;
  account_number: string;
  merchant_code: string;
  currency: string;
  is_closed: boolean;
  status: string;
}

/** POST /api/v1/payments/xendit/simulate — request body */
export interface SimulatePaymentRequest {
  external_id: string;
  bank_code: string;
  payment_id: string;
  account_number: string;
  amount: number;
  transaction_timestamp: string;
}
