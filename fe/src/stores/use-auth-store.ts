"use client";

import { create } from "zustand";
import type { User } from "@/src/types/api";

type AuthState = {
  token: string | null;
  user: User | null;
  setSession: (token: string, user: User) => void;
  setUser: (user: User) => void;
  clearSession: () => void;
};

/** Access token stays in memory because the backend currently returns a JSON JWT, not an HttpOnly cookie. */
export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  user: null,
  setSession: (token, user) => set({ token, user }),
  setUser: (user) => set({ user }),
  clearSession: () => set({ token: null, user: null }),
}));
