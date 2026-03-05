"use client";

import { TransactionList } from "@/components/features/wallet/TransactionList";

export default function AdminTransactionsPage() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <h2 className="text-base font-semibold text-gray-700">Riwayat Transaksi</h2>
        <p className="mt-1 text-sm text-gray-500">
          Menampilkan transaksi akun admin. Fitur tampilan seluruh transaksi pengguna
          memerlukan endpoint khusus yang dapat ditambahkan di iterasi berikutnya.
        </p>
      </div>

      <TransactionList />
    </div>
  );
}
