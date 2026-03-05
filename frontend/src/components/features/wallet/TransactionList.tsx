"use client";

import { useState } from "react";
import { useTransactionHistory } from "@/domain/use-cases/useWallet";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Spinner } from "@/components/ui/Spinner";
import { formatCurrency, formatDate, shortId } from "@/lib/format";
import type { TransactionStatus, TransactionType } from "@/domain/models/transaction.types";

const statusBadge: Record<TransactionStatus, { variant: "success" | "danger" | "warning" | "neutral"; label: string }> = {
  SUCCESS: { variant: "success", label: "Selesai" },
  FAILED:  { variant: "danger",  label: "Gagal" },
  PENDING: { variant: "warning", label: "Proses" },
};

const typeLabel: Record<TransactionType, string> = {
  TOPUP:   "Top Up",
  P2P:     "Transfer",
  PAYMENT: "Pembayaran",
};

export function TransactionList() {
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const { transactions, total, isLoading, error } = useTransactionHistory(page, pageSize);

  if (isLoading) return <div className="flex justify-center py-10"><Spinner size="md" /></div>;
  if (error) return <p className="py-6 text-center text-sm text-gray-500">Gagal memuat transaksi.</p>;
  if (!transactions.length) return <p className="py-6 text-center text-sm text-gray-500">Belum ada transaksi.</p>;

  return (
    <div className="flex flex-col gap-3">
      <div className="overflow-x-auto rounded-xl border border-gray-200">
        <table className="min-w-full divide-y divide-gray-100 bg-white">
          <thead>
            <tr className="text-left text-xs font-medium uppercase tracking-wide text-gray-400">
              <th className="px-4 py-3">Referensi</th>
              <th className="px-4 py-3">Jenis</th>
              <th className="px-4 py-3">Jumlah</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Tanggal</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {transactions.map((tx) => {
              const sb = statusBadge[tx.status] ?? statusBadge.PENDING;
              return (
                <tr key={tx.id} className="text-sm text-gray-700 hover:bg-gray-50">
                  <td className="px-4 py-3 font-mono text-xs text-gray-500">#{shortId(tx.reference_no)}</td>
                  <td className="px-4 py-3">{typeLabel[tx.type] ?? tx.type}</td>
                  <td className="px-4 py-3 font-medium">{formatCurrency(tx.amount)}</td>
                  <td className="px-4 py-3">
                    <Badge variant={sb.variant}>{sb.label}</Badge>
                  </td>
                  <td className="px-4 py-3 text-xs text-gray-400">{formatDate(tx.created_at)}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {total > pageSize && (
        <div className="flex items-center justify-between text-sm text-gray-500">
          <span>Hal {page} dari {Math.ceil(total / pageSize)}</span>
          <div className="flex gap-2">
            <Button variant="secondary" size="sm" onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1}>
              ← Sebelumnya
            </Button>
            <Button variant="secondary" size="sm" onClick={() => setPage((p) => p + 1)} disabled={page * pageSize >= total}>
              Berikutnya →
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
