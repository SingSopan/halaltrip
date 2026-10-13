"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { MapPin, Search, ExternalLink, Navigation } from "lucide-react";
import { getDestinations, getHotels, getMosques, getRestaurants } from "@/src/services";
import type { Place } from "@/src/types/place";

const filters = ["Semua", "Destination", "Restaurant", "Hotel", "Mosque"] as const;
type MapFilter = (typeof filters)[number];

function hasCoordinates(place: Place): place is Place & { latitude: number; longitude: number } {
  return typeof place.latitude === "number" && typeof place.longitude === "number";
}

function osmEmbedURL(place: Place & { latitude: number; longitude: number }) {
  const padLat = 0.009;
  const padLon = 0.012;
  const bounds = [place.longitude - padLon, place.latitude - padLat, place.longitude + padLon, place.latitude + padLat].join(",");
  return `https://www.openstreetmap.org/export/embed.html?bbox=${bounds}&layer=mapnik&marker=${place.latitude},${place.longitude}`;
}

export function MapPage() {
  const [city, setCity] = useState("Jakarta");
  const [cityInput, setCityInput] = useState("Jakarta");
  const [filter, setFilter] = useState<MapFilter>("Semua");
  const [selectedKey, setSelectedKey] = useState("");
  const placesQuery = useQuery({
    queryKey: ["map-places", city],
    queryFn: async () => {
      const filters = { city, page: 1, page_size: 100 };
      const [destinations, restaurants, hotels, mosques] = await Promise.all([
        getDestinations(filters), getRestaurants(filters), getHotels(filters), getMosques(filters),
      ]);
      return [...destinations, ...restaurants, ...hotels, ...mosques];
    },
  });
  const places = useMemo(() => (placesQuery.data ?? []).filter(hasCoordinates), [placesQuery.data]);
  const filteredPlaces = filter === "Semua" ? places : places.filter((place) => place.category === filter);
  const selected = filteredPlaces.find((place) => `${place.category}:${place.id}` === selectedKey) ?? filteredPlaces[0];

  return (
    <section className="map-page">
      <header className="map-page-heading">
        <div><p className="eyebrow">PETA TEMPAT</p><h1>Temukan tempat di sekitar tujuan.</h1><p>Pilih pin dari katalog HalalTrip untuk melihat lokasinya di peta.</p></div>
        <form className="map-city-search" onSubmit={(event) => { event.preventDefault(); setCity(cityInput.trim()); setSelectedKey(""); }}>
          <Search size={17} /><input value={cityInput} onChange={(event) => setCityInput(event.target.value)} aria-label="Cari kota" placeholder="Cari kota" />
          <button className="button primary" type="submit">Cari</button>
        </form>
      </header>
      <div className="map-toolbar" role="group" aria-label="Filter kategori peta">
        {filters.map((option) => <button type="button" key={option} className={filter === option ? "map-filter active" : "map-filter"} onClick={() => { setFilter(option); setSelectedKey(""); }}>{option === "Semua" ? option : `${option}${option.endsWith("s") ? "" : "s"}`}</button>)}
        <span>{placesQuery.isSuccess ? `${filteredPlaces.length} tempat dengan koordinat` : ""}</span>
      </div>
      <div className="map-workspace">
        <aside className="map-place-list" aria-label="Daftar tempat di peta">
          {placesQuery.isLoading ? <div className="empty-state"><p>Memuat tempat…</p></div> : placesQuery.isError ? <div className="empty-state"><h2>Tempat tidak dapat dimuat</h2><p>{placesQuery.error.message}</p><button type="button" className="button outline" onClick={() => void placesQuery.refetch()}>Coba lagi</button></div> : filteredPlaces.length ? filteredPlaces.map((place, index) => {
            const key = `${place.category}:${place.id}`;
            return <button type="button" className={selected?.id === place.id && selected?.category === place.category ? "map-place-row selected" : "map-place-row"} key={key || `${place.name}-${index}`} onClick={() => setSelectedKey(key)}>
              <span className="map-place-pin"><MapPin size={16} /></span><span><strong>{place.name}</strong><small>{place.category} · {place.city}</small></span>
            </button>;
          }) : <div className="empty-state"><h2>Belum ada pin di {city || "kota ini"}</h2><p>Tambahkan koordinat pada data tempat backend agar muncul di peta.</p></div>}
        </aside>
        <div className="map-canvas">
          {selected ? <>
            <iframe title={`Peta lokasi ${selected.name}`} src={osmEmbedURL(selected)} loading="lazy" referrerPolicy="strict-origin-when-cross-origin" />
            <div className="map-selected-card">
              <span className="eyebrow">{selected.category.toUpperCase()}</span><h2>{selected.name}</h2><p>{selected.description || selected.address || `${selected.city}, ${selected.country}`}</p>
              {selected.id && <Link className="button outline" href={`/${selected.category === "Destination" ? "destinations" : `${selected.category.toLowerCase()}s`}/${encodeURIComponent(selected.id)}`}>Lihat detail <ExternalLink size={14} /></Link>}
              <a className="map-directions" href={`https://www.openstreetmap.org/?mlat=${selected.latitude}&mlon=${selected.longitude}#map=16/${selected.latitude}/${selected.longitude}`} target="_blank" rel="noreferrer"><Navigation size={14} /> Buka lokasi di OpenStreetMap</a>
            </div>
          </> : <div className="map-empty"><MapPin size={30} /><h2>Peta menunggu data lokasi</h2><p>Tempat tanpa koordinat tetap bisa dilihat di katalog. Peta hanya menampilkan koordinat yang tersedia dari backend.</p></div>}
          <span className="map-attribution">© <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noreferrer">OpenStreetMap contributors</a></span>
        </div>
      </div>
    </section>
  );
}
