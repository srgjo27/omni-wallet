"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";
import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { Spinner } from "@/components/ui/Spinner";
import type { ReactNode } from "react";

interface Props { children: ReactNode }

export default function AdminLayout({ children }: Props) {
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const user = useAuthStore((s) => s.user);
  const hasHydrated = useAuthStore((s) => s._hasHydrated);

  useEffect(() => {
    if (!hasHydrated) return; // Wait for localStorage restore before acting
    if (!isAuthenticated) {
      router.replace("/login");
    } else if (user && !user.email.endsWith("@admin.omniwallet")) {
      router.replace("/dashboard");
    }
  }, [hasHydrated, isAuthenticated, user, router]);

  if (!hasHydrated || !isAuthenticated || !user?.email.endsWith("@admin.omniwallet")) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    );
  }

  return (
    <DashboardLayout title="Admin Panel" role="admin">
      {children}
    </DashboardLayout>
  );
}
