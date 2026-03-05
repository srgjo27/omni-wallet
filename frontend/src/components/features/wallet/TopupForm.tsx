"use client";

import { useState } from "react";
import { useTransfer } from "@/domain/use-cases/useTransfer";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { formatCurrency } from "@/lib/format";

const quickAmounts = [50_000, 100_000, 200_000, 500_000];

export function TopupForm() {
  const { topup } = useTransfer();
  const [amount, setAmount] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleTopup = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    const numAmount = Number(amount);
    if (!numAmount || numAmount < 10_000) {
      setError("Minimum top-up adalah Rp10.000");
      return;
    }
    setIsLoading(true);
    try {
      await topup(numAmount);
      setSuccess(`Top-up ${formatCurrency(numAmount)} berhasil!`);
      setAmount("");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Top-up gagal.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Top Up Saldo</CardTitle>
      </CardHeader>

      <form onSubmit={handleTopup} className="flex flex-col gap-4 px-5 pb-5">
        {/* Quick select */}
        <div className="grid grid-cols-4 gap-2">
          {quickAmounts.map((q) => (
            <button
              key={q}
              type="button"
              onClick={() => setAmount(String(q))}
              className="rounded-lg border border-gray-200 px-2 py-1.5 text-center text-sm font-medium text-gray-700 transition hover:border-indigo-400 hover:text-indigo-700"
            >
              {formatCurrency(q)}
            </button>
          ))}
        </div>

        <Input
          label="Jumlah (Rp)"
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0"
          min={10000}
          hint="Minimum Rp10.000"
          required
        />

        {error && (
          <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600" role="alert">
            {error}
          </p>
        )}
        {success && (
          <p className="rounded-md bg-green-50 px-3 py-2 text-sm text-green-700" role="status">
            {success}
          </p>
        )}

        <Button type="submit" isLoading={isLoading} className="w-full">
          Top Up Sekarang
        </Button>
      </form>
    </Card>
  );
}
