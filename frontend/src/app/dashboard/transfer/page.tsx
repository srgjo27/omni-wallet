import { TransferForm } from "@/components/features/wallet/TransferForm";
import { BalanceCard } from "@/components/features/wallet/BalanceCard";
import type { Metadata } from "next";

export const metadata: Metadata = { title: "Transfer | OmniWallet" };

export default function TransferPage() {
  return (
    <div className="mx-auto flex max-w-lg flex-col gap-6">
      <BalanceCard />
      <TransferForm />
    </div>
  );
}
