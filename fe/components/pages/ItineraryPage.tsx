"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { ArrowDown, ArrowLeft, ArrowUp, Trash2 } from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createItineraryItem, deleteItineraryItem, getTrip, listItinerary, reorderItinerary, updateItineraryItem } from "@/src/services";
import { useAuthStore } from "@/src/stores/use-auth-store";
import type { ItineraryInput, ItineraryItem } from "@/src/types/api";
import { EmptyState } from "@/components/common";

const blankItem: ItineraryInput = { title: "", place_type: "custom_activity" };
const placeTypes: ItineraryInput["place_type"][] = ["destination", "restaurant", "mosque", "hotel", "custom_activity"];

export function Itinerary({ id }: { id: string }) {
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();
  const [form, setForm] = useState<ItineraryInput>(blankItem);
  const [editing, setEditing] = useState<ItineraryItem | null>(null);
  const tripQuery = useQuery({ queryKey: ["trip", user?.id, id], queryFn: () => getTrip(token!, id), enabled: Boolean(token && user && id) });
  const itemsQuery = useQuery({ queryKey: ["itinerary", user?.id, id], queryFn: () => listItinerary(token!, id), enabled: Boolean(token && user && id) });
  const refresh = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["itinerary", user?.id, id] }),
      queryClient.invalidateQueries({ queryKey: ["trip", user?.id, id] }),
    ]);
  };
  const addMutation = useMutation({ mutationFn: (input: ItineraryInput) => createItineraryItem(token!, id, input), onSuccess: refresh });
  const editMutation = useMutation({ mutationFn: () => updateItineraryItem(token!, id, editing!.id, editing!), onSuccess: async () => { setEditing(null); await refresh(); } });
  const deleteMutation = useMutation({ mutationFn: (itemId: string) => deleteItineraryItem(token!, id, itemId), onSuccess: refresh });
  const orderMutation = useMutation({ mutationFn: (ids: string[]) => reorderItinerary(token!, id, ids), onSuccess: refresh });

  function submit(event: FormEvent) {
    event.preventDefault();
    const input = Object.fromEntries(Object.entries(form).filter(([, value]) => value !== "")) as ItineraryInput;
    addMutation.mutate(input);
  }
  function moveItem(index: number, offset: -1 | 1) {
    const items = itemsQuery.data ?? [];
    const target = index + offset;
    if (target < 0 || target >= items.length) return;
    const reordered = [...items];
    [reordered[index], reordered[target]] = [reordered[target], reordered[index]];
    orderMutation.mutate(reordered.map((item) => item.id));
  }

  if (!token) return <EmptyState title="Masuk untuk mengelola itinerary" action={<Link href="/profile" className="button primary">Masuk</Link>} />;
  if (tripQuery.isLoading || itemsQuery.isLoading) return <p aria-live="polite">Memuat itinerary...</p>;
  if (tripQuery.isError || itemsQuery.isError) return <div className="empty-state"><h3>Itinerary tidak dapat dimuat</h3><p>{(tripQuery.error ?? itemsQuery.error)?.message}</p><button className="button outline" onClick={() => { void tripQuery.refetch(); void itemsQuery.refetch(); }}>Coba lagi</button><Link href="/trips" className="back-link">Kembali ke perjalanan</Link></div>;
  const trip = tripQuery.data;

  return <>
    <Link href={`/trips/${id}`} className="back-link"><ArrowLeft size={15} /> {trip?.title ?? "Detail perjalanan"}</Link>
    <div className="itinerary-heading"><div><div className="eyebrow">TRIP ITINERARY</div><h1>{trip?.title ?? "Rencana perjalanan"}</h1><p>{trip?.start_date} – {trip?.end_date}</p></div></div>
    <form className="itinerary-form" onSubmit={submit}>
      <h2>Tambah aktivitas</h2>
      <label className="form-label">Nama aktivitas<input className="form-input" required maxLength={200} value={form.title} onChange={(event) => setForm({ ...form, title: event.target.value })} /></label>
      <div className="form-two"><label className="form-label">Jenis<select className="form-input" value={form.place_type} onChange={(event) => setForm({ ...form, place_type: event.target.value as ItineraryInput["place_type"] })}>{placeTypes.map((type) => <option key={type} value={type}>{type.replaceAll("_", " ")}</option>)}</select></label><label className="form-label">Tanggal<input className="form-input" type="date" min={trip?.start_date} max={trip?.end_date} value={form.scheduled_date ?? ""} onChange={(event) => setForm({ ...form, scheduled_date: event.target.value })} /></label></div>
      <div className="form-two"><label className="form-label">Mulai<input className="form-input" type="time" value={form.start_time ?? ""} onChange={(event) => setForm({ ...form, start_time: event.target.value })} /></label><label className="form-label">Selesai<input className="form-input" type="time" value={form.end_time ?? ""} onChange={(event) => setForm({ ...form, end_time: event.target.value })} /></label></div>
      <label className="form-label">Deskripsi<textarea className="form-input" value={form.description ?? ""} onChange={(event) => setForm({ ...form, description: event.target.value })} /></label>
      <label className="form-label">Moda transportasi<input className="form-input" value={form.transport_mode ?? ""} onChange={(event) => setForm({ ...form, transport_mode: event.target.value })} /></label>
      <button className="button primary" disabled={addMutation.isPending}>{addMutation.isPending ? "Menyimpan..." : "Tambah ke itinerary"}</button>
      {addMutation.isError && <p className="error-text" role="alert">{addMutation.error.message}</p>}
    </form>
    {itemsQuery.data?.length ? <section className="itinerary-items" aria-label="Aktivitas itinerary">
      {itemsQuery.data.map((item, index) => <article className="itinerary-entry" key={item.id}>
        {editing?.id === item.id ? <div className="itinerary-edit">
          <label className="form-label">Nama<input className="form-input" value={editing.title} onChange={(event) => setEditing({ ...editing, title: event.target.value })} /></label>
          <div className="form-two"><label className="form-label">Jenis<select className="form-input" value={editing.place_type} onChange={(event) => setEditing({ ...editing, place_type: event.target.value as ItineraryItem["place_type"] })}>{placeTypes.map((type) => <option key={type} value={type}>{type.replaceAll("_", " ")}</option>)}</select></label><label className="form-label">Tanggal<input className="form-input" type="date" min={trip?.start_date} max={trip?.end_date} value={editing.scheduled_date ?? ""} onChange={(event) => setEditing({ ...editing, scheduled_date: event.target.value })} /></label></div>
          <div className="form-two"><label className="form-label">Mulai<input className="form-input" type="time" value={editing.start_time?.slice(0, 5) ?? ""} onChange={(event) => setEditing({ ...editing, start_time: event.target.value })} /></label><label className="form-label">Selesai<input className="form-input" type="time" value={editing.end_time?.slice(0, 5) ?? ""} onChange={(event) => setEditing({ ...editing, end_time: event.target.value })} /></label></div>
          <label className="form-label">Deskripsi<textarea className="form-input" value={editing.description ?? ""} onChange={(event) => setEditing({ ...editing, description: event.target.value })} /></label>
          <label className="form-label">Moda transportasi<input className="form-input" value={editing.transport_mode ?? ""} onChange={(event) => setEditing({ ...editing, transport_mode: event.target.value })} /></label>
          <label className="form-label">Catatan<textarea className="form-input" value={editing.notes ?? ""} onChange={(event) => setEditing({ ...editing, notes: event.target.value })} /></label>
          <button className="button primary" disabled={editMutation.isPending} onClick={() => editMutation.mutate()}>Simpan</button> <button className="button outline" onClick={() => setEditing(null)}>Batal</button>
          {editMutation.isError && <p className="error-text">{editMutation.error.message}</p>}
        </div> : <>
          <div className="itinerary-entry-head"><div><span className="eyebrow">{item.scheduled_date || `URUTAN ${item.sequence}`} · {item.place_type.replaceAll("_", " ")}</span><h3>{item.title}</h3>{(item.start_time || item.end_time) && <p>{item.start_time} – {item.end_time}</p>}{item.description && <p>{item.description}</p>}</div><div className="itinerary-entry-actions"><button aria-label="Pindah ke atas" title="Pindah ke atas" disabled={index === 0 || orderMutation.isPending} onClick={() => moveItem(index, -1)}><ArrowUp size={16} /></button><button aria-label="Pindah ke bawah" title="Pindah ke bawah" disabled={index === (itemsQuery.data?.length ?? 1) - 1 || orderMutation.isPending} onClick={() => moveItem(index, 1)}><ArrowDown size={16} /></button><button onClick={() => setEditing(item)}>Edit</button><button aria-label={`Hapus ${item.title}`} disabled={deleteMutation.isPending} onClick={() => deleteMutation.mutate(item.id)}><Trash2 size={15} /></button></div></div>
        </>}
      </article>)}
      {(orderMutation.isError || deleteMutation.isError) && <p className="error-text" role="alert">{orderMutation.error?.message ?? deleteMutation.error?.message}</p>}
    </section> : <EmptyState title="Belum ada aktivitas" description="Tambahkan aktivitas untuk mulai menyusun perjalanan." />}
  </>;
}
