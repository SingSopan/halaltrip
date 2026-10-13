"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { ArrowDown, ArrowLeft, ArrowRight, ArrowUp, Check, MapPin, Sparkles, Trash2 } from "lucide-react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createItineraryItem, createTrip, getDestinations, getHotels, getMosques, getRestaurants } from "@/src/services";
import type { Activity, Interest, TravelType, ItineraryInput } from "@/src/types/api";
import type { Place, PlaceCategory } from "@/src/types/place";
import { useAuthStore } from "@/src/stores/use-auth-store";

const interests: { value: Interest; label: string }[] = [
  { value: "culinary", label: "Kuliner" }, { value: "culture", label: "Budaya" },
  { value: "nature", label: "Alam" }, { value: "shopping", label: "Belanja" },
  { value: "adventure", label: "Petualangan" }, { value: "history", label: "Sejarah" },
];
const travelTypes: TravelType[] = ["Solo", "Couple", "Family", "Friends"];
const activities: Activity[] = ["Relaxing", "Exploration", "Adventure"];
const steps = ["Tujuan", "Tanggal", "Rombongan", "Prioritas", "Anggaran", "Transportasi", "Itinerary"];
type DraftStop = { key: string; place: Place; date: string; start: string; end: string };
const kindFor: Record<PlaceCategory, ItineraryInput["place_type"]> = {
  Destination: "destination", Restaurant: "restaurant", Hotel: "hotel", Mosque: "mosque",
};

function dateOffset(days: number) {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date.toISOString().slice(0, 10);
}

