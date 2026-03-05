-- Migration: 001_create_users_table.sql
-- Description: Creates the initial `users` table for the User Service.
-- Run this script once against the target database before starting the service.

CREATE TABLE IF NOT EXISTS `users` (
    `id`            VARCHAR(36)  NOT NULL                             COMMENT 'UUID primary key',
    `name`          VARCHAR(100) NOT NULL                             COMMENT 'Full display name of the user',
    `email`         VARCHAR(255) NOT NULL                             COMMENT 'Unique email address used for authentication',
    `password_hash` VARCHAR(255) NOT NULL                             COMMENT 'bcrypt hash of the user login password',
    `pin_hash`      VARCHAR(255) NOT NULL DEFAULT ''                  COMMENT 'bcrypt hash of the 6-digit transaction PIN; empty until user sets it',
    `kyc_status`    ENUM('UNVERIFIED', 'PENDING', 'VERIFIED')
                    NOT NULL DEFAULT 'UNVERIFIED'                     COMMENT 'KYC verification status of the user',
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP   COMMENT 'Timestamp when the record was created',
    `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
                    ON UPDATE CURRENT_TIMESTAMP                       COMMENT 'Timestamp of the last update',

    PRIMARY KEY (`id`),
    UNIQUE KEY `uq_users_email` (`email`)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci
  COMMENT='Stores user profile and authentication data for OmniWallet';
