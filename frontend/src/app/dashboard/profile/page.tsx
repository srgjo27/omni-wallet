"use client";

import { useState } from "react";
import { useAuthStore } from "@/store/auth.store";
import { useAuth } from "@/domain/use-cases/useAuth";
import { Card, CardHeader, CardTitle } from "@/components/ui/Card";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";
import { formatDate } from "@/lib/format";
import type { KycStatus } from "@/domain/models/auth.types";

const kycBadge: Record<KycStatus, { variant: "success" | "warning" | "neutral"; label: string }> = {
  VERIFIED:   { variant: "success",  label: "Terverifikasi" },
  PENDING:    { variant: "warning",  label: "Proses KYC" },
  UNVERIFIED: { variant: "neutral",  label: "Belum KYC" },
};

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user);
  const { setPin } = useAuth();

  const [pin, setPin_]       = useState("");
  const [pinConf, setPinConf] = useState("");
  const [pinLoading, setPinLoading] = useState(false);
  const [pinError, setPinError]   = useState<string | null>(null);
  const [pinSuccess, setPinSuccess] = useState<string | null>(null);

  if (!user) return null;

  const kb = kycBadge[user.kyc_status];

  const handleSetPin = async (e: React.FormEvent) => {
    e.preventDefault();
    if (pin.length !== 6) return setPinError("PIN harus 6 digit.");
    if (pin !== pinConf)  return setPinError("Konfirmasi PIN tidak cocok.");
    setPinError(null);
    setPinLoading(true);
    try {
      await setPin({ pin, confirm_pin: pinConf });
      setPinSuccess("PIN berhasil diperbarui.");
      setPin_("");
      setPinConf("");
    } catch (err: unknown) {
      setPinError(err instanceof Error ? err.message : "Gagal mengubah PIN.");
    } finally {
      setPinLoading(false);
    }
  };

  return (
    <div className="mx-auto flex max-w-2xl flex-col gap-6">
      {/* User info */}
      <Card>
        <CardHeader>
          <CardTitle>Informasi Akun</CardTitle>
        </CardHeader>
        <dl className="grid grid-cols-2 gap-x-4 gap-y-3 px-5 pb-5 text-sm">
          <div>
            <dt className="text-gray-400">Nama</dt>
            <dd className="font-medium text-gray-900">{user.name}</dd>
          </div>
          <div>
            <dt className="text-gray-400">Email</dt>
            <dd className="font-medium text-gray-900">{user.email}</dd>
          </div>
          <div>
            <dt className="text-gray-400">Status KYC</dt>
            <dd><Badge variant={kb.variant}>{kb.label}</Badge></dd>
          </div>
          <div>
            <dt className="text-gray-400">Bergabung</dt>
            <dd className="font-medium text-gray-900">{formatDate(user.created_at)}</dd>
          </div>
        </dl>
      </Card>

      {/* Set PIN */}
      <Card>
        <CardHeader>
          <CardTitle>Ubah PIN Transaksi</CardTitle>
        </CardHeader>
        <form onSubmit={handleSetPin} className="flex flex-col gap-4 px-5 pb-5">
          <Input
            label="PIN Baru (6 digit)"
            type="password"
            value={pin}
            onChange={(e) => setPin_(e.target.value.replace(/\D/g, "").slice(0, 6))}
            inputMode="numeric"
            maxLength={6}
            required
          />
          <Input
            label="Konfirmasi PIN"
            type="password"
            value={pinConf}
            onChange={(e) => setPinConf(e.target.value.replace(/\D/g, "").slice(0, 6))}
            inputMode="numeric"
            maxLength={6}
            required
          />
          {pinError   && <p className="rounded-md bg-red-50   px-3 py-2 text-sm text-red-600"   role="alert">{pinError}</p>}
          {pinSuccess && <p className="rounded-md bg-green-50 px-3 py-2 text-sm text-green-700" role="status">{pinSuccess}</p>}
          <Button type="submit" isLoading={pinLoading} className="w-full">Simpan PIN</Button>
        </form>
      </Card>
    </div>
  );
}
