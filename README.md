# OmniWallet

Platform dompet digital (e-wallet) fullstack yang dibangun dengan arsitektur microservices, mensimulasikan ekosistem pembayaran seperti ShopeePay.

## Fitur Utama

- **Double-Entry Ledger** — setiap transaksi dijamin ACID, tidak ada dana yang hilang atau tergandakan
- **Idempotent API** — retry request tidak menyebabkan transaksi ganda
- **P2P Transfer** — transfer antar pengguna secara real-time
- **Top-up via Xendit Virtual Account** — integrasi payment gateway Xendit (Fixed VA), dilengkapi webhook callback dan simulate endpoint
- **Riwayat Transaksi** — mutasi masuk/keluar dengan pagination
- **Admin Dashboard** — manajemen pengguna dan monitoring transaksi
- **Async Notification** — notifikasi event transaksi via RabbitMQ (consumer tetap menerima pesan meski worker sempat down)

## Arsitektur

```
Browser / Frontend (Next.js 15)
        │
        ▼
   API Gateway :8080          ← single entry point, JWT auth, rate limiting
   ┌────────────────┐
   │  user-service  │ :8081   ← registrasi, login, KYC, profil
   │ wallet-service │ :8082   ← saldo, transfer, mutasi, Xendit VA
   └────────────────┘
        │                              │
   MySQL  │  Redis  │  RabbitMQ    Xendit API (test mode)
                         │              ↑ callback webhook
              notification-worker   ← async consumer event transaksi
```

### Alur Request (Synchronous)
```
Client ──HTTP──▶ API Gateway ──HTTP──▶ service tujuan
```
API Gateway meneruskan **full path** original ke upstream menggunakan `httputil.ReverseProxy` — hanya host yang diganti, path tetap utuh.

