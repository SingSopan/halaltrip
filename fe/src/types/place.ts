export type PlaceCategory = "Destination" | "Restaurant" | "Hotel" | "Mosque";

export interface Place {
  /** Local UUID or provider identifier; empty if the provider supplies no stable ID. */
  id: string;
  name: string;
  country: string;
  city: string;
  category: PlaceCategory;
  description?: string;
  address?: string;
  image?: string;
  image_attribution?: string;
  image_license_url?: string;
  score?: number;
  rating?: number;
  review_count?: number;
  cuisine?: string;
  price_range?: string;
  halal_status?: "certified" | "muslim_owned" | "halal_friendly" | "no_pork" | "unknown";
  halal_verification_note?: string;
  facilities?: string[];
  muslim_friendly_features?: { name: string; verified: boolean; verification_note?: string }[];
  latitude?: number | null;
  longitude?: number | null;
  source_url?: string;
  source: "directory" | "provider";
}
