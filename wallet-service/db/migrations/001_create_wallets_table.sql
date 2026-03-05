-- Migration: 001_create_wallets_table.sql
-- Description: Creates the `wallets` table — the balance snapshot for each user.

CREATE TABLE IF NOT EXISTS `wallets` (
    `id`         VARCHAR(36)  NOT NULL                             COMMENT 'UUID primary key',
    `user_id`    VARCHAR(36)  NOT NULL                             COMMENT 'Foreign key to users.id in User Service',
    `balance`    BIGINT       NOT NULL DEFAULT 0                   COMMENT 'Current balance in smallest currency unit (e.g. cents)',
    `status`     ENUM('ACTIVE', 'INACTIVE', 'FROZEN')
                 NOT NULL DEFAULT 'ACTIVE'                         COMMENT 'Operational status of the wallet',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
                 ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_wallets_user_id` (`user_id`),
    KEY `idx_wallets_status` (`status`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Wallet balance snapshots for OmniWallet users';
