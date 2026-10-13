import Link from "next/link";
import { CalendarDays, Trash2 } from "lucide-react";
import type { Trip } from "@/src/types/api";

export function TripCard({ trip, onDelete, deleting = false }: { trip: Trip; onDelete: () => void; deleting?: boolean }) {
  return <article className="trip-card">
    <div className="trip-info">
      <span className="eyebrow">{trip.origin || "PERJALANAN"}</span>
      <h3>{trip.title}</h3>
      <div><CalendarDays size={15} /> {trip.start_date} – {trip.end_date}</div>
      <div className="trip-card-bottom"><span>Status: {trip.status}</span><Link href={`/trips/${trip.id}`} className="button small outline">Lihat perjalanan</Link></div>
      <div className="trip-actions"><button disabled={deleting} onClick={onDelete}><Trash2 size={13} /> {deleting ? "Menghapus..." : "Hapus"}</button></div>
    </div>
  </article>;
}
