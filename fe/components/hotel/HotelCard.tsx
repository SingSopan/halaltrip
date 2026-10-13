import type { Place } from "@/src/types/place";
import { PlaceCard } from "@/components/common/PlaceCard";

export function HotelCard({ place }: { place: Place }) {
  return <PlaceCard place={place} />;
}
