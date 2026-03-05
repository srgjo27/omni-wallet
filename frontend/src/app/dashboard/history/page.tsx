"use client";

import { useState } from "react";
import { MutationList } from "@/components/features/wallet/MutationList";
import { TransactionList } from "@/components/features/wallet/TransactionList";
import { cn } from "@/lib/cn";

const tabs = [
  { id: "mutations",    label: "Mutasi Saldo" },
  { id: "transactions", label: "Riwayat Transaksi" },
] as const;

type TabId = (typeof tabs)[number]["id"];

export default function HistoryPage() {
  const [active, setActive] = useState<TabId>("mutations");

  return (
    <div className="flex flex-col gap-4">
      {/* Tab bar */}
      <div className="flex gap-1 rounded-xl border border-gray-200 bg-gray-100 p-1 w-fit">
        {tabs.map(({ id, label }) => (
          <button
            key={id}
            onClick={() => setActive(id)}
            className={cn(
              "rounded-lg px-4 py-1.5 text-sm font-medium transition-colors",
              active === id
                ? "bg-white text-gray-900 shadow-sm"
                : "text-gray-500 hover:text-gray-700",
            )}
          >
            {label}
          </button>
        ))}
      </div>

      {active === "mutations" ? <MutationList /> : <TransactionList />}
    </div>
  );
}
