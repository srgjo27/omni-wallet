"use client";

import { Wallet, RefreshCw } from "lucide-react";
import { useWallet } from "@/domain/use-cases/useWallet";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Spinner } from "@/components/ui/Spinner";
import { formatCurrency } from "@/lib/format";

const statusMap = {
  ACTIVE:   { variant: "success" as const, label: "Aktif" },
  INACTIVE: { variant: "neutral" as const, label: "Tidak Aktif" },
  FROZEN:   { variant: "danger"  as const, label: "Dibekukan" },
};

export function BalanceCard() {
  const { balance, isBalanceLoading, balanceError, refreshBalance } = useWallet();

  if (isBalanceLoading) {
    return (
      <Card className="flex min-h-[140px] items-center justify-center">
        <Spinner size="md" />
      </Card>
    );
  }

  if (balanceError || !balance) {
    return (
      <Card className="flex min-h-[140px] flex-col items-center justify-center gap-2">
        <p className="text-sm text-gray-500">Gagal memuat saldo</p>
        <button
          onClick={() => refreshBalance()}
          className="flex items-center gap-1 text-sm font-medium text-indigo-600 hover:underline"
        >
          <RefreshCw className="h-3 w-3" /> Coba lagi
        </button>
      </Card>
    );
  }

  const status = statusMap[balance.status] ?? statusMap.INACTIVE;

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-gray-500">
            <Wallet className="h-4 w-4" /> Saldo Tersedia
          </CardTitle>
          <Badge variant={status.variant}>{status.label}</Badge>
        </div>
      </CardHeader>

      <div className="px-5 pb-5">
        <p className="text-3xl font-bold tracking-tight text-gray-900">
          {formatCurrency(balance.balance)}
        </p>
        <p className="mt-1 text-xs text-gray-400">
          Dompet #{balance.wallet_id.split("-")[0]}
        </p>
      </div>
    </Card>
  );
}
