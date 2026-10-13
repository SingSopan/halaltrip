"use client";

import { useDeferredValue, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listPlaces, type PlaceFilters, type PlaceKind } from "@/src/services/catalog";
import { LoadingState, PlaceCard, SearchBox } from "@/components/common";
import { PageIntro } from "./PageIntro";
import { LISTING_TITLES, LISTING_DESCRIPTIONS } from "./page-data";

const validKinds: PlaceKind[] = ["destinations", "restaurants", "hotels", "mosques"];

export function Listing({ kind }: { kind: string }) {
  const [search, setSearch] = useState("");
  const [city, setCity] = useState("");
  const [halalStatus, setHalalStatus] = useState("");
  const [cuisine, setCuisine] = useState("");
  const [feature, setFeature] = useState("");
  const [facility, setFacility] = useState("");
  const [page, setPage] = useState(1);
  const q = useDeferredValue(search.trim());
  const validKind = validKinds.includes(kind as PlaceKind) ? kind as PlaceKind : undefined;
  const filters: PlaceFilters = { q, city: city.trim(), page, page_size: 12 };
  if (kind === "restaurants") Object.assign(filters, { halal_status: halalStatus, cuisine });
  if (kind === "hotels") Object.assign(filters, { feature });
  if (kind === "mosques") Object.assign(filters, { facility });
  const query = useQuery({
    queryKey: ["places", kind, filters],
    queryFn: () => listPlaces(validKind!, filters),
    enabled: Boolean(validKind),
  });
  const items = query.data?.items ?? [];
  const localPage = query.data?.pagination;
  const title = LISTING_TITLES[kind] ?? "Jelajahi tempat";
  const description = LISTING_DESCRIPTIONS[kind] ?? "Tempat dari backend HalalTrip.";

  return <>
    <PageIntro eyebrow={`JELAJAHI ${kind.toUpperCase()}`} title={title} description={description} />
    <div className="listing-tools"><SearchBox placeholder={`Cari ${kind}...`} value={search} onChange={(value) => { setSearch(value); setPage(1); }} />
      {kind !== "destinations" && <label className="filter-control">Kota<input className="form-input" value={city} onChange={(event) => { setCity(event.target.value); setPage(1); }} placeholder="Semua kota" /></label>}
      {kind === "restaurants" && <>
        <label className="filter-control">Status halal<select className="form-input" value={halalStatus} onChange={(event) => { setHalalStatus(event.target.value); setPage(1); }}><option value="">Semua status</option><option value="certified">Bersertifikat</option><option value="muslim_owned">Muslim-owned</option><option value="halal_friendly">Halal-friendly</option><option value="no_pork">Tanpa pork</option><option value="unknown">Belum diketahui</option></select></label>
        <label className="filter-control">Kuliner<input className="form-input" value={cuisine} onChange={(event) => { setCuisine(event.target.value); setPage(1); }} placeholder="Semua jenis" /></label>
      </>}
      {kind === "hotels" && <label className="filter-control">Fitur<input className="form-input" value={feature} onChange={(event) => { setFeature(event.target.value); setPage(1); }} placeholder="Fitur ramah Muslim" /></label>}
      {kind === "mosques" && <label className="filter-control">Fasilitas<input className="form-input" value={facility} onChange={(event) => { setFacility(event.target.value); setPage(1); }} placeholder="Fasilitas" /></label>}
    </div>
    <div className="listing-layout listing-no-sidebar"><section className="listing-results" aria-live="polite">
      <div className="results-top"><span><strong>{localPage?.total ?? items.length}</strong> hasil dari backend</span>{localPage && <span>Halaman {localPage.page} dari {localPage.total_pages}</span>}</div>
      {!validKind ? <div className="empty-state"><h3>Kategori tidak tersedia</h3></div> : query.isLoading ? <LoadingState /> : query.isError ? <div className="empty-state"><h3>Katalog tidak dapat dimuat</h3><p>{query.error.message}</p><button type="button" className="button outline" onClick={() => void query.refetch()}>Coba lagi</button></div> : items.length ? <>
        <div className="destination-grid">{items.map((place, index) => <PlaceCard key={place.id || `${place.name}-${index}`} place={place} />)}</div>
        {localPage && localPage.total_pages > 1 && <div className="pagination-actions"><button className="button outline" disabled={page <= 1} onClick={() => setPage((current) => current - 1)}>Sebelumnya</button><button className="button outline" disabled={page >= localPage.total_pages} onClick={() => setPage((current) => current + 1)}>Berikutnya</button></div>}
      </> : <div className="empty-state"><h3>Belum ada data yang cocok</h3><p>Katalog backend tidak mengembalikan tempat untuk filter ini.</p><button className="button outline" onClick={() => { setSearch(""); setCity(""); setHalalStatus(""); setCuisine(""); setFeature(""); setFacility(""); setPage(1); }}>Hapus filter</button></div>}
    </section></div>
  </>;
}