### Alur Notifikasi (Asynchronous via RabbitMQ)
```
wallet-service (setelah transaksi sukses)
    └─ publish [TOPUP_SUCCESS / TRANSFER_SUCCESS] ──▶ RabbitMQ
                                                          │
                                              notification-worker (consumer)
```
Jika notification-worker down, pesan tetap tersimpan di queue RabbitMQ dan akan diproses otomatis ketika worker hidup kembali.

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
- [Go 1.22+](https://go.dev/dl/) *(opsional — hanya jika ingin menjalankan wallet-service secara lokal tanpa Docker)*

> Untuk menjalankan via Docker Compose, Go dan Node.js tidak perlu diinstal secara lokal.

---

## Cara Clone

```bash
git clone https://github.com/srgjo27/omni-wallet.git
cd omni-wallet
```

---

## Cara Menjalankan

### 1. Salin file environment

Semua nilai konfigurasi disimpan di folder `envs/` (satu file per service) dan tidak ter-commit ke Git.

```bash
# Buat folder envs dan file konfigurasi masing-masing service
mkdir -p envs
```

Buat file berikut sesuai kebutuhan (contoh nilai sudah tersedia di bagian [Konfigurasi Environment](#konfigurasi-environment)):

```
envs/
├── mysql.env
├── rabbitmq.env
├── user-service.env
├── wallet-service.env
├── api-gateway.env
├── notification-worker.env
└── frontend.env
```

> ⚠️ `JWT_SECRET` harus **identik** di `user-service.env`, `api-gateway.env`, dan `wallet-service.env`.

### 2. Salin file environment frontend

```bash
cp frontend/.env.example frontend/.env.local
```

### 3. Jalankan semua service

```bash
docker compose up --build -d
```

Perintah ini akan:
- Build image Docker untuk semua service
- Menjalankan MySQL, Redis, dan RabbitMQ
- Menjalankan migrasi database secara otomatis (dengan retry loop hingga MySQL siap)
- Meluncurkan seluruh stack di background

> Proses build pertama kali membutuhkan waktu sekitar 2–5 menit tergantung koneksi internet.

### 4. Buka aplikasi

| Service | URL |
|---|---|
| Frontend (Web App) | http://localhost:3000 |
| API Gateway | http://localhost:8080 |
| RabbitMQ Management UI | http://localhost:15672 |

> Kredensial RabbitMQ Management: `omni_user` / `omni_password`

### 5. Menghentikan semua service

```bash
# Hentikan tanpa menghapus data
docker compose down

# Hentikan dan hapus semua data (database, cache)
docker compose down -v
```

### Clean Rebuild (hapus semua cache)

Gunakan jika terjadi masalah build atau ingin memulai dari awal:

```bash
docker compose down -v --rmi all
docker builder prune -af
docker compose up --build -d
```

> ⚠️ Setelah mengubah file `envs/*.env`, **jangan** gunakan `docker compose restart` karena tidak membaca ulang env_file. Gunakan `docker compose up -d <service>` untuk recreate container.

---

## Konfigurasi Environment

Berikut contoh isi masing-masing file di folder `envs/`:

<details>
<summary><strong>envs/mysql.env</strong></summary>

```env
MYSQL_ROOT_PASSWORD=
MYSQL_DATABASE=omni_wallet_users
MYSQL_USER=
MYSQL_PASSWORD=
```
</details>

<details>
<summary><strong>envs/rabbitmq.env</strong></summary>

```env
RABBITMQ_DEFAULT_USER=
RABBITMQ_DEFAULT_PASS=
```
</details>

<details>
<summary><strong>envs/user-service.env</strong></summary>

```env
DB_HOST=mysql
DB_PORT=3306
DB_USER=
DB_PASSWORD=
DB_NAME=omni_wallet_users
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
JWT_SECRET=your-secret-min-32-chars-here
JWT_TTL=24h
SERVER_PORT=8081
```
</details>

<details>
<summary><strong>envs/wallet-service.env</strong></summary>

```env
DB_HOST=mysql
DB_PORT=3306
DB_USER=
DB_PASSWORD=
DB_NAME=omni_wallet_transactions
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
RABBITMQ_URL=amqp://omni_user:omni_password@rabbitmq:5672/
JWT_SECRET=your-secret-min-32-chars-here
USER_SERVICE_BASE_URL=http://user-service:8081
XENDIT_API_KEY=xnd_development_xxxxxx
SERVER_PORT=8082
```
</details>

<details>
<summary><strong>envs/api-gateway.env</strong></summary>

```env
JWT_SECRET=your-secret-min-32-chars-here
USER_SERVICE_URL=http://user-service:8081
WALLET_SERVICE_URL=http://wallet-service:8082
SERVER_PORT=8080
```
</details>

<details>
<summary><strong>envs/notification-worker.env</strong></summary>

```env
RABBITMQ_URL=amqp://omni_user:omni_password@rabbitmq:5672/
```
</details>

<details>
<summary><strong>envs/frontend.env</strong></summary>

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```
</details>

---

## Integrasi Xendit (Virtual Account)

Proyek ini menggunakan Xendit Fixed Virtual Account untuk fitur top-up (mode test).

1. Daftar akun Xendit dan dapatkan API key dari [Xendit Dashboard](https://dashboard.xendit.co)
2. Pastikan permission **Fixed Virtual Account → Write** dan **Fixed Virtual Account → Read** aktif di API key
3. Set `XENDIT_API_KEY` di `envs/wallet-service.env`

### Alur Top-up via Xendit VA
```
Client POST /api/v1/transfers/topup/va
    │
    ▼
wallet-service → Xendit API (buat Fixed VA)
    │
    ▼ (kembalikan nomor VA ke client)

Client melakukan pembayaran ke VA
    │
    ▼
Xendit → POST /api/v1/payments/xendit/callback (webhook)
    │
    ▼
wallet-service update saldo + catat transaksi
```

### Simulate Pembayaran (Test Mode)
```bash
POST /api/v1/payments/xendit/simulate
Authorization: Bearer <token>
Content-Type: application/json

{
  "external_id": "VA-xxxx",
  "amount": 100000
}
```

---

## Struktur Proyek

```
omni-wallet/
├── api-gateway/          # Go — single entry point, JWT auth, rate limiting
│   └── internal/
│       └── adapter/proxy/router.go   # routing table ke upstream services
├── user-service/         # Go — registrasi, login, profil, KYC
│   ├── db/migrations/    # SQL migration files
│   ├── .env.local        # env untuk local dev (127.0.0.1 hosts)
│   └── Makefile          # make dev / build / run / stop-docker / start-docker
├── wallet-service/       # Go — saldo, transfer P2P, mutasi, Xendit VA
│   ├── db/migrations/
│   ├── .env.local        # env untuk local dev (127.0.0.1 hosts)
│   └── Makefile          # make dev / build / run / stop-docker / start-docker
├── notification-worker/  # Go — RabbitMQ consumer, async notification
├── frontend/             # Next.js 15 — dashboard user & admin
│   └── src/
│       ├── app/          # Next.js App Router pages
│       ├── components/   # UI components (atoms, features, layouts)
│       ├── domain/       # models, use-cases
│       ├── infrastructure/ # API client
│       └── store/        # Zustand stores
├── db/                   # init.sql — inisialisasi database & user grants
├── envs/                 # ⚠️ di-gitignore — berisi *.env per service
└── docker-compose.yml
```

---

## Endpoint API Utama

Semua request melalui API Gateway di `http://localhost:8080`.

### Auth
| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| POST | `/api/v1/users/register` | — | Registrasi pengguna baru |
| POST | `/api/v1/users/login` | — | Login, mendapatkan JWT |
| POST | `/api/v1/users/logout` | JWT | Logout, invalidasi session |

### Profil
| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| GET | `/api/v1/users/profile` | JWT | Profil pengguna |
| PUT | `/api/v1/users/pin` | JWT | Set PIN transaksi |
| PUT | `/api/v1/users/kyc` | JWT | Submit KYC |

### Wallet
| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| GET | `/api/v1/wallets/balance` | JWT | Cek saldo |
| GET | `/api/v1/wallets/mutations` | JWT | Riwayat mutasi |
| GET | `/api/v1/wallets/transactions` | JWT | Riwayat transaksi |
| POST | `/api/v1/transfers/topup/va` | JWT | Buat Xendit Fixed VA untuk top-up |
| POST | `/api/v1/transfers/p2p` | JWT | Transfer antar pengguna |

### Payments (Xendit)
| Method | Endpoint | Auth | Keterangan |
|---|---|---|---|
| POST | `/api/v1/payments/xendit/callback` | — | Webhook dari Xendit (publik) |
| POST | `/api/v1/payments/xendit/simulate` | JWT | Simulasi pembayaran VA (test mode) |

---

## Menjalankan wallet-service Secara Lokal

Berguna untuk debugging dengan log langsung di terminal, tanpa rebuild Docker image.

### Prasyarat
- MySQL, Redis, RabbitMQ tetap berjalan di Docker
- Port `3306`, `6379`, `5672`, dan `8081` sudah ter-expose ke host (sudah dikonfigurasi di `docker-compose.yml`)

### Langkah

```bash
cd wallet-service

# Hentikan container wallet-service agar port 8082 bebas
make stop-docker

# Jalankan lokal dengan .env.local (host = 127.0.0.1)
make dev
```

### Makefile targets tersedia

| Target | Perintah | Keterangan |
|---|---|---|
| `make dev` | `go run ./cmd/api` | Jalankan dengan hot-reload manual, baca `.env.local` |
| `make build` | `go build -o bin/wallet-service` | Compile binary |
| `make run` | `./bin/wallet-service` | Jalankan binary hasil build |
| `make stop-docker` | `docker compose stop wallet-service` | Hentikan container, bebaskan port 8082 |
| `make start-docker` | `docker compose up -d wallet-service` | Kembalikan ke Docker |

> File `wallet-service/.env.local` berisi konfigurasi yang sama dengan `envs/wallet-service.env` namun dengan semua hostname diganti `127.0.0.1`.

---

## Lisensi

Proyek ini dibuat untuk keperluan pembelajaran dan demonstrasi portofolio.
