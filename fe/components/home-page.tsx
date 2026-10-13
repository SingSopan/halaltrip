"use client";

import Image from "next/image";
import Link from "next/link";
import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowRight,
  Building2,
  CalendarDays,
  Check,
  ChevronDown,
  Compass,
  Map as MapIcon,
  MoonStar,
  Sparkles,
  TrainFront,
  Users,
  Utensils,
} from "lucide-react";
import type { Place } from "@/src/types/place";
import { getDestinations, getHotels, getMosques, getRestaurants } from "@/src/services";
import { PlaceCard, SearchBox } from "@/components/common";
import { Footer, Header } from "@/components/layout";

const CATEGORY_TAB_MAP: Record<Place["category"], string> = {
  Destination: "Destinations",
  Restaurant: "Restaurants",
  Hotel: "Hotels",
  Mosque: "Mosques",
};

const EXPLORE_TABS = [
  "All",
  "Destinations",
  "Restaurants",
  "Hotels",
  "Mosques",
] as const;

type ExploreTab = (typeof EXPLORE_TABS)[number];
const EMPTY_PLACES: Place[] = [];

const FEATURES = [
  {
    icon: Utensils,
    title: "Food you can trust",
    description: "Halal options, clearly verified and close by.",
  },
  {
    icon: MoonStar,
    title: "Space to pause",
    description: "Prayer spaces built into the flow of your day.",
  },
  {
    icon: Building2,
    title: "Stays that understand",
    description: "Thoughtful amenities for a more comfortable stay.",
  },
  {
    icon: TrainFront,
    title: "Easy to get around",
    description: "Local transport insights that keep plans simple.",
  },
];

