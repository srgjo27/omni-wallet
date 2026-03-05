import type { Metadata } from "next";
import { RegisterForm } from "@/components/features/auth/RegisterForm";

export const metadata: Metadata = { title: "Daftar | OmniWallet" };

export default function RegisterPage() {
  return (
    <>
      <h2 className="mb-6 text-center text-xl font-bold text-gray-900">Buat Akun Baru</h2>
      <RegisterForm />
    </>
  );
}
