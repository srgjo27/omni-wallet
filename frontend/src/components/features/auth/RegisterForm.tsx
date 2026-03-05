"use client";

import { useState } from "react";
import Link from "next/link";
import { useAuth } from "@/domain/use-cases/useAuth";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/ui/Input";

export function RegisterForm() {
  const { register } = useAuth();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsLoading(true);
    try {
      await register({ name, email, password });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Registrasi gagal. Coba lagi.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4">
      <Input
        label="Nama Lengkap"
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Budi Santoso"
        autoComplete="name"
        required
      />

      <Input
        label="Email"
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="kamu@contoh.com"
        autoComplete="email"
        required
      />

      <Input
        label="Password"
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Min. 8 karakter"
        autoComplete="new-password"
        minLength={8}
        required
      />

      {error && (
        <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600" role="alert">
          {error}
        </p>
      )}

      <Button type="submit" isLoading={isLoading} className="mt-1 w-full">
        Buat Akun
      </Button>

      <p className="text-center text-sm text-gray-500">
        Sudah punya akun?{" "}
        <Link href="/login" className="font-medium text-indigo-600 hover:underline">
          Masuk
        </Link>
      </p>
    </form>
  );
}
