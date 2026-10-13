"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowRight, Heart, MapPin, Star } from "lucide-react";
import type { Place } from "@/src/types/place";
import { useSavedPlace } from "@/src/hooks/use-saved-place";
import { Score } from "./Score";
import { HalalBadge } from "@/components/restaurant/HalalBadge";

const categoryRoutes = { Destination: "destinations", Restaurant: "restaurants", Hotel: "hotels", Mosque: "mosques" } as const;

export function PlaceCard({ place }: { place: Place }) {
  const router = useRouter();
  const { token, canSave, isSaved, mutation: saveMutation } = useSavedPlace(place);
  const href = place.id ? `/${categoryRoutes[place.category]}/${encodeURIComponent(place.id)}` : undefined;
  const safeImage = place.image && (/^https?:\/\//i.test(place.image) || place.image.startsWith("/")) ? place.image : undefined;
  const isCertified = place.category === "Restaurant" && place.halal_status === "certified";
  const hotelFeature = place.muslim_friendly_features?.find((feature) => feature.verified);

  return (
    <article className="place-card">
      <div className="card-image">
        {safeImage ? <img src={safeImage} alt={`${place.name}${place.country ? `, ${place.country}` : ""}`} /> : <div className="place-image-empty" aria-hidden="true"><MapPin size={24} /></div>}
        {href && <Link className="card-image-link" href={href} aria-label={`Lihat ${place.name}`} />}
        <button className={`heart-button ${isSaved ? "active" : ""}`} aria-label={`${isSaved ? "Hapus" : "Simpan"} ${place.name}`} onClick={() => {
          if (!token) { router.push("/profile"); return; }
          if (canSave) saveMutation.mutate();
        }} disabled={saveMutation.isPending || (!canSave && Boolean(token))} title={!canSave ? "Saved places hanya tersedia untuk restoran, masjid, dan hotel lokal." : undefined}>
          <Heart size={18} fill={isSaved ? "currentColor" : "none"} />
        </button>
      </div>
      <div className="card-body">
        {(place.city || place.country) && <div className="card-location"><MapPin size={13} /> {[place.city, place.country].filter(Boolean).join(", ")}</div>}
        <div className="card-title-row">
          {href ? <Link href={href} className="card-title">{place.name}</Link> : <strong className="card-title">{place.name}</strong>}
          {place.score !== undefined ? <Score value={place.score} /> : place.rating !== undefined ? <span className="rating"><Star size={14} fill="currentColor" /> {place.rating}</span> : null}
        </div>
        {place.description && <p className="card-description">{place.description}</p>}
        <div className="card-foot">
          {isCertified ? <HalalBadge /> : place.category === "Restaurant" && place.halal_status && place.halal_status !== "unknown" ? <span className="verified">{place.halal_status.replaceAll("_", " ")}</span> : null}
          {place.category === "Mosque" && place.facilities?.length ? <span className="verified">{place.facilities.join(", ")}</span> : null}
          {hotelFeature && <span className="verified">{hotelFeature.name}</span>}
          {href && <Link href={href} className="arrow-link" aria-label={`Lihat ${place.name}`}><ArrowRight size={17} /></Link>}
        </div>
        {!token && <small className="place-save-hint">Masuk untuk menyimpan tempat lokal.</small>}
      </div>
    </article>
  );
}
