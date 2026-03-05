-- Migration: 003_create_wallet_mutations_table.sql
-- Description: Creates the `wallet_mutations` table — the immutable double-entry ledger.
--              Every balance change MUST produce an entry here for full auditability.

CREATE TABLE IF NOT EXISTS `wallet_mutations` (
    `id`             VARCHAR(36)       NOT NULL                            COMMENT 'UUID primary key',
    `wallet_id`      VARCHAR(36)       NOT NULL                            COMMENT 'The wallet whose balance changed',
    `transaction_id` VARCHAR(36)       NOT NULL                            COMMENT 'The transaction that caused this mutation',
    `direction`      ENUM('CREDIT', 'DEBIT')
                     NOT NULL                                              COMMENT 'CREDIT = balance increased, DEBIT = balance decreased',
    `amount`         BIGINT            NOT NULL                            COMMENT 'Absolute amount of the change (always positive)',
    `balance_after`  BIGINT            NOT NULL                            COMMENT 'Wallet balance immediately after this mutation (audit snapshot)',
    `description`    VARCHAR(255)      NULL     DEFAULT NULL               COMMENT 'Optional human-readable description',
    `created_at`     DATETIME          NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    KEY `idx_mutations_wallet_id`      (`wallet_id`),
    KEY `idx_mutations_transaction_id` (`transaction_id`),
    KEY `idx_mutations_created_at`     (`created_at`),
    KEY `idx_mutations_wallet_created` (`wallet_id`, `created_at`)         COMMENT 'Optimised index for paginated mutation history queries'
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Double-entry ledger: immutable record of every wallet balance change';
