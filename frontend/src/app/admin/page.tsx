"use client";

import useSWR from "swr";
import { Users, ArrowRightLeft, TrendingUp, ShieldCheck } from "lucide-react";
import { StatsCard } from "@/components/features/admin/StatsCard";
import { authApi } from "@/infrastructure/api/auth.api";
import { walletApi } from "@/infrastructure/api/wallet.api";
import { formatCurrency } from "@/lib/format";

function useAdminStats() {
  const { data: userStats, isLoading: userLoading } = useSWR(
    "/api/v1/users/stats",
    () => authApi.adminGetUserStats().then((r) => r.data),
    { refreshInterval: 30_000 },
  );
  const { data: walletStats, isLoading: walletLoading } = useSWR(
    "/api/v1/wallets/stats",
    () => walletApi.adminGetWalletStats().then((r) => r.data),
    { refreshInterval: 30_000 },
  );
  return {
    totalUsers:    userStats?.total_users,
    verifiedUsers: userStats?.verified_users,
    totalTx:       walletStats?.total_transactions,
    totalVolume:   walletStats?.total_volume,
    isLoading: userLoading || walletLoading,
  };
}

export default function AdminOverviewPage() {
  const { totalUsers, verifiedUsers, totalTx, totalVolume, isLoading } = useAdminStats();

  const fmt = (v: number | undefined) => (isLoading || v === undefined ? "…" : String(v));

  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatsCard
          label="Total Pengguna"
          value={fmt(totalUsers)}
          icon={<Users className="h-5 w-5" />}
          colorClass="text-indigo-600 bg-indigo-50"
          description="Semua pengguna terdaftar"
        />
        <StatsCard
          label="Total Transaksi"
          value={fmt(totalTx)}
          icon={<ArrowRightLeft className="h-5 w-5" />}
          colorClass="text-green-600 bg-green-50"
          description="Transaksi berhasil"
        />
        <StatsCard
          label="Volume (Rp)"
          value={isLoading || totalVolume === undefined ? "…" : formatCurrency(totalVolume)}
          icon={<TrendingUp className="h-5 w-5" />}
          colorClass="text-blue-600 bg-blue-50"
          description="Total nominal berpindah"
        />
        <StatsCard
          label="Pengguna Terverifikasi"
          value={fmt(verifiedUsers)}
          icon={<ShieldCheck className="h-5 w-5" />}
          colorClass="text-emerald-600 bg-emerald-50"
          description="KYC selesai"
        />
      </div>
    </div>
  );
}
