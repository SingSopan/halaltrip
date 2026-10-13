import { apiRequest } from "./client";
import type { AuthResponse, Preferences, User } from "@/src/types/api";

export function register(input: { name: string; email: string; password: string }) {
  return apiRequest<AuthResponse>("/auth/register", { method: "POST", body: input });
}

export function login(input: { email: string; password: string }) {
  return apiRequest<AuthResponse>("/auth/login", { method: "POST", body: input });
}

export function getMe(token: string) {
  return apiRequest<{ user: User }>("/users/me", { token }).then((result) => result.user);
}

export function updateMe(token: string, input: { name?: string; email?: string }) {
  return apiRequest<{ user: User }>("/users/me", { method: "PATCH", token, body: input }).then((result) => result.user);
}

export function getPreferences(token: string) {
  return apiRequest<{ preferences: Preferences }>("/users/me/preferences", { token }).then((result) => result.preferences);
}

export function savePreferences(token: string, preferences: Preferences) {
  return apiRequest<{ preferences: Preferences }>("/users/me/preferences", { method: "PATCH", token, body: preferences }).then((result) => result.preferences);
}
