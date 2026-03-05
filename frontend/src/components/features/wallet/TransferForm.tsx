"use client";

import { useState } from "react";
import { useTransfer } from "@/domain/use-cases/useTransfer";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { formatCurrency } from "@/lib/format";

export function TransferForm() {
  const { transfer } = useTransfer();
  const [targetUserId, setTargetUserId] = useState("");
  const [amount, setAmount] = useState("");
  const [pin, setPin] = useState("");
  const [description, setDescription] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleTransfer = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    const numAmount = Number(amount);
    if (!targetUserId.trim()) return setError("ID penerima wajib diisi.");
    if (!numAmount || numAmount < 1_000) return setError("Minimum transfer adalah Rp1.000");
    if (pin.length !== 6) return setError("PIN harus 6 digit.");

    setIsLoading(true);
    try {
      await transfer(targetUserId.trim(), numAmount, pin, description);
      setSuccess(`Transfer ${formatCurrency(numAmount)} berhasil!`);
      setTargetUserId("");
      setAmount("");
      setPin("");
      setDescription("");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Transfer gagal.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Transfer ke Pengguna</CardTitle>
      </CardHeader>

      <form onSubmit={handleTransfer} className="flex flex-col gap-4 px-5 pb-5">
        <Input
          label="ID Pengguna Tujuan"
          type="text"
          value={targetUserId}
          onChange={(e) => setTargetUserId(e.target.value)}
          placeholder="user-uuid-..."
          hint="Masukkan UUID akun tujuan"
          required
        />

        <Input
          label="Jumlah (Rp)"
          type="number"
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          placeholder="0"
          min={1000}
          hint="Minimum Rp1.000"
          required
        />

        <Input
          label="Catatan (opsional)"
          type="text"
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Pembayaran tagihan..."
        />

        <Input
          label="PIN Transaksi (6 digit)"
          type="password"
          value={pin}
          onChange={(e) => setPin(e.target.value.replace(/\D/g, "").slice(0, 6))}
          placeholder="••••••"
          inputMode="numeric"
          maxLength={6}
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
          Kirim Transfer
        </Button>
      </form>
    </Card>
  );
}
