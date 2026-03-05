"use client";

import { useState } from "react";
import { ArrowUpRight, ArrowDownLeft } from "lucide-react";
import { useMutations } from "@/domain/use-cases/useWallet";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Spinner } from "@/components/ui/Spinner";
import { formatCurrency, formatDate } from "@/lib/format";

export function MutationList() {
  const [page, setPage] = useState(1);
  const pageSize = 10;
  const { mutations, total, isLoading, error } = useMutations(page, pageSize);

  if (isLoading) return <div className="flex justify-center py-10"><Spinner size="md" /></div>;
  if (error) return <p className="py-6 text-center text-sm text-gray-500">Gagal memuat mutasi.</p>;
  if (!mutations.length) return <p className="py-6 text-center text-sm text-gray-500">Belum ada mutasi.</p>;

  return (
    <div className="flex flex-col gap-3">
      <ul className="divide-y divide-gray-100 rounded-xl border border-gray-200 bg-white">
        {mutations.map((m) => {
          const isCredit = m.direction === "CREDIT";
          return (
            <li key={m.id} className="flex items-center gap-3 px-4 py-3">
              <div
                className={`flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full ${
                  isCredit ? "bg-green-50 text-green-600" : "bg-red-50 text-red-600"
                }`}
              >
                {isCredit ? <ArrowDownLeft className="h-4 w-4" /> : <ArrowUpRight className="h-4 w-4" />}
              </div>

              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-gray-900">
                  {m.description || (isCredit ? "Masuk" : "Keluar")}
                </p>
                <p className="mt-0.5 text-xs text-gray-400">
                  {formatDate(m.created_at)}
                </p>
              </div>

              <span
                className={`text-sm font-semibold ${isCredit ? "text-green-600" : "text-red-600"}`}
              >
                {isCredit ? "+" : "−"}{formatCurrency(m.amount)}
              </span>
            </li>
          );
        })}
      </ul>

      {/* Pagination */}
      {total > pageSize && (
        <div className="flex items-center justify-between text-sm text-gray-500">
          <span>
            Hal {page} dari {Math.ceil(total / pageSize)}
          </span>
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
