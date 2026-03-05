# Product Requirements Document (PRD): OmniWallet Platform

## 1. Tujuan & Solusi

**Tujuan:** Membangun core system dompet digital (e-wallet) yang aman, stabil, dan mampu menangani ribuan transaksi per detik (TPS), menyimulasikan ekosistem pembayaran seperti ShopeePay.

**Solusi & Fitur Utama:**

1. Double-Entry Ledger System: Sistem pencatatan keuangan internal yang menjamin tidak ada dana yang "hilang" atau "tergandakan" (ACID Compliance).
2. Idempotent API: Memastikan jika terjadi retry network (klik bayar 2 kali), transaksi tetap hanya diproses satu kali.
3. P2P Transfer & Merchant Payment: Transfer antar pengguna dan pembayaran ke sistem eksternal (Merchant).
4. High-Concurrency Handling: Menggunakan sistem Distributed Lock untuk mencegah Race Condition saat saldo pengguna diakses bersamaan.

## 2. Alur Bisnis (Business Flow)

1. User Registration & KYC: Pengguna mendaftar. Sistem membuatkan entitas User sekaligus entitas Wallet dengan saldo awal Rp0.
2. Top-up Balance: Pengguna melakukan Top-up via mock Virtual Account Bank. Sistem akan memvalidasi webhook dari bank dan menambah saldo Wallet.
3. P2P Transfer: Pengguna A mengirim uang ke Pengguna B. Sistem melakukan validasi saldo, memotong saldo A, menambah saldo B dalam satu Database Transaction.
4. Transaction History (Mutation): Pengguna melihat mutasi masuk dan keluar secara real-time atau berdasarkan filter tanggal.

## 3. Stack Teknologi (Fullstack Advance)

1. Backend (Core System): Golang (Framework: Gin atau Fiber karena sangat cepat dan umum digunakan di ekosistem Go).
2. Database: MySQL (Relasional, sangat ditekankan untuk transaksi finansial karena fitur ACID-nya).
3. Caching & Concurrency Control: Redis (Untuk menyimpan session, idempotency key, dan Distributed Lock / Mutex).
4. Message Broker: RabbitMQ atau Apache Kafka (Untuk memproses log riwayat transaksi secara asynchronous agar API utama merespons lebih cepat).
5. Frontend (Dashboard): Next.js (React) dipadukan dengan TailwindCSS v4++ untuk membuat dasbor User dan Admin/Merchant.
6. Infrastruktur: Docker & Docker Compose (Simulasi environment Linux).

## 4. Skema Sistem (System Architecture)

Sistem ini akan dipecah menjadi beberapa Microservices kecil agar skalabel.
1. API Gateway: Menangani routing, Rate Limiting (mencegah spam/DDoS), dan Autentikasi (JWT).
2. User Service: Mengelola data pengguna, PIN transaksi, dan KYC.
3. Wallet Service (Core Ledger): Menangani logika mutasi saldo, Top-up, dan Transfer. Ini adalah modul paling krusial.
4. Notification Worker: Mengonsumsi pesan dari RabbitMQ/Kafka untuk mengirimkan notifikasi (misal: "Transfer Berhasil") tanpa memblokir proses transaksi.

## 5. Skema Database (MySQL)

Sistem E-wallet yang baik tidak hanya menyimpan "Saldo" sebagai satu angka, tetapi mencatat setiap pergerakan.

Table Name,Kolom Penting,Deskripsi
users,"id, name, email, pin_hash",Data profil pengguna.
wallets,"id, user_id, balance, status",Menyimpan snapshot saldo saat ini.
transactions,"id, reference_no (Unique), type (TOPUP, P2P, PAYMENT), amount, status (PENDING, SUCCESS, FAILED), created_at",Tabel utama untuk mencatat niat/status sebuah transaksi.
wallet_mutations,"id, wallet_id, transaction_id, amount (Bisa + atau -), balance_after",Ledger sebenarnya. Digunakan untuk audit dan rekonsiliasi.

| Table Name | Kolom Penting | Deskripsi |
| --- | --- | --- |
| `users` | id, name, email, pin_hash | Data profil pengguna. |
| `wallets` | id, user_id, balance, status | Menyimpan snapshot saldo saat ini. |
| `transactions` | id, reference_no (Unique), type (TOPUP, P2P, PAYMENT), amount, status (PENDING, SUCCESS, FAILED), created_at | Tabel utama untuk mencatat niat/status sebuah transaksi. |
| `wallet_mutations` | id, wallet_id, transaction_id, amount (Bisa + atau -), balance_after | Ledger sebenarnya. Digunakan untuk audit dan rekonsiliasi. |