-- Migration: 002_create_transactions_table.sql
-- Description: Creates the `transactions` table — the master record of every
--              financial intent with its lifecycle status (PENDING → SUCCESS | FAILED).

CREATE TABLE IF NOT EXISTS `transactions` (
    `id`               VARCHAR(36)   NOT NULL                              COMMENT 'UUID primary key',
    `reference_no`     VARCHAR(64)   NOT NULL                              COMMENT 'Unique external reference / idempotency key',
    `type`             ENUM('TOPUP', 'P2P', 'PAYMENT')
                       NOT NULL                                            COMMENT 'Transaction type',
    `amount`           BIGINT        NOT NULL                              COMMENT 'Transaction amount in smallest currency unit',
    `status`           ENUM('PENDING', 'SUCCESS', 'FAILED')
                       NOT NULL DEFAULT 'PENDING'                          COMMENT 'Lifecycle status of the transaction',
    `source_wallet_id` VARCHAR(36)   NULL     DEFAULT NULL                 COMMENT 'Sender wallet ID (NULL for TOPUP)',
    `target_wallet_id` VARCHAR(36)   NULL     DEFAULT NULL                 COMMENT 'Receiver wallet ID',
    `description`      VARCHAR(255)  NULL     DEFAULT NULL                 COMMENT 'Optional human-readable description',
    `created_at`       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP
                       ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_transactions_reference_no` (`reference_no`),
    KEY `idx_transactions_source_wallet` (`source_wallet_id`),
    KEY `idx_transactions_target_wallet` (`target_wallet_id`),
    KEY `idx_transactions_status`        (`status`),
    KEY `idx_transactions_type`          (`type`),
    KEY `idx_transactions_created_at`    (`created_at`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Master record of all financial transactions in OmniWallet';
