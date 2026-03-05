import { create } from "zustand";
import { persist } from "zustand/middleware";
import type { User } from "@/domain/models/auth.types";

interface AuthState {
  token: string | null;
  user: User | null;
  isAuthenticated: boolean;
  /** True once the persist middleware has finished reading from localStorage. */
  _hasHydrated: boolean;
  setAuth: (token: string, user: User) => void;
  clearAuth: () => void;
  updateUser: (user: Partial<User>) => void;
  _setHasHydrated: (value: boolean) => void;
}

/**
 * Global authentication store persisted to localStorage.
 * The token is also written to localStorage["omni_token"] so the
 * API client (which runs outside React) can read it synchronously.
 */
export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      isAuthenticated: false,
      _hasHydrated: false,

      setAuth: (token, user) => {
        // Write to a plain key so the infrastructure api/client.ts can read it
        // without importing from this store (infrastructure ≠ store dependency).
        // Guard: only write if token is a real non-empty string.
        if (typeof window !== "undefined" && token) {
          localStorage.setItem("omni_token", token);
        }
        set({ token, user, isAuthenticated: true });
      },

      clearAuth: () => {
        if (typeof window !== "undefined") {
          localStorage.removeItem("omni_token");
        }
        set({ token: null, user: null, isAuthenticated: false });
      },

      updateUser: (partial) =>
        set((state) => ({
          user: state.user ? { ...state.user, ...partial } : null,
        })),

      _setHasHydrated: (value) => set({ _hasHydrated: value }),
    }),
    {
      name: "omni-auth",
      // Persist token, user, AND isAuthenticated so hard refresh restores full state.
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        isAuthenticated: state.isAuthenticated,
      }),
      onRehydrateStorage: () => (state) => {
        // Re-sync the plain localStorage key after hydration so the API client
        // is consistent on hard refresh even before React trees mount.
        if (state?.token && typeof window !== "undefined") {
          localStorage.setItem("omni_token", state.token);
        }
        // Signal that hydration is done — layouts wait for this before redirecting.
        state?._setHasHydrated(true);
      },
    },
  ),
);
