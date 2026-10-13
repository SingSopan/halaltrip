import type { Metadata } from "next";
import "./globals.css";
import "./integration.css";
import { Providers } from "@/components/providers";

export const metadata: Metadata = {
  title: "HalalTrip | Muslim-Friendly Travel Planner",
  description: "Discover Muslim-friendly destinations, halal food, prayer facilities, and plan your perfect trip.",
  openGraph: { title: "HalalTrip | Travel the World, Stay Halal.", description: "Discover and plan Muslim-friendly trips around the world." },
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="en" className="h-full antialiased">
      <body className="min-h-full flex flex-col"><Providers>{children}</Providers></body>
    </html>
  );
}
