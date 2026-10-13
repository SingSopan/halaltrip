"use client";

import Link from "next/link";
import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { PlaceCard } from "@/components/common";
import { getSavedPlaces } from "@/src/services";
import { useAuthStore } from "@/src/stores/use-auth-store";
import { PageIntro } from "./PageIntro";
import { SAVED_TABS } from "./page-data";

const categoryForTab: Record<string, string> = { Destinations: "Destination", Restaurants: "Restaurant", Hotels: "Hotel", Mosques: "Mosque" };

export function SavedPage() {
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const [tab, setTab] = useState("All");
  const query = useQuery({ queryKey: ["savedPlaces", user?.id], queryFn: () => getSavedPlaces(token!), enabled: Boolean(token && user) });
  const saved = query.data ?? [];
  const filtered = saved.filter((entry) => tab === "All" || entry.place.category === categoryForTab[tab]);

  return <>
    <PageIntro eyebrow="YOUR LITTLE BLACK BOOK" title="Tempat tersimpan." description="Daftar ini berasal dari akun Anda dan disimpan oleh backend HalalTrip." />
    {!token ? <div className="empty-state"><h3>Masuk untuk melihat tempat tersimpan</h3><p>Saved places terhubung ke akun HalalTrip.</p><Link href="/profile" className="button primary">Masuk atau daftar</Link></div> : query.isLoading ? <p aria-live="polite">Memuat tempat tersimpan...</p> : query.isError ? <div className="empty-state"><h3>Tempat tersimpan gagal dimuat</h3><p>{query.error.message}</p><button className="button outline" onClick={() => void query.refetch()}>Coba lagi</button></div> : <>
      <div className="tabs saved-tabs">{SAVED_TABS.map((tabName) => <button key={tabName} type="button" className={tab === tabName ? "tab active" : "tab"} onClick={() => setTab(tabName)}>{tabName}</button>)}</div>
      {filtered.length ? <div className="destination-grid">{filtered.map(({ place }) => <PlaceCard key={`${place.category}-${place.id}`} place={place} />)}</div> : <div className="empty-state"><span aria-hidden="true">♡</span><h3>{saved.length ? "Tidak ada tempat pada kategori ini." : "Belum ada tempat tersimpan."}</h3><p>Simpan restoran, hotel, atau masjid lokal dari halaman katalog.</p><Link href="/restaurants" className="button primary">Jelajahi katalog</Link></div>}
    </>}
  </>;
}