export function Planner({ initial }: { initial?: { city?: string; startDate?: string; endDate?: string; travelType?: string } }) {
  const router = useRouter();
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const [step, setStep] = useState(1);
  const [city, setCity] = useState(initial?.city || "Jakarta");
  const [title, setTitle] = useState("");
  const [startDate, setStartDate] = useState(initial?.startDate || dateOffset(7));
  const [endDate, setEndDate] = useState(initial?.endDate || dateOffset(9));
  const [origin, setOrigin] = useState("");
  const [travelType, setTravelType] = useState<TravelType>(initial?.travelType && travelTypes.includes(initial.travelType as TravelType) ? initial.travelType as TravelType : "Solo");
  const [activity, setActivity] = useState<Activity>("Exploration");
  const [interestsSelected, setInterestsSelected] = useState<Interest[]>(["culture", "culinary", "history"]);
  const [budget, setBudget] = useState("Moderate");
  const [transport, setTransport] = useState("Public transport");
  const [catalog, setCatalog] = useState<Place[]>([]);
  const [draft, setDraft] = useState<DraftStop[]>([]);
  const [formError, setFormError] = useState("");

  const days = useMemo(() => {
    if (!startDate || !endDate || endDate < startDate) return [];
    const result: string[] = [];
    const cursor = new Date(`${startDate}T12:00:00`);
    const last = new Date(`${endDate}T12:00:00`);
    while (cursor <= last && result.length < 14) {
      result.push(cursor.toISOString().slice(0, 10));
      cursor.setDate(cursor.getDate() + 1);
    }
    return result;
  }, [startDate, endDate]);

  const generateMutation = useMutation({
    mutationFn: async () => {
      const filters = { city: city.trim(), page: 1, page_size: 100 };
      const [destinations, restaurants, hotels, mosques] = await Promise.all([
        getDestinations(filters), getRestaurants(filters), getHotels(filters), getMosques(filters),
      ]);
      return [...destinations, ...restaurants, ...hotels, ...mosques];
    },
    onSuccess: (places) => {
      setCatalog(places);
      const dates = days.length ? days : [startDate];
      const destinations = places.filter((place) => place.category === "Destination");
      const restaurants = places.filter((place) => place.category === "Restaurant");
      const hotels = places.filter((place) => place.category === "Hotel");
      const mosques = places.filter((place) => place.category === "Mosque");
      const stops: DraftStop[] = [];
      const add = (place: Place | undefined, day: number, start: string, end: string) => {
        if (!place) return;
        // The same place can be listed in multiple catalogs, so avoid duplicate stops by name.
        if (stops.some((stop) => stop.place.name.toLowerCase() === place.name.toLowerCase())) return;
        stops.push({ key: `${place.category}:${place.id || place.name}:${stops.length}`, place, date: dates[Math.min(day, dates.length - 1)], start, end });
      };
      if (hotels[0]) add(hotels[0], 0, "14:00", "14:30");
      for (let day = 0; day < Math.min(dates.length, 7); day += 1) {
        add(destinations[day], day, "09:00", "11:00");
        add(mosques[day], day, "12:00", "12:45");
        add(restaurants[day], day, "13:00", "14:00");
      }
      if (!stops.length) {
        setFormError(`Belum ada data tempat untuk ${city}. Coba kota Jakarta atau tambahkan katalog tempat terlebih dahulu.`);
        return;
      }
      setDraft(stops);
      setStep(7);
    },
  });

  const saveMutation = useMutation({
    mutationFn: async () => {
      const trip = await createTrip(token!, {
        title: title.trim() || `Perjalanan ${city}`,
        start_date: startDate,
        end_date: endDate,
        origin: origin.trim() || undefined,
        status: "planning",
        notes: `Preferensi: ${interestsSelected.join(", ") || "bebas"}; aktivitas ${activity}; ${travelType}; anggaran ${budget}; transportasi ${transport}.`,
      });
      for (const stop of draft) {
        await createItineraryItem(token!, trip.id, {
          title: stop.place.name,
          description: stop.place.description || stop.place.address || `Rencana ${stop.place.category.toLowerCase()} di ${stop.place.city}.`,
          place_type: kindFor[stop.place.category],
          ...(stop.place.id ? { external_place_id: stop.place.id } : {}),
          scheduled_date: stop.date,
          start_time: stop.start,
          end_time: stop.end,
          transport_mode: transport,
        });
      }
      return trip;
    },
    onSuccess: async (trip) => {
      await queryClient.invalidateQueries({ queryKey: ["trips", user?.id] });
      router.push(`/trips/${trip.id}`);
    },
  });

  function moveStop(index: number, direction: -1 | 1) {
    setDraft((current) => {
      const next = [...current];
      const target = index + direction;
      if (target < 0 || target >= next.length) return current;
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  }
  function updateStop(index: number, changes: Partial<DraftStop>) {
    setDraft((current) => current.map((stop, currentIndex) => currentIndex === index ? { ...stop, ...changes } : stop));
  }
  function addPlace(placeId: string) {
    const place = catalog.find((item) => `${item.category}:${item.id || item.name}` === placeId);
    if (place && !draft.some((stop) => stop.place.name === place.name)) {
      setDraft((current) => [...current, { key: `${place.category}:${place.id || place.name}:${Date.now()}`, place, date: days[0] || startDate, start: "15:00", end: "16:00" }]);
    }
  }
  function generate() {
    setFormError("");
    if (!city.trim()) return setFormError("Isi kota tujuan terlebih dahulu.");
    if (!startDate || !endDate || endDate < startDate) return setFormError("Pastikan tanggal selesai tidak sebelum tanggal mulai.");
    generateMutation.mutate();
  }
  function save() {
    setFormError("");
    if (!token) return router.push("/login");
    if (!draft.length) return setFormError("Tambahkan setidaknya satu tempat ke itinerary.");
    if (draft.some((stop) => !stop.date || stop.date < startDate || stop.date > endDate || stop.end <= stop.start)) return setFormError("Periksa tanggal dan jam setiap tempat.");
    saveMutation.mutate();
  }

  const byDay = days.map((date) => ({ date, stops: draft.filter((stop) => stop.date === date) }));
  const addable = catalog.filter((place) => !draft.some((stop) => stop.place.name === place.name));

  return <div className="planner-layout"><div className="planner-main">
    <Link href="/" className="back-link"><ArrowLeft size={15} /> Kembali ke beranda</Link>
    <div className="eyebrow">RENCANA PERJALANAN YANG BISA DISESUAIKAN</div><h1>Susun perjalanan Anda.</h1>
    <p className="planner-intro">Pilih kota dan preferensi. HalalTrip menyusun beberapa tempat dari katalog yang tersedia, lalu Anda dapat mengubah urutan, waktu, atau tempatnya.</p>
    <div className="stepper" aria-label="Langkah trip planner">{steps.map((label, index) => <div key={label} className={step === index + 1 ? "step active" : step > index + 1 ? "step done" : "step"}><span>{step > index + 1 ? <Check size={13} /> : index + 1}</span><small>{label}</small></div>)}</div>

    {step === 1 && <section className="planner-form-panel"><div className="planner-question"><span>LANGKAH 1</span><h2>Mau ke mana?</h2><p>Mulai dengan kota tujuan perjalanan Anda.</p></div><label className="form-label">Kota tujuan<input className="form-input" value={city} onChange={(event) => setCity(event.target.value)} placeholder="Contoh: Jakarta" /></label><label className="form-label">Nama perjalanan<input className="form-input" value={title} onChange={(event) => setTitle(event.target.value)} placeholder={`Perjalanan ${city || "saya"}`} /></label><div className="planner-nav"><span /><button className="button primary" onClick={() => city.trim() ? setStep(2) : setFormError("Isi kota tujuan terlebih dahulu.")}>Lanjutkan <ArrowRight size={15} /></button></div></section>}

    {step === 2 && <section className="planner-form-panel"><div className="planner-question"><span>LANGKAH 2</span><h2>Kapan Anda bepergian?</h2><p>Itinerary otomatis akan mengikuti tanggal ini dan dapat diedit.</p></div><div className="form-two"><label className="form-label">Tanggal mulai<input className="form-input" type="date" value={startDate} onChange={(event) => setStartDate(event.target.value)} /></label><label className="form-label">Tanggal selesai<input className="form-input" type="date" min={startDate} value={endDate} onChange={(event) => setEndDate(event.target.value)} /></label></div><label className="form-label">Kota asal<input className="form-input" value={origin} onChange={(event) => setOrigin(event.target.value)} placeholder="Opsional" /></label><div className="planner-nav"><button className="button outline" onClick={() => setStep(1)}><ArrowLeft size={15} /> Kembali</button><button className="button primary" onClick={() => setStep(3)}>Lanjutkan <ArrowRight size={15} /></button></div></section>}

    {step === 3 && <section className="planner-form-panel"><div className="planner-question"><span>LANGKAH 3</span><h2>Siapa yang ikut?</h2><p>Kami gunakan ini sebagai konteks perjalanan.</p></div><label className="form-label">Tipe perjalanan<select className="form-input" value={travelType} onChange={(event) => setTravelType(event.target.value as TravelType)}>{travelTypes.map((type) => <option key={type} value={type}>{type}</option>)}</select></label><div className="planner-nav"><button className="button outline" onClick={() => setStep(2)}><ArrowLeft size={15} /> Kembali</button><button className="button primary" onClick={() => setStep(4)}>Lanjutkan <ArrowRight size={15} /></button></div></section>}

    {step === 4 && <section className="planner-form-panel"><div className="planner-question"><span>LANGKAH 4</span><h2>Apa yang Anda sukai?</h2><p>Tempat akan disusun dari data kota yang tersedia. Pilihan ini disimpan sebagai konteks perjalanan.</p></div><fieldset className="recommend-fieldset"><legend>Minat perjalanan</legend><div className="preference-options">{interests.map((option) => <label className="choice" key={option.value}><input type="checkbox" checked={interestsSelected.includes(option.value)} onChange={(event) => setInterestsSelected((current) => event.target.checked ? [...current, option.value] : current.filter((value) => value !== option.value))} /><span>{option.label}</span></label>)}</div></fieldset><label className="form-label">Gaya aktivitas<select className="form-input" value={activity} onChange={(event) => setActivity(event.target.value as Activity)}>{activities.map((option) => <option key={option} value={option}>{option}</option>)}</select></label><div className="planner-nav"><button className="button outline" onClick={() => setStep(3)}><ArrowLeft size={15} /> Kembali</button><button className="button primary" onClick={() => setStep(5)}>Lanjutkan <ArrowRight size={15} /></button></div></section>}

    {step === 5 && <section className="planner-form-panel"><div className="planner-question"><span>LANGKAH 5</span><h2>Atur anggaran perjalanan</h2><p>Anggaran menjadi konteks saat Anda meninjau rencana dan memilih tempat.</p></div><label className="form-label">Kisaran anggaran<select className="form-input" value={budget} onChange={(event) => setBudget(event.target.value)}><option value="Budget">Hemat</option><option value="Moderate">Menengah</option><option value="Comfortable">Nyaman</option><option value="Luxury">Premium</option></select></label><div className="planner-nav"><button className="button outline" onClick={() => setStep(4)}><ArrowLeft size={15} /> Kembali</button><button className="button primary" onClick={() => setStep(6)}>Lanjutkan <ArrowRight size={15} /></button></div></section>}

    {step === 6 && <section className="planner-form-panel"><div className="planner-question"><span>LANGKAH 6</span><h2>Bagaimana Anda ingin berkeliling?</h2><p>Mode transportasi akan dicantumkan pada setiap aktivitas dan bisa disesuaikan nanti.</p></div><label className="form-label">Transportasi<select className="form-input" value={transport} onChange={(event) => setTransport(event.target.value)}>{["Public transport", "Walking", "Car", "Ride-hailing", "Flexible"].map((option) => <option key={option}>{option}</option>)}</select></label><p className="planner-note"><MapPin size={15} /> Rencana awal memakai destinasi, restoran, hotel, dan masjid yang tersedia di katalog backend kota {city}.</p>{(formError || generateMutation.error) && <p className="error-text" role="alert">{formError || generateMutation.error?.message}</p>}<div className="planner-nav"><button className="button outline" onClick={() => setStep(5)}><ArrowLeft size={15} /> Kembali</button><button className="button primary" disabled={generateMutation.isPending} onClick={generate}>{generateMutation.isPending ? "Menyusun rencana..." : "Buat itinerary otomatis"} <Sparkles size={15} /></button></div></section>}

    {step === 7 && <section className="planner-form-panel"><div className="planner-question"><span>LANGKAH 7 · DRAF ANDA</span><h2>Itinerary {city}</h2><p>Ubah tempat, hari, dan waktu; gunakan panah untuk mengatur urutan.</p></div>
      {byDay.map(({ date, stops }) => <section className="planner-day" key={date}><h3>{new Date(`${date}T12:00:00`).toLocaleDateString("id-ID", { weekday: "long", day: "numeric", month: "long" })}</h3>{stops.map((stop) => { const index = draft.findIndex((item) => item.key === stop.key); return <article className="planner-stop" key={stop.key}>
        {stop.place.image ? <img src={stop.place.image} alt="" /> : <div className="planner-stop-image"><MapPin size={19} /></div>}
        <div className="planner-stop-content"><span className="eyebrow">{stop.place.category.toUpperCase()}</span><strong>{stop.place.name}</strong><p>{stop.place.description || stop.place.address || `${stop.place.city}, ${stop.place.country}`}</p><div className="planner-stop-edit"><label>Tanggal<input type="date" min={startDate} max={endDate} value={stop.date} onChange={(event) => updateStop(index, { date: event.target.value })} /></label><label>Mulai<input type="time" value={stop.start} onChange={(event) => updateStop(index, { start: event.target.value })} /></label><label>Selesai<input type="time" value={stop.end} onChange={(event) => updateStop(index, { end: event.target.value })} /></label></div></div>
        <div className="planner-stop-actions"><button aria-label="Pindah ke atas" disabled={index === 0} onClick={() => moveStop(index, -1)}><ArrowUp size={15} /></button><button aria-label="Pindah ke bawah" disabled={index === draft.length - 1} onClick={() => moveStop(index, 1)}><ArrowDown size={15} /></button><button aria-label={`Hapus ${stop.place.name}`} onClick={() => setDraft((current) => current.filter((item) => item.key !== stop.key))}><Trash2 size={15} /></button></div>
      </article>; })}{stops.length === 0 && <p className="planner-day-empty">Belum ada tempat untuk hari ini.</p>}</section>)}
      {addable.length > 0 && <label className="planner-add-place">Tambahkan tempat<select className="form-input" value="" onChange={(event) => addPlace(event.target.value)}><option value="">Pilih tempat dari katalog</option>{addable.map((place, index) => <option key={`${place.category}:${place.id || index}`} value={`${place.category}:${place.id || place.name}`}>{place.name} · {place.category}</option>)}</select></label>}
      {!token && <p>Masuk untuk menyimpan itinerary ini ke akun Anda. Draf ini tersedia selama halaman tetap terbuka.</p>}
      {(formError || saveMutation.error) && <p className="error-text" role="alert">{formError || saveMutation.error?.message}</p>}
      <div className="planner-nav"><button className="button outline" onClick={() => setStep(6)}><ArrowLeft size={15} /> Kembali</button><button className="button primary" disabled={saveMutation.isPending || !draft.length} onClick={save}>{saveMutation.isPending ? "Menyimpan itinerary..." : token ? "Simpan perjalanan" : "Masuk untuk menyimpan"} <ArrowRight size={15} /></button></div>
    </section>}
    <p className="planner-footnote">Perkiraan itinerary memakai data katalog tempat. Periksa kembali jam buka dan perjalanan sebelum berangkat.</p>
  </div><aside className="planner-aside"><div className="aside-card"><span className="eyebrow">RINGKASAN</span><h3>{city || "Pilih kota"}</h3><p>{title || `Perjalanan ${city}`}</p><p>{startDate} — {endDate}</p><p>{travelType} · {activity}</p><p>Anggaran: {budget}</p><p>{draft.length} tempat dalam draf</p></div></aside></div>;
}
