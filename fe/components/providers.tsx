"use client";

import { useState } from "react";
import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore } from "@/src/stores/use-auth-store";

export function Providers({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [client] = useState(() => new QueryClient({
    defaultOptions: { queries: { staleTime: 30_000, retry: (count, error) => (error instanceof Error && "status" in error && (Number(error.status) === 0 || Number(error.status) >= 500) && count < 1) } },
  }));
  useEffect(() => {
    const onExpired = () => { useAuthStore.getState().clearSession(); client.clear(); router.replace("/login"); };
    window.addEventListener("halaltrip:auth-expired", onExpired);
    return () => window.removeEventListener("halaltrip:auth-expired", onExpired);
  }, [client, router]);
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}
