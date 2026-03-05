import { TopupForm } from "@/components/features/wallet/TopupForm";
import { BalanceCard } from "@/components/features/wallet/BalanceCard";
import type { Metadata } from "next";

export const metadata: Metadata = { title: "Top Up | OmniWallet" };

export default function TopupPage() {
  return (
    <div className="mx-auto flex max-w-lg flex-col gap-6">
      <BalanceCard />
      <TopupForm />
    </div>
  );
}
