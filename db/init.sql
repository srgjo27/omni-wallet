-- OmniWallet — MySQL Initialisation Script
-- Executed once when the MySQL container is first created.
-- Ensures both service databases exist and the shared user has full access.

CREATE DATABASE IF NOT EXISTS `omni_wallet_users`
    CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE DATABASE IF NOT EXISTS `omni_wallet_wallets`
    CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

GRANT ALL PRIVILEGES ON `omni_wallet_users`.*  TO 'omni_user'@'%';
GRANT ALL PRIVILEGES ON `omni_wallet_wallets`.* TO 'omni_user'@'%';
FLUSH PRIVILEGES;
