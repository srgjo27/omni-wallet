"use client";

import useSWR from "swr";
import { apiClient } from "@/infrastructure/api/client";
import { authApi } from "@/infrastructure/api/auth.api";
import { UserTable } from "@/components/features/admin/UserTable";
import { Spinner } from "@/components/ui/Spinner";
import type { User } from "@/domain/models/auth.types";

interface AdminUsersResponse {
  users: User[];
  total: number;
}

async function fetcher(url: string): Promise<AdminUsersResponse> {
  const res = await apiClient.get<AdminUsersResponse>(url);
  if (!res.success || !res.data) {
    throw new Error(res.message ?? "Gagal mengambil data pengguna");
  }
  return res.data;
}

export default function AdminUsersPage() {
  const { data, isLoading, error, mutate } = useSWR("/api/v1/users", fetcher);

  const handleVerify = async (userId: string) => {
    await authApi.adminVerifyKyc(userId);
    // Revalidate the user list so the badge updates immediately
    await mutate();
  };

  return (
    <div className="flex flex-col gap-4">
      <h2 className="text-base font-semibold text-gray-700">
        Daftar Pengguna{data ? ` (${data.total})` : ""}
      </h2>

      {isLoading && (
        <div className="flex justify-center py-10">
          <Spinner size="md" />
        </div>
      )}
      {error && (
        <p className="py-6 text-center text-sm text-red-500">
          Gagal memuat data pengguna.
        </p>
      )}
      {data && <UserTable users={data.users} onVerify={handleVerify} />}
    </div>
  );
}
