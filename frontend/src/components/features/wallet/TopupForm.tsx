"use client";

import { useState } from "react";
import { useTransfer } from "@/domain/use-cases/useTransfer";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { formatCurrency } from "@/lib/format";
import type { VirtualAccountResponse } from "@/domain/models/transaction.types";

const BANKS = [
  { code: "BNI", label: "Bank BNI" },
  { code: "MANDIRI", label: "Bank Mandiri" },
  { code: "BRI", label: "Bank BRI" },
  { code: "PERMATA", label: "Bank Permata" },
  { code: "BSI", label: "Bank BSI" },
] as const;

const quickAmounts = [50_000, 100_000, 200_000, 500_000];

export function TopupForm() {
  const { requestVA, simulatePayment, isLoading, error, clearError } = useTransfer();

  const [step, setStep] = useState<1 | 2>(1);
  const [selectedBank, setSelectedBank] = useState<string>("");
  const [va, setVA] = useState<VirtualAccountResponse | null>(null);
  const [simAmount, setSimAmount] = useState("");
  const [simSuccess, setSimSuccess] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleRequestVA = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();
    setSimSuccess(null);
    if (!selectedBank) return;
    const result = await requestVA(selectedBank);
    if (result) {
      setVA(result);
      setStep(2);
    }
  };

  const handleSimulate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!va) return;
    clearError();
    setSimSuccess(null);
    const numAmount = Number(simAmount);
    if (!numAmount || numAmount < 10_000) return;
    const tx = await simulatePayment(va, numAmount);
    if (tx) {
      setSimSuccess(`Simulasi ${formatCurrency(numAmount)} berhasil! Saldo akan diperbarui.`);
      setSimAmount("");
    }
  };

  const handleCopy = async () => {
    if (!va) return;
    await navigator.clipboard.writeText(va.account_number);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleReset = () => {
    setStep(1);
    setVA(null);
    setSelectedBank("");
    setSimSuccess(null);
    clearError();
  };

  if (step === 1) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Top Up via Virtual Account</CardTitle>
        </CardHeader>

        <form onSubmit={handleRequestVA} className="flex flex-col gap-4 px-5 pb-5">
          <p className="text-sm text-gray-500">
            Pilih bank untuk mendapatkan nomor Virtual Account (VA) kamu.
            VA bersifat permanen — kamu bisa top-up berulang kali ke nomor yang sama.
          </p>

          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {BANKS.map(({ code, label }) => (
              <button
                key={code}
                type="button"
                onClick={() => setSelectedBank(code)}
                className={`rounded-xl border-2 px-4 py-3 text-left text-sm font-medium transition ${
                  selectedBank === code
                    ? "border-indigo-500 bg-indigo-50 text-indigo-700"
                    : "border-gray-200 text-gray-700 hover:border-indigo-300"
                }`}
              >
                {label}
              </button>
            ))}
          </div>

          {error && (
            <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600" role="alert">
              {error}
            </p>
          )}

          <Button
            type="submit"
            isLoading={isLoading}
            disabled={!selectedBank || isLoading}
            className="w-full"
          >
            Dapatkan Nomor Virtual Account
          </Button>
        </form>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Nomor Virtual Account</CardTitle>
      </CardHeader>

      <div className="flex flex-col gap-5 px-5 pb-5">
        {va && (
          <div className="rounded-xl bg-indigo-50 p-4">
            <p className="mb-1 text-xs font-semibold uppercase tracking-wide text-indigo-500">
              {va.bank_code} — {va.merchant_code}
            </p>
            <div className="flex items-center gap-2">
              <span className="text-2xl font-bold tracking-widest text-indigo-800">
                {va.account_number}
              </span>
              <button
                type="button"
                onClick={handleCopy}
                className="rounded-md border border-indigo-300 px-2 py-0.5 text-xs text-indigo-600 hover:bg-indigo-100"
              >
                {copied ? "Tersalin!" : "Salin"}
              </button>
            </div>
            <p className="mt-1 text-xs text-indigo-600">
              Atas nama: <span className="font-medium">{va.name}</span>
            </p>
            <p className="mt-1 text-xs text-gray-500">
              Lakukan transfer ke nomor VA di atas melalui ATM / Mobile Banking.
            </p>
          </div>
        )}

        <div className="rounded-xl border border-dashed border-amber-300 bg-amber-50 p-4">
          <p className="mb-3 text-xs font-semibold text-amber-700">
            Mode Test — Simulasikan Pembayaran
          </p>

          <form onSubmit={handleSimulate} className="flex flex-col gap-3">
            <div className="grid grid-cols-4 gap-2">
              {quickAmounts.map((q) => (
                <button
                  key={q}
                  type="button"
                  onClick={() => setSimAmount(String(q))}
                  className="rounded-lg border border-amber-200 px-2 py-1.5 text-center text-xs font-medium text-amber-700 transition hover:border-amber-400"
                >
                  {formatCurrency(q)}
                </button>
              ))}
            </div>

            <Input
              label="Jumlah Simulasi (Rp)"
              type="number"
              value={simAmount}
              onChange={(e) => setSimAmount(e.target.value)}
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
            {simSuccess && (
              <p className="rounded-md bg-green-50 px-3 py-2 text-sm text-green-700" role="status">
                {simSuccess}
              </p>
            )}

            <Button
              type="submit"
              isLoading={isLoading}
              disabled={isLoading || !simAmount}
              className="w-full"
            >
              Simulasikan Pembayaran
            </Button>
          </form>
        </div>

        <button
          type="button"
          onClick={handleReset}
          className="text-center text-sm text-indigo-600 hover:underline"
        >
          ← Pilih bank lain
        </button>
      </div>
    </Card>
  );
}
