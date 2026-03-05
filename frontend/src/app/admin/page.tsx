import { Users, ArrowRightLeft, TrendingUp, ShieldCheck } from "lucide-react";
import { StatsCard } from "@/components/features/admin/StatsCard";
import type { Metadata } from "next";

export const metadata: Metadata = { title: "Admin Overview | OmniWallet" };

/**
 * Admin overview page with placeholder stats.
 * In production, wire these up to real admin API endpoints.
 */
export default function AdminOverviewPage() {
  return (
    <div className="flex flex-col gap-6">
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatsCard
          label="Total Pengguna"
          value="—"
          icon={<Users className="h-5 w-5" />}
          colorClass="text-indigo-600 bg-indigo-50"
          description="Data dari user-service"
        />
        <StatsCard
          label="Total Transaksi"
          value="—"
          icon={<ArrowRightLeft className="h-5 w-5" />}
          colorClass="text-green-600 bg-green-50"
          description="Semua jenis transaksi"
        />
        <StatsCard
          label="Volume (Rp)"
          value="—"
          icon={<TrendingUp className="h-5 w-5" />}
          colorClass="text-blue-600 bg-blue-50"
          description="Total nominal berpindah"
        />
        <StatsCard
          label="Pengguna Terverifikasi"
          value="—"
          icon={<ShieldCheck className="h-5 w-5" />}
          colorClass="text-emerald-600 bg-emerald-50"
          description="KYC selesai"
        />
      </div>

      <p className="text-sm text-gray-400">
        Hubungkan endpoint admin ke API Gateway untuk menampilkan data nyata.
      </p>
    </div>
  );
}
