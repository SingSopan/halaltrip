import type { Place } from "@/src/types/place";
import { PlaceCard } from "@/components/common/PlaceCard";

export function DestinationCard({ place }: { place: Place }) {
  return <PlaceCard place={place} />;
}
