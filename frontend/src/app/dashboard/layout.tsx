"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth.store";
import { DashboardLayout } from "@/components/layouts/DashboardLayout";
import { Spinner } from "@/components/ui/Spinner";
import type { ReactNode } from "react";

interface Props { children: ReactNode }

export default function UserDashboardLayout({ children }: Props) {
  const router = useRouter();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const hasHydrated = useAuthStore((s) => s._hasHydrated);

  useEffect(() => {
    // Only redirect after the store has restored state from localStorage.
    // Without this guard, the layout redirects on every hard refresh because
    // isAuthenticated is false during the brief hydration window.
    if (hasHydrated && !isAuthenticated) {
      router.replace("/login");
    }
  }, [hasHydrated, isAuthenticated, router]);

  // Show spinner while hydrating or while unauthenticated (redirect pending)
  if (!hasHydrated || !isAuthenticated) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    );
  }

  return (
    <DashboardLayout title="Dashboard" role="user">
      {children}
    </DashboardLayout>
  );
}
