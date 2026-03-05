import { Sidebar } from "./Sidebar";
import { Header } from "./Header";
import type { ReactNode } from "react";

interface DashboardLayoutProps {
  children: ReactNode;
  title: string;
  role?: "user" | "admin";
}

/**
 * Shared two-column shell for both the User and Admin dashboards.
 * Sidebar is on the left, header + content on the right.
 */
export function DashboardLayout({ children, title, role = "user" }: DashboardLayoutProps) {
  return (
    <div className="flex h-screen overflow-hidden bg-gray-50">
      <Sidebar role={role} />

      <div className="flex flex-1 flex-col overflow-hidden">
        <Header title={title} />

        <main className="flex-1 overflow-y-auto p-6">
          {children}
        </main>
      </div>
    </div>
  );
}
