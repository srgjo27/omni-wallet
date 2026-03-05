import type { ReactNode } from "react";
import { Card } from "@/components/ui/Card";

interface StatsCardProps {
  label: string;
  value: string | number;
  icon?: ReactNode;
  description?: string;
  colorClass?: string; // e.g. "text-indigo-600 bg-indigo-50"
}

export function StatsCard({
  label,
  value,
  icon,
  description,
  colorClass = "text-indigo-600 bg-indigo-50",
}: StatsCardProps) {
  return (
    <Card className="flex items-start gap-4 p-5">
      {icon && (
        <div className={`flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg ${colorClass}`}>
          {icon}
        </div>
      )}
      <div className="min-w-0">
        <p className="text-xs font-medium uppercase tracking-wide text-gray-400">{label}</p>
        <p className="mt-1 text-2xl font-bold text-gray-900">{value}</p>
        {description && <p className="mt-0.5 text-xs text-gray-400">{description}</p>}
      </div>
    </Card>
  );
}
