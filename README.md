# OmniWallet

Platform dompet digital (e-wallet) fullstack yang dibangun dengan arsitektur microservices, mensimulasikan ekosistem pembayaran seperti ShopeePay.

## Fitur Utama

- **Double-Entry Ledger** — setiap transaksi dijamin ACID, tidak ada dana yang hilang atau tergandakan
- **Idempotent API** — retry request tidak menyebabkan transaksi ganda
- **P2P Transfer** — transfer antar pengguna secara real-time
- **Top-up Balance** — simulasi Virtual Account (mock webhook bank)
- **Riwayat Transaksi** — mutasi masuk/keluar dengan pagination
- **Admin Dashboard** — manajemen pengguna dan monitoring transaksi
- **Async Notification** — notifikasi event transaksi via RabbitMQ

## Arsitektur

```
Browser / Frontend (Next.js 15)
        │
        ▼
   API Gateway :8080          ← single entry point, JWT auth, rate limiting
   ┌────────────────┐
   │  user-service  │ :8081   ← registrasi, login, KYC, profil
   │ wallet-service │ :8082   ← saldo, transfer, mutasi
   └────────────────┘
        │
   MySQL  │  Redis  │  RabbitMQ
                         │
              notification-worker   ← async consumer event transaksi
```

## Tech Stack

| Layer | Teknologi |
|---|---|
| Backend | Go 1.22, Gin, Clean Architecture |
| Database | MySQL 8.0 |
| Cache & Lock | Redis 7 |
| Message Broker | RabbitMQ 3 |
| Frontend | Next.js 15, React 19, TypeScript, Tailwind CSS v4 |
| State Management | Zustand v5, SWR v2 |
| Infrastruktur | Docker, Docker Compose |

## Prasyarat

Pastikan sudah terinstal di mesin lokal Anda:

- [Docker](https://www.docker.com/get-started) v24+ dan Docker Compose v2.20+
- [Git](https://git-scm.com/)

> Tidak perlu menginstal Go atau Node.js secara lokal — semua dijalankan di dalam container Docker.

---

## Cara Clone

```bash
git clone https://github.com/srgjo27/omni-wallet.git
cd omni-wallet
```

---

## Cara Menjalankan

### 1. Salin file environment frontend

```bash
cp frontend/.env.example frontend/.env.local
```

File `.env.local` sudah dikonfigurasi untuk berjalan dengan Docker Compose dan tidak perlu diubah.

### 2. Jalankan semua service

```bash
docker compose up --build
```

Perintah ini akan:
- Build image Docker untuk semua service (user-service, wallet-service, api-gateway, notification-worker, frontend)
- Menjalankan MySQL, Redis, dan RabbitMQ
- Menjalankan migrasi database secara otomatis
- Meluncurkan seluruh stack

> Proses build pertama kali membutuhkan waktu sekitar 2–5 menit tergantung koneksi internet.

### 3. Buka aplikasi

| Service | URL |
|---|---|
| Frontend (Web App) | http://localhost:3000 |
| API Gateway | http://localhost:8080 |
| RabbitMQ Management UI | http://localhost:15672 |

> Kredensial RabbitMQ Management: `omni_user` / `omni_password`

### 4. Menghentikan semua service

```bash
# Hentikan tanpa menghapus data
docker compose down

# Hentikan dan hapus semua data (database, cache)
docker compose down -v
```

---

## Struktur Proyek

```
omni-wallet/
├── api-gateway/          # Go — single entry point, JWT auth, rate limiting
├── user-service/         # Go — registrasi, login, profil, KYC
│   └── db/migrations/    # SQL migration files
├── wallet-service/       # Go — saldo, transfer P2P, mutasi
│   └── db/migrations/
├── notification-worker/  # Go — RabbitMQ consumer, async notification
├── frontend/             # Next.js 15 — dashboard user & admin
│   └── src/
│       ├── app/          # Next.js App Router pages
│       ├── components/   # UI components (atoms, features, layouts)
│       ├── domain/       # models, use-cases
│       ├── infrastructure/ # API client
│       └── store/        # Zustand stores
├── db/                   # init.sql — inisialisasi database
└── docker-compose.yml
```

---

## Endpoint API Utama

Semua request melalui API Gateway di `http://localhost:8080`.

### Auth
| Method | Endpoint | Keterangan |
|---|---|---|
| POST | `/api/v1/users/register` | Registrasi pengguna baru |
| POST | `/api/v1/users/login` | Login, mendapatkan JWT |
| POST | `/api/v1/users/logout` | Logout, invalidasi session |

### Profil
| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/api/v1/users/me` | Profil pengguna |
| POST | `/api/v1/users/pin` | Set PIN transaksi |
| POST | `/api/v1/users/kyc` | Submit KYC |

### Wallet
| Method | Endpoint | Keterangan |
|---|---|---|
| GET | `/api/v1/wallets/balance` | Cek saldo |
| GET | `/api/v1/wallets/mutations` | Riwayat mutasi |
| GET | `/api/v1/wallets/transactions` | Riwayat transaksi |
| POST | `/api/v1/transfers/topup` | Top-up saldo |
| POST | `/api/v1/transfers/p2p` | Transfer antar pengguna |

---

## Lisensi

Proyek ini dibuat untuk keperluan pembelajaran dan demonstrasi portofolio.
