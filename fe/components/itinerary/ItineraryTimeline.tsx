import { ItineraryItem, type ItineraryActivity } from "./ItineraryItem";

export function ItineraryTimeline({ items = [] }: { items?: ItineraryActivity[] }) {
  if (items.length === 0) {
    return <div className="empty-state"><h3>Belum ada aktivitas itinerary</h3><p>Aktivitas perjalanan akan tampil di sini saat tersedia dari backend.</p></div>;
  }
  return <div className="timeline">{items.map((activity) => <ItineraryItem key={`${activity.time}-${activity.title}`} activity={activity} />)}</div>;
}