export function HomePage() {
  const [query, setQuery] = useState("");
  const [plannerCity, setPlannerCity] = useState("Jakarta");
  const [plannerStart, setPlannerStart] = useState(() => new Date(Date.now() + 7 * 86400000).toISOString().slice(0, 10));
  const [plannerEnd, setPlannerEnd] = useState(() => new Date(Date.now() + 9 * 86400000).toISOString().slice(0, 10));
  const [plannerTravelType, setPlannerTravelType] = useState("Solo");
  const plannerHref = `/planner?${new URLSearchParams({ city: plannerCity, startDate: plannerStart, endDate: plannerEnd, travelType: plannerTravelType })}`;
  const [tab, setTab] = useState<ExploreTab>("All");
  const destinationsQuery = useQuery({ queryKey: ["destinations"], queryFn: () => getDestinations() });
  const restaurantsQuery = useQuery({ queryKey: ["restaurants"], queryFn: () => getRestaurants() });
  const hotelsQuery = useQuery({ queryKey: ["hotels"], queryFn: () => getHotels() });
  const mosquesQuery = useQuery({ queryKey: ["mosques"], queryFn: () => getMosques() });
  const destinations = destinationsQuery.data ?? EMPTY_PLACES;
  const restaurants = restaurantsQuery.data ?? EMPTY_PLACES;
  const hotels = hotelsQuery.data ?? EMPTY_PLACES;
  const mosques = mosquesQuery.data ?? EMPTY_PLACES;

  const allPlaces = useMemo(
    () => [
      ...destinations,
      ...restaurants.slice(0, 3),
      ...hotels.slice(0, 3),
      ...mosques.slice(0, 3),
    ],
    [destinations, restaurants, hotels, mosques]
  );

  const filteredPlaces = useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();

    return allPlaces.filter((place) => {
      const matchesTab =
        tab === "All" || CATEGORY_TAB_MAP[place.category] === tab;

      const matchesQuery =
        !normalizedQuery ||
        `${place.name} ${place.city} ${place.country}`
          .toLowerCase()
          .includes(normalizedQuery);

      return matchesTab && matchesQuery;
    });
  }, [allPlaces, tab, query]);

  return (
    <>
      <Header />
      <main>
        {/* Hero Section */}
        <section className="hero">
          <div className="hero-image">
            <Image
              src="https://images.unsplash.com/photo-1542051841857-5f90071e7989?auto=format&fit=crop&w=2200&q=90"
            alt="City lights at dusk"
              fill
              priority
              sizes="100vw"
            />
          </div>
          <div className="hero-shade" />

          <div className="hero-inner">
            <div className="hero-copy">
              <div className="eyebrow light">
                <span /> JOURNEYS WITH PEACE OF MIND
              </div>
              <h1>
                Travel the world.
                <br />
                <em>Stay halal.</em>
              </h1>
              <p>
                Discover the places that make you feel at home, wherever you go.
              </p>
              <div className="hero-proof">
                <span>
                  <Check size={14} /> Halal food
                </span>
                <span>
                  <Check size={14} /> Prayer access
                </span>
                <span>
                  <Check size={14} /> Thoughtful stays
                </span>
              </div>
            </div>

            <div className="plan-card">
              <div className="plan-head">
                <div>
                  <span className="overline">YOUR NEXT JOURNEY</span>
                  <h2>Where to?</h2>
                </div>
                <span className="plan-icon">
                  <Sparkles size={18} />
                </span>
              </div>

              <div className="field-label">DESTINATION</div>
              <div className="destination-input">
                <Compass size={17} />
                <input
                  value={plannerCity}
                  onChange={(e) => setPlannerCity(e.target.value)}
                  placeholder="Contoh: Jakarta"
                  aria-label="Kota tujuan perjalanan"
                />
                <ChevronDown size={16} />
              </div>

              <div className="field-row">
                <label>
                  <span className="field-label">CHECK IN</span>
                  <span className="field-value"><CalendarDays size={16} /><input type="date" value={plannerStart} onChange={(event) => { setPlannerStart(event.target.value); if (plannerEnd < event.target.value) setPlannerEnd(event.target.value); }} aria-label="Tanggal mulai" /></span>
                </label>
                <label>
                  <span className="field-label">CHECK OUT</span>
                  <span className="field-value"><CalendarDays size={16} /><input type="date" min={plannerStart} value={plannerEnd} onChange={(event) => setPlannerEnd(event.target.value)} aria-label="Tanggal selesai" /></span>
                </label>
              </div>

              <div className="field-row">
                <label>
                  <span className="field-label">TRAVELERS</span>
                  <span className="field-value"><Users size={16} /><select aria-label="Tipe perjalanan" value={plannerTravelType} onChange={(event) => setPlannerTravelType(event.target.value)}><option value="Solo">Solo</option><option value="Couple">Couple</option><option value="Family">Family</option><option value="Friends">Friends</option></select></span>
                </label>
                <label>
                  <span className="field-label">TRIP STYLE</span>
                  <span className="field-value"><Sparkles size={15} />Rencana fleksibel</span>
                </label>
              </div>

              <Link
                className="button primary full"
                href={plannerHref}
              >
                Plan my trip <ArrowRight size={17} />
              </Link>

              <div className="quick-search">
                POPULAR:{" "}
                {destinations.slice(0, 5).map((destination) => (
                  destination.id ? <Link key={destination.id} href={`/destinations/${encodeURIComponent(destination.id)}`}>{destination.name}</Link> : <span key={destination.name}>{destination.name}</span>
                ))}
              </div>
            </div>
          </div>

          <div className="hero-caption">
          </div>
        </section>

        {/* Trust Strip */}
        <section className="trust-strip">
          <div className="trust-inner">
            <div>
              <span className="trust-symbol">
                <MoonStar size={18} />
              </span>
              <strong>Travel with confidence</strong>
            </div>
            <span>
              Thoughtful details that make every journey feel a little more like
              home.
            </span>
            <div className="trust-checks">
              <span>
                <Check size={14} /> Status halal transparan
              </span>
              <span>
                <Check size={14} /> Fasilitas salat
              </span>
            </div>
          </div>
        </section>

        {/* Curated Destinations Section */}
        <section className="section section-destinations">
          <div className="section-heading">
            <div>
              <div className="eyebrow">CURATED FOR YOU</div>
              <h2>Places to feel at home</h2>
              <p>Thoughtfully chosen destinations, ready for your next story.</p>
            </div>
            <Link href="/destinations" className="text-link">
              Explore all destinations <ArrowRight size={16} />
            </Link>
            <Link href="/map" className="text-link home-map-link"><MapIcon size={15} /> Map tempat
            </Link>
          </div>
          <div className="destination-grid">
            {destinations.slice(0, 3).map((place) => (
              <PlaceCard key={`Destination-${place.id || place.name}`} place={place} />
            ))}
            {destinationsQuery.isLoading && <p>Memuat destinasi…</p>}
            {destinationsQuery.isError && <div className="api-message error-text" role="alert"><p>{destinationsQuery.error.message}</p><button className="button outline" onClick={() => void destinationsQuery.refetch()}>Coba lagi</button></div>}
          </div>
        </section>

        {/* Feature Band */}
        <section className="feature-band">
          <div className="feature-copy">
            <div className="eyebrow light">A BETTER WAY TO WANDER</div>
            <h2>
              The little things
              <br />
              make a big trip.
            </h2>
            <p>
              Find a place to pray between stops. A meal you can order with
              confidence. A stay that understands what matters.
            </p>
            <Link href={plannerHref} className="button light-button">
              Plan your journey <ArrowRight size={16} />
            </Link>
          </div>
          <div className="feature-list">
            {FEATURES.map(({ icon: Icon, title, description }) => (
              <div key={title}>
                <span>
                  <Icon size={19} />
                </span>
                <div>
                  <strong>{title}</strong>
                  <p>{description}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        {/* Explore Section */}
        <section className="section explore-section">
          <div className="section-heading">
            <div>
              <div className="eyebrow">FIND YOUR NEXT FAVORITE</div>
              <h2>A little inspiration</h2>
              <p>Places worth making room for in your itinerary.</p>
            </div>
            <Link href="/destinations" className="text-link">
              Discover more <ArrowRight size={16} />
            </Link>
          </div>

          <div className="explore-controls">
            <div className="tabs">
              {EXPLORE_TABS.map((tabName) => (
                <button
                  key={tabName}
                  type="button"
                  className={tab === tabName ? "tab active" : "tab"}
                  onClick={() => setTab(tabName)}
                >
                  {tabName}
                </button>
              ))}
            </div>
            <div className="explore-search">
              <SearchBox
                placeholder="Find a place..."
                value={query}
                onChange={setQuery}
              />
            </div>
          </div>

          <div className="destination-grid">
            {filteredPlaces.slice(0, 3).map((place) => (
              <PlaceCard key={`${place.category}-${place.id || place.name}`} place={place} />
            ))}
            {filteredPlaces.length === 0 && (destinationsQuery.isLoading || restaurantsQuery.isLoading || hotelsQuery.isLoading || mosquesQuery.isLoading) && <p aria-live="polite">Memuat katalog backend...</p>}
            {filteredPlaces.length === 0 && !(destinationsQuery.isLoading || restaurantsQuery.isLoading || hotelsQuery.isLoading || mosquesQuery.isLoading) && <div className="api-message"><p>{[destinationsQuery, restaurantsQuery, hotelsQuery, mosquesQuery].some((result) => result.isError) ? "Beberapa katalog backend gagal dimuat." : "Belum ada tempat yang cocok dengan pencarian."}</p>{[destinationsQuery, restaurantsQuery, hotelsQuery, mosquesQuery].some((result) => result.isError) && <button className="button outline" onClick={() => { void destinationsQuery.refetch(); void restaurantsQuery.refetch(); void hotelsQuery.refetch(); void mosquesQuery.refetch(); }}>Coba lagi</button>}</div>}
          </div>
        </section>

        {/* Planner CTA */}
        <section className="planner-cta">
          <div className="cta-icon">
            <Sparkles size={23} />
          </div>
          <div>
            <div className="eyebrow">MADE FOR YOUR KIND OF TRIP</div>
            <h2>Your next great trip starts here.</h2>
            <p>Tell us what matters to you. We’ll help make room for all of it.</p>
          </div>
          <Link href={plannerHref} className="button primary">
            Build my itinerary <ArrowRight size={17} />
          </Link>
        </section>
      </main>
      <Footer />
    </>
  );
}
