"use client";

import { LogOut, ChevronDown } from "lucide-react";
import { useAuth } from "@/domain/use-cases/useAuth";
import { Button } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";
import type { KycStatus } from "@/domain/models/auth.types";

const kycBadgeMap: Record<KycStatus, { variant: "success" | "warning" | "neutral"; label: string }> = {
  VERIFIED:   { variant: "success",  label: "Terverifikasi" },
  PENDING:    { variant: "warning",  label: "Proses KYC" },
  UNVERIFIED: { variant: "neutral",  label: "Belum KYC" },
};

interface HeaderProps {
  title: string;
}

export function Header({ title }: HeaderProps) {
  const { user, logout } = useAuth();
  const kyc = user ? kycBadgeMap[user.kyc_status] : null;

  return (
    <header className="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-6">
      <h1 className="text-lg font-semibold text-gray-900">{title}</h1>

      <div className="flex items-center gap-3">
        {kyc && <Badge variant={kyc.variant}>{kyc.label}</Badge>}

        {user && (
          <div className="flex items-center gap-2">
            <div className="flex h-8 w-8 items-center justify-center rounded-full bg-indigo-100">
              <span className="text-sm font-semibold text-indigo-700">
                {user.name.charAt(0).toUpperCase()}
              </span>
            </div>
            <span className="hidden text-sm font-medium text-gray-700 sm:block">
              {user.name}
            </span>
          </div>
        )}

        <Button
          variant="ghost"
          size="sm"
          onClick={logout}
          className="gap-1.5 text-gray-500 hover:text-red-600"
          aria-label="Logout"
        >
          <LogOut className="h-4 w-4" />
          <span className="hidden sm:block">Keluar</span>
        </Button>
      </div>
    </header>
  );
}
