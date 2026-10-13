import type { Place } from "@/src/types/place";
import { PlaceCard } from "@/components/common";

export function MosqueCard({ place }: { place: Place }) {
  return <PlaceCard place={place} />;
}
