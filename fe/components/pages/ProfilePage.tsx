"use client";

import Link from "next/link";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getMe, getPreferences, savePreferences, updateMe } from "@/src/services";
import { APIError } from "@/src/services/client";
import { useAuthStore } from "@/src/stores/use-auth-store";
import type { Activity, Interest, Preferences, TravelType } from "@/src/types/api";
import { AuthPage } from "./AuthPage";
import { PageIntro } from "./PageIntro";

const interests: { id: Interest; label: string }[] = [
  { id: "culinary", label: "Kuliner" }, { id: "culture", label: "Budaya" },
  { id: "nature", label: "Alam" }, { id: "shopping", label: "Belanja" },
  { id: "adventure", label: "Petualangan" }, { id: "history", label: "Sejarah" },
];
const activities: Activity[] = ["Relaxing", "Exploration", "Adventure"];
const travelTypes: TravelType[] = ["Solo", "Couple", "Family", "Friends"];
const labels: Record<string, string> = { Relaxing: "Santai", Exploration: "Eksplorasi", Adventure: "Petualangan", Solo: "Sendiri", Couple: "Berdua", Family: "Keluarga", Friends: "Teman" };

export function Profile() {
  const token = useAuthStore((state) => state.token);
  const sessionUser = useAuthStore((state) => state.user);
  const setUser = useAuthStore((state) => state.setUser);
  const clearSession = useAuthStore((state) => state.clearSession);
  const queryClient = useQueryClient();
  const [name, setName] = useState(sessionUser?.name ?? "");
  const [email, setEmail] = useState(sessionUser?.email ?? "");
  const [preferencesDraft, setPreferencesDraft] = useState<Preferences | undefined>();

  const profileQuery = useQuery({ queryKey: ["me", sessionUser?.id], queryFn: () => getMe(token!), enabled: Boolean(token) });
  const preferencesQuery = useQuery({
    queryKey: ["preferences", sessionUser?.id], queryFn: () => getPreferences(token!), enabled: Boolean(token), retry: false,
  });
  const preferences: Preferences = preferencesDraft ?? preferencesQuery.data ?? { interests: [], activity: "Exploration", travel_type: "Solo" };

  const profileMutation = useMutation({
    mutationFn: () => updateMe(token!, { name: name.trim(), email: email.trim() }),
    onSuccess: (user) => { setUser(user); void queryClient.invalidateQueries({ queryKey: ["me", user.id] }); },
  });
  const preferencesMutation = useMutation({
    mutationFn: () => savePreferences(token!, preferences),
    onSuccess: async (saved) => { setPreferencesDraft(saved); await queryClient.setQueryData(["preferences", sessionUser?.id], saved); },
  });

  if (!token) return <AuthPage />;
  const preferencesMissing = preferencesQuery.error instanceof APIError && preferencesQuery.error.status === 404;
  const profileError = profileQuery.error instanceof Error ? profileQuery.error.message : "";

  return <>
    <PageIntro eyebrow="A LITTLE ABOUT YOU" title="Travel, your way." description="Perbarui akun dan preferensi yang digunakan layanan rekomendasi HalalTrip." />
    {profileQuery.isLoading ? <p aria-live="polite">Memuat profil...</p> : profileQuery.isError ? <div className="empty-state"><h3>Profil belum dapat dimuat</h3><p>{profileError}</p><button className="button outline" onClick={() => void profileQuery.refetch()}>Coba lagi</button></div> : <div className="profile-layout">
      <aside className="profile-card">
        <div className="profile-avatar">{(sessionUser?.name ?? name).slice(0, 1).toUpperCase()}</div>
        <h2>{profileQuery.data?.name ?? sessionUser?.name}</h2><p>{profileQuery.data?.email ?? sessionUser?.email}</p>
        <button className="button outline" onClick={() => { clearSession(); queryClient.clear(); }}>Keluar</button>
      </aside>
      <div className="profile-preferences">
        <section>
          <div className="eyebrow">ACCOUNT</div><h2>Informasi akun</h2>
          <label className="form-label">Nama<input className="form-input" value={name} onChange={(event) => setName(event.target.value)} maxLength={100} /></label>
          <label className="form-label">Email<input className="form-input" type="email" value={email} onChange={(event) => setEmail(event.target.value)} maxLength={254} /></label>
          <button className="button primary" disabled={profileMutation.isPending || !name.trim() || !email.includes("@")} onClick={() => profileMutation.mutate()}>{profileMutation.isPending ? "Menyimpan..." : "Simpan profil"}</button>
          {profileMutation.isError && <p className="error-text" role="alert">{profileMutation.error.message}</p>}
          {profileMutation.isSuccess && <p role="status">Profil berhasil diperbarui.</p>}
        </section>
        <section>
          <div className="eyebrow">WHAT MATTERS MOST</div><h2>Minat perjalanan</h2>
          {preferencesMissing && <p>Preferensi belum tersimpan. Pilih setidaknya satu minat lalu simpan.</p>}
          {preferencesQuery.isError && !preferencesMissing && <div className="error-text">{preferencesQuery.error.message} <button className="button outline" onClick={() => void preferencesQuery.refetch()}>Coba lagi</button></div>}
          <div className="preference-options">{interests.map(({ id, label }) => <label className="choice" key={id}><input type="checkbox" checked={preferences.interests.includes(id)} onChange={(event) => setPreferencesDraft({ ...preferences, interests: event.target.checked ? [...preferences.interests, id] : preferences.interests.filter((item) => item !== id) })} /><span>{label}</span></label>)}</div>
          <label className="form-label">Aktivitas<select className="form-input" value={preferences.activity} onChange={(event) => setPreferencesDraft({ ...preferences, activity: event.target.value as Activity })}>{activities.map((item) => <option key={item} value={item}>{labels[item]}</option>)}</select></label>
          <label className="form-label">Tipe perjalanan<select className="form-input" value={preferences.travel_type} onChange={(event) => setPreferencesDraft({ ...preferences, travel_type: event.target.value as TravelType })}>{travelTypes.map((item) => <option key={item} value={item}>{labels[item]}</option>)}</select></label>
          <button className="button primary" disabled={preferencesMutation.isPending || preferences.interests.length === 0} onClick={() => preferencesMutation.mutate()}>{preferencesMutation.isPending ? "Menyimpan..." : "Simpan preferensi"}</button>
          {preferencesMutation.isError && <p className="error-text" role="alert">{preferencesMutation.error.message}</p>}
          {preferencesMutation.isSuccess && <p role="status">Preferensi tersimpan.</p>}
        </section>
        <p>Access token hanya disimpan sementara di memori browser. Muat ulang halaman akan mengakhiri sesi; backend saat ini belum mengeluarkan cookie HttpOnly.</p>
        <Link href="/trips" className="text-link">Lihat perjalanan saya</Link>
      </div>
    </div>}
  </>;
}
