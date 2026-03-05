import type { Metadata } from "next";
import { LoginForm } from "@/components/features/auth/LoginForm";

export const metadata: Metadata = { title: "Masuk | OmniWallet" };

export default function LoginPage() {
  return (
    <>
      <h2 className="mb-6 text-center text-xl font-bold text-gray-900">Masuk ke Akun</h2>
      <LoginForm />
    </>
  );
}
