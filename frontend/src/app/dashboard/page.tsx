import { BalanceCard } from "@/components/features/wallet/BalanceCard";
import { TransactionList } from "@/components/features/wallet/TransactionList";
import type { Metadata } from "next";

export const metadata: Metadata = { title: "Dashboard | OmniWallet" };

export default function DashboardPage() {
  return (
    <div className="flex flex-col gap-6">
      <BalanceCard />

      <section>
        <h2 className="mb-3 text-base font-semibold text-gray-700">Transaksi Terbaru</h2>
        <TransactionList />
      </section>
    </div>
  );
}
