"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  ArrowDownToLine,
  ArrowRightLeft,
  ScrollText,
  UserCircle,
  Users,
  ShieldCheck,
} from "lucide-react";
import { cn } from "@/lib/cn";
import { useAuthStore } from "@/store/auth.store";

interface NavItem {
  label: string;
  href: string;
  icon: React.ElementType;
}

const userNavItems: NavItem[] = [
  { label: "Dashboard", href: "/dashboard", icon: LayoutDashboard },
  { label: "Top Up", href: "/dashboard/topup", icon: ArrowDownToLine },
  { label: "Transfer", href: "/dashboard/transfer", icon: ArrowRightLeft },
  { label: "Riwayat", href: "/dashboard/history", icon: ScrollText },
  { label: "Profil", href: "/dashboard/profile", icon: UserCircle },
];

const adminNavItems: NavItem[] = [
  { label: "Overview", href: "/admin", icon: ShieldCheck },
  { label: "Pengguna", href: "/admin/users", icon: Users },
  { label: "Transaksi", href: "/admin/transactions", icon: ScrollText },
];

interface SidebarProps {
  role?: "user" | "admin";
}

export function Sidebar({ role = "user" }: SidebarProps) {
  const pathname = usePathname();
  const navItems = role === "admin" ? adminNavItems : userNavItems;

  return (
    <aside className="flex h-full w-60 flex-shrink-0 flex-col border-r border-gray-200 bg-white">
      {/* Logo */}
      <div className="flex h-16 items-center gap-2 border-b border-gray-200 px-5">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600">
          <span className="text-sm font-bold text-white">OW</span>
        </div>
        <span className="text-base font-semibold text-gray-900">OmniWallet</span>
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto p-3">
        <ul className="flex flex-col gap-1">
          {navItems.map(({ label, href, icon: Icon }) => {
            const isActive =
              href === "/dashboard" || href === "/admin"
                ? pathname === href
                : pathname.startsWith(href);

            return (
              <li key={href}>
                <Link
                  href={href}
                  className={cn(
                    "flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-indigo-50 text-indigo-700"
                      : "text-gray-600 hover:bg-gray-50 hover:text-gray-900",
                  )}
                >
                  <Icon className="h-4 w-4 flex-shrink-0" />
                  {label}
                </Link>
              </li>
            );
          })}
        </ul>
      </nav>
    </aside>
  );
}
