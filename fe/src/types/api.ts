export type User = { id: string; name: string; email: string; created_at: string; updated_at: string };
export type Preferences = { interests: Interest[]; activity: Activity; travel_type: TravelType };
export type Interest = "culinary" | "culture" | "nature" | "shopping" | "adventure" | "history";
export type Activity = "Relaxing" | "Exploration" | "Adventure";
export type TravelType = "Solo" | "Couple" | "Family" | "Friends";
export type AuthResponse = { user: User; access_token: string; token_type: "Bearer"; expires_in: number };
export type TripStatus = "planning" | "active" | "completed" | "cancelled";
export type Trip = {
  id: string; title: string; start_date: string; end_date: string; origin?: string;
  notes?: string; status: TripStatus; created_at: string; updated_at: string;
};
export type TripInput = Pick<Trip, "title" | "start_date" | "end_date"> & Partial<Pick<Trip, "origin" | "notes" | "status">>;
export type TripPatch = Partial<TripInput>;
export type ItineraryItem = {
  id: string; trip_id: string; title: string; description?: string;
  place_type: "destination" | "restaurant" | "mosque" | "hotel" | "custom_activity";
  external_place_id?: string | null; scheduled_date?: string | null; start_time?: string | null;
  end_time?: string | null; sequence: number; transport_mode?: string; notes?: string;
  created_at: string; updated_at: string;
};
export type ItineraryInput = Omit<ItineraryItem, "id" | "trip_id" | "sequence" | "created_at" | "updated_at">;
export type ItineraryPatch = Partial<ItineraryInput>;
export type SavedPlace = {
  place_type: "restaurant" | "mosque" | "hotel";
  place: import("./place").Place;
  saved_at: string;
};
export type Pagination = { page: number; page_size: number; total: number; total_pages: number };
