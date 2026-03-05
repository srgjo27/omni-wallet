/**
 * Formats a raw integer amount (stored in smallest currency unit, e.g. cents)
 * into a human-readable Rupiah string.
 * Example: 1500000 → "Rp 1.500.000"
 */
export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

/**
 * Formats an ISO date string into a localised Indonesian date-time.
 * Example: "2026-03-04T07:00:00Z" → "4 Mar 2026, 14.00"
 */
export function formatDate(iso: string): string {
  return new Intl.DateTimeFormat("id-ID", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(iso));
}

/**
 * Truncates a UUID to a short reference display.
 * Example: "abc123de-..." → "abc123de"
 */
export function shortId(id: string): string {
  return id.split("-")[0].toUpperCase();
}
