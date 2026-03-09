"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { formatDate, shortId } from "@/lib/format";
import type { User } from "@/domain/models/auth.types";
import type { KycStatus } from "@/domain/models/auth.types";

const kycBadge: Record<KycStatus, { variant: "success" | "warning" | "neutral"; label: string }> = {
  VERIFIED:   { variant: "success",  label: "Terverifikasi" },
  PENDING:    { variant: "warning",  label: "Proses KYC" },
  UNVERIFIED: { variant: "neutral",  label: "Belum KYC" },
};

interface UserTableProps {
  users: User[];
  onVerify?: (userId: string) => Promise<void>;
}

export function UserTable({ users, onVerify }: UserTableProps) {
  const [loadingId, setLoadingId] = useState<string | null>(null);

  if (!users.length) {
    return <p className="py-6 text-center text-sm text-gray-500">Tidak ada pengguna ditemukan.</p>;
  }

  const handleVerify = async (userId: string) => {
    if (!onVerify) return;
    setLoadingId(userId);
    try {
      await onVerify(userId);
    } finally {
      setLoadingId(null);
    }
  };

  return (
    <div className="overflow-x-auto rounded-xl border border-gray-200">
      <table className="min-w-full divide-y divide-gray-100 bg-white">
        <thead>
          <tr className="text-left text-xs font-medium uppercase tracking-wide text-gray-400">
            <th className="px-4 py-3">Nama</th>
            <th className="px-4 py-3">Email</th>
            <th className="px-4 py-3">ID</th>
            <th className="px-4 py-3">KYC</th>
            <th className="px-4 py-3">Bergabung</th>
            {onVerify && <th className="px-4 py-3">Aksi</th>}
          </tr>
        </thead>
        <tbody className="divide-y divide-gray-50">
          {users.map((user) => {
            const kb = kycBadge[user.kyc_status];
            return (
              <tr key={user.id} className="text-sm text-gray-700 hover:bg-gray-50">
                <td className="px-4 py-3 font-medium">{user.name}</td>
                <td className="px-4 py-3 text-gray-500">{user.email}</td>
                <td className="px-4 py-3 font-mono text-xs text-gray-400">#{shortId(user.id)}</td>
                <td className="px-4 py-3">
                  <Badge variant={kb.variant}>{kb.label}</Badge>
                </td>
                <td className="px-4 py-3 text-xs text-gray-400">{formatDate(user.created_at)}</td>
                {onVerify && (
                  <td className="px-4 py-3">
                    {user.kyc_status === "PENDING" && (
                      <Button
                        size="sm"
                        variant="secondary"
                        isLoading={loadingId === user.id}
                        disabled={loadingId !== null}
                        onClick={() => handleVerify(user.id)}
                      >
                        Verifikasi
                      </Button>
                    )}
                  </td>
                )}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
