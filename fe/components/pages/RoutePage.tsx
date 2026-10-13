"use client";

import { Footer, Header } from "@/components/layout";
import { Listing } from "./ListingPage";
import { Detail } from "./DetailPage";
import { Planner } from "./PlannerPage";
import { TripsPage, TripDetail } from "./TripsPage";
import { Itinerary } from "./ItineraryPage";
import { SavedPage } from "./SavedPage";
import { Profile } from "./ProfilePage";
import { AuthPage } from "./AuthPage";
import { MapPage } from "./MapPage";
import { NotFoundPage as NotFound } from "./NotFoundPage";

export function RoutePage({ segments, plannerInitial }: { segments: string[]; plannerInitial?: { city?: string; startDate?: string; endDate?: string; travelType?: string } }) {
  const top = segments[0] ?? "";
  const slug = segments[1];

  const renderContent = () => {
    if (["destinations", "restaurants", "hotels", "mosques"].includes(top)) {
      return slug ? <Detail kind={top} slug={slug} /> : <Listing kind={top} />;
    }
    if (top === "planner" || (top === "trips" && slug === "new")) {
      return <Planner initial={plannerInitial} />;
    }
    if (top === "trips" && slug) {
      return <TripDetail id={slug} />;
    }
    if (top === "trips") {
      return <TripsPage />;
    }
    if (top === "itinerary" && slug) {
      return <Itinerary id={slug} />;
    }
    if (top === "saved") {
      return <SavedPage />;
    }
    if (top === "profile") {
      return <Profile />;
    }
    if (top === "map") return <MapPage />;
    if (top === "login") return <AuthPage initialMode="login" />;
    if (top === "register") return <AuthPage initialMode="register" />;
    return <NotFound />;
  };

  const isPlannerShell =
    top === "planner" || (top === "trips" && slug === "new");

  return (
    <>
      <Header />
      <main className={`page-shell ${isPlannerShell ? "planner-shell" : ""}`}>
        {renderContent()}
      </main>
      <Footer />
    </>
  );
}
