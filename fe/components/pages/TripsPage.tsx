"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { ArrowLeft, ArrowRight, CalendarDays, Plus } from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { EmptyState } from "@/components/common";
import { TripCard } from "@/components/trip/TripCard";
import { useAuthStore } from "@/src/stores/use-auth-store";
import { deleteTrip as removeTrip, getTrip, listTrips, updateTrip } from "@/src/services";
import type { TripStatus } from "@/src/types/api";
import { PageIntro } from "./PageIntro";
import { NotFoundPage as NotFound } from "./NotFoundPage";

export function TripsPage() {
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();
  const tripsQuery = useQuery({ queryKey: ["trips", user?.id], queryFn: () => listTrips(token!), enabled: Boolean(token && user) });
  const deleteMutation = useMutation({
    mutationFn: (id: string) => removeTrip(token!, id),
    onSuccess: async () => queryClient.invalidateQueries({ queryKey: ["trips", user?.id] }),
  });

  const trips = tripsQuery.data ?? [];
  return <>
    <PageIntro eyebrow="YOUR TRAVEL NOTEBOOK" title="Your trips." description="Perjalanan tersimpan di akun dan hanya dapat diakses oleh pemiliknya." />
    {!token ? <EmptyState title="Masuk untuk melihat perjalanan" description="Trip disimpan aman pada akun HalalTrip." action={<Link href="/profile" className="button primary">Masuk atau daftar</Link>} /> : tripsQuery.isLoading ? <p aria-live="polite">Memuat perjalanan...</p> : tripsQuery.isError ? <div className="empty-state"><h3>Perjalanan belum dapat dimuat</h3><p>{tripsQuery.error.message}</p><button className="button outline" onClick={() => void tripsQuery.refetch()}>Coba lagi</button></div> : <>
      <div className="trips-toolbar"><h2>Perjalanan <span>{trips.length}</span></h2><Link href="/planner" className="button primary"><Plus size={16} /> Buat perjalanan</Link></div>
      {deleteMutation.isError && <p className="error-text" role="alert">{deleteMutation.error.message}</p>}
      {trips.length ? <div className="trip-grid">{trips.map((trip) => <TripCard key={trip.id} trip={trip} deleting={deleteMutation.isPending && deleteMutation.variables === trip.id} onDelete={() => deleteMutation.mutate(trip.id)} />)}</div> : <EmptyState title="Belum ada perjalanan" description="Mulai rencanakan perjalanan pertama Anda." action={<Link href="/planner" className="button primary">Rencanakan perjalanan</Link>} />}
    </>}
  </>;
}

export function TripDetail({ id }: { id: string }) {
  const router = useRouter();
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();
  const tripQuery = useQuery({ queryKey: ["trip", user?.id, id], queryFn: () => getTrip(token!, id), enabled: Boolean(token && user && id) });
  const [form, setForm] = useState<{ title: string; start_date: string; end_date: string; origin: string; notes: string; status: TripStatus } | null>(null);
  const updateMutation = useMutation({
    mutationFn: () => updateTrip(token!, id, form!),
    onSuccess: async (trip) => { await queryClient.setQueryData(["trip", user?.id, id], trip); await queryClient.invalidateQueries({ queryKey: ["trips", user?.id] }); setForm(null); },
  });
  const deleteMutation = useMutation({
    mutationFn: () => removeTrip(token!, id),
    onSuccess: async () => { await queryClient.invalidateQueries({ queryKey: ["trips", user?.id] }); router.push("/trips"); },
  });

  if (!token) return <EmptyState title="Masuk untuk melihat perjalanan" action={<Link href="/profile" className="button primary">Masuk</Link>} />;
  if (tripQuery.isLoading) return <p aria-live="polite">Memuat perjalanan...</p>;
  if (tripQuery.isError) return <div className="empty-state"><h3>Perjalanan tidak dapat dimuat</h3><p>{tripQuery.error.message}</p><button className="button outline" onClick={() => void tripQuery.refetch()}>Coba lagi</button><Link href="/trips" className="back-link">Kembali ke perjalanan</Link></div>;
  const trip = tripQuery.data;
  if (!trip) return <NotFound />;
  const beginEdit = () => setForm({ title: trip.title, start_date: trip.start_date, end_date: trip.end_date, origin: trip.origin ?? "", notes: trip.notes ?? "", status: trip.status });

  return <>
    <Link href="/trips" className="back-link"><ArrowLeft size={15} /> Perjalanan saya</Link>
    <div className="trip-detail-head"><div><div className="eyebrow">{trip.status.toUpperCase()}</div><h1>{trip.title}</h1><p><CalendarDays size={16} /> {trip.start_date} – {trip.end_date}</p></div><Link href={`/itinerary/${trip.id}`} className="button primary">Buka itinerary <ArrowRight size={16} /></Link></div>
    {form ? <form className="profile-preferences" onSubmit={(event) => { event.preventDefault(); updateMutation.mutate(); }}>
      <h2>Edit perjalanan</h2>
      <label className="form-label">Nama perjalanan<input className="form-input" required maxLength={160} value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} /></label>
      <div className="form-two"><label className="form-label">Mulai<input className="form-input" type="date" required value={form.start_date} onChange={(event) => setForm({ ...form, start_date: event.target.value })} /></label><label className="form-label">Selesai<input className="form-input" type="date" required value={form.end_date} onChange={(event) => setForm({ ...form, end_date: event.target.value })} /></label></div>
      <label className="form-label">Asal<input className="form-input" value={form.origin} onChange={(event) => setForm({ ...form, origin: event.target.value })} /></label>
      <label className="form-label">Catatan<textarea className="form-input" value={form.notes} onChange={(event) => setForm({ ...form, notes: event.target.value })} /></label>
      <label className="form-label">Status<select className="form-input" value={form.status} onChange={(event) => setForm({ ...form, status: event.target.value as TripStatus })}>{["planning", "active", "completed", "cancelled"].map((status) => <option key={status}>{status}</option>)}</select></label>
      <button className="button primary" disabled={updateMutation.isPending}>{updateMutation.isPending ? "Menyimpan..." : "Simpan perubahan"}</button> <button className="button outline" type="button" onClick={() => setForm(null)}>Batal</button>
      {updateMutation.isError && <p className="error-text">{updateMutation.error.message}</p>}
    </form> : <div className="detail-actions"><button className="button outline" onClick={beginEdit}>Edit perjalanan</button><button className="button outline" disabled={deleteMutation.isPending} onClick={() => { if (window.confirm("Hapus perjalanan dan semua item itinerary?")) deleteMutation.mutate(); }}>{deleteMutation.isPending ? "Menghapus..." : "Hapus perjalanan"}</button>{deleteMutation.isError && <p className="error-text">{deleteMutation.error.message}</p>}</div>}
    <p>{trip.origin ? `Berangkat dari ${trip.origin}. ` : ""}{trip.notes}</p>
    <Link href={`/itinerary/${trip.id}`} className="button primary">Kelola itinerary</Link>
  </>;
}
