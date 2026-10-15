"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, MapPin } from "lucide-react";
import { LoadingState } from "@/components/common";
import { SavePlaceButton } from "@/components/common/SavePlaceButton";
import { getPlaceByKind } from "@/src/services";
import type { PlaceKind } from "@/src/services/catalog";
import { NotFoundPage as NotFound } from "./NotFoundPage";

const validKinds: PlaceKind[] = ["destinations", "restaurants", "hotels", "mosques"];

export function Detail({ kind, slug }: { kind: string; slug: string }) {
  const validKind = validKinds.includes(kind as PlaceKind) ? kind as PlaceKind : undefined;
  const query = useQuery({ queryKey: ["place", kind, slug], queryFn: () => getPlaceByKind(validKind!, slug), enabled: Boolean(validKind && slug) });
  if (!validKind) return <NotFound />;
  if (query.isLoading) return <LoadingState count={1} />;
  if (query.isError) return <div className="empty-state"><h3>Tempat tidak dapat dimuat</h3><p>{query.error.message}</p><button className="button outline" onClick={() => void query.refetch()}>Coba lagi</button><Link href={`/${kind}`} className="back-link">Kembali ke katalog</Link></div>;
  const place = query.data;
  if (!place) return <NotFound />;
  const safeImage = place.image && (/^https?:\/\//i.test(place.image) || place.image.startsWith("/")) ? place.image : undefined;

  return <>
    <Link href={`/${kind}`} className="back-link"><ArrowLeft size={15} /> Semua {kind}</Link>
    <section className={`detail-hero ${safeImage ? "has-image" : ""}`}>
      {/* Image hosts are supplied by the backend catalog, so keep rendering independent of a build-time host allowlist. */}
      {safeImage && <img className="detail-background" src={safeImage} alt="" />}
      <div className="detail-overlay" />
      <div className="detail-hero-content"><div className="eyebrow light">{[place.city, place.country].filter(Boolean).join(" · ") || place.category}</div><h1>{place.name}</h1>{place.address && <p><MapPin size={16} /> {place.address}</p>}</div>
    </section>
    <div className="detail-actions"><SavePlaceButton place={place} /><Link className="button primary" href="/planner">Rencanakan perjalanan</Link></div>
    <div className="detail-grid"><main className="detail-main">
      {place.description && <section className="content-section"><div className="eyebrow">TENTANG TEMPAT INI</div><h2>{place.name}</h2><p>{place.description}</p></section>}
      {place.image_attribution && <p className="image-credit">Kredit foto: {place.image_attribution}{place.image_license_url && <> · <a href={place.image_license_url} target="_blank" rel="noreferrer">Lihat lisensi</a></>}</p>}
      <section className="content-section"><h2>Detail tempat</h2>
        <dl className="fact-grid">
          {place.city && <div><dt>Kota</dt><dd>{place.city}</dd></div>}{place.country && <div><dt>Negara</dt><dd>{place.country}</dd></div>}
          {place.cuisine && <div><dt>Jenis kuliner</dt><dd>{place.cuisine}</dd></div>}{place.price_range && <div><dt>Kisaran harga</dt><dd>{place.price_range}</dd></div>}
          {place.halal_status && <div><dt>Status halal</dt><dd>{place.halal_status.replaceAll("_", " ")}</dd></div>}
          {place.halal_verification_note && <div><dt>Catatan verifikasi</dt><dd>{place.halal_verification_note}</dd></div>}
          {place.score !== undefined && <div><dt>Skor provider</dt><dd>{place.score}</dd></div>}{place.rating !== undefined && <div><dt>Rating provider</dt><dd>{place.rating}</dd></div>}
          {place.source_url && <div><dt>Sumber</dt><dd><a href={place.source_url} target="_blank" rel="noreferrer">Lihat sumber</a></dd></div>}
        </dl>
        {place.facilities?.length ? <div><h3>Fasilitas</h3><ul>{place.facilities.map((facility) => <li key={facility}>{facility}</li>)}</ul></div> : null}
        {place.muslim_friendly_features?.length ? <div><h3>Fitur ramah Muslim</h3><ul>{place.muslim_friendly_features.map((feature) => <li key={feature.name}>{feature.name}{feature.verified ? " · Terverifikasi" : ""}{feature.verification_note ? ` — ${feature.verification_note}` : ""}</li>)}</ul></div> : null}
      </section>
    </main></div>
  </>;
}
