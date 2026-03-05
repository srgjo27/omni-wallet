// Base HTTP client for all API calls.
// All requests are routed through the Next.js rewrite proxy → API Gateway → microservices.
//
// IMPORTANT: base URL must be "" (empty) so browser requests go to relative paths like
// /api/v1/... — Next.js intercepts them server-side and rewrites to the actual API Gateway
// URL (configured in next.config.ts). Using NEXT_PUBLIC_API_URL here would embed the
// Docker-internal hostname (api-gateway:8080) into the JS bundle, causing ERR_NAME_NOT_RESOLVED
// in the browser.
const API_BASE_URL = "";

export interface ApiResponse<T = unknown> {
  success: boolean;
  message: string;
  data?: T;
  error?: unknown;
}

class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public data?: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * Reads the JWT token from localStorage.
 * Returns null when called server-side (no window object).
 */
function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("omni_token");
}

/**
 * Core fetch wrapper.
 * Automatically attaches the Bearer token and deserialises the response.
 * Throws ApiError for non-2xx responses.
 */
async function request<T>(
  path: string,
  options: RequestInit = {},
): Promise<ApiResponse<T>> {
  const token = getToken();

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...(options.headers ?? {}),
  };

  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers,
  });

  const json: ApiResponse<T> = await res.json();

  if (!res.ok) {
    throw new ApiError(res.status, json.message ?? "Request failed", json.error);
  }

  return json;
}

export const apiClient = {
  get: <T>(path: string) => request<T>(path, { method: "GET" }),

  post: <T>(path: string, body: unknown) =>
    request<T>(path, {
      method: "POST",
      body: JSON.stringify(body),
    }),

  put: <T>(path: string, body: unknown) =>
    request<T>(path, {
      method: "PUT",
      body: JSON.stringify(body),
    }),

  delete: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};

export { ApiError };
