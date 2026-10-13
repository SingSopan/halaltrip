import { apiRequest } from "./client";
import { normalizeProviderDestination, recordsFromPayload } from "./recommendations";
import type { Pagination } from "@/src/types/api";
import type { Place, PlaceCategory } from "@/src/types/place";

export type PlaceKind = "destinations" | "restaurants" | "hotels" | "mosques";
export type PlaceFilters = {
  q?: string; city?: string; country?: string; cuisine?: string; halal_status?: string;
  price_range?: string; facility?: string; feature?: string; page?: number; page_size?: number;
};

type DirectoryPlace = Omit<Place, "category" | "source"> & { image_url?: string };
type LocalDestination = Omit<Place, "category" | "source" | "image"> & { image_url?: string; slug: string };
const categoryByKind: Record<PlaceKind, PlaceCategory> = {
  destinations: "Destination", restaurants: "Restaurant", hotels: "Hotel", mosques: "Mosque",
};

function queryString(filters: PlaceFilters = {}) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) if (value !== undefined && value !== "") params.set(key, String(value));
  const query = params.toString();
  return query ? `?${query}` : "";
}

export async function listPlaces(kind: PlaceKind, filters: PlaceFilters = {}) {
  if (kind === "destinations") {
    const [providerResult, localResult] = await Promise.allSettled([
      apiRequest<unknown>("/destinations"),
      apiRequest<{ data: LocalDestination[] }>(`/catalog/destinations${queryString({ city: filters.city })}`),
    ]);
    const providerItems = providerResult.status === "fulfilled"
      ? recordsFromPayload(providerResult.value).flatMap((raw) => {
          const place = normalizeProviderDestination(raw);
          return place ? [place] : [];
        })
      : [];
    const localItems = localResult.status === "fulfilled"
      ? localResult.value.data.map((place) => ({ ...place, image: place.image_url, category: "Destination" as const, source: "directory" as const }))
      : [];
    if (providerResult.status === "rejected" && localResult.status === "rejected") throw providerResult.reason;
    const seen = new Set<string>();
    const items = [...localItems, ...providerItems].filter((place) => {
      const key = `${place.name.toLocaleLowerCase()}|${place.city.toLocaleLowerCase()}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    });
    const query = filters.q?.trim().toLocaleLowerCase();
    const city = filters.city?.trim().toLocaleLowerCase();
    const filteredItems = items.filter((place) =>
      (!query || `${place.name} ${place.city} ${place.country}`.toLocaleLowerCase().includes(query)) &&
      (!city || place.city.toLocaleLowerCase().includes(city)),
    );
    return { items: filteredItems, pagination: undefined as Pagination | undefined };
  }
  const response = await apiRequest<{ data: DirectoryPlace[]; pagination: Pagination }>(`/${kind}${queryString(filters)}`);
  return {
    items: response.data.map((place) => ({ ...place, image: place.image_url ?? place.image, category: categoryByKind[kind], source: "directory" as const })),
    pagination: response.pagination,
  };
}

export async function getPlace(kind: PlaceKind, id: string): Promise<Place> {
  if (kind === "destinations") {
    if (/^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(id)) {
      const response = await apiRequest<{ data: LocalDestination }>(`/catalog/destinations/${encodeURIComponent(id)}`);
      return { ...response.data, image: response.data.image_url, category: "Destination", source: "directory" };
    }
    const payload = await apiRequest<unknown>(`/destinations/${encodeURIComponent(id)}`);
    const place = normalizeProviderDestination(payload);
    if (!place) throw new Error("Provider mengirim format detail destinasi yang tidak dikenali.");
    return place;
  }
  const response = await apiRequest<{ data: DirectoryPlace }>(`/${kind}/${encodeURIComponent(id)}`);
  return { ...response.data, image: response.data.image_url ?? response.data.image, category: categoryByKind[kind], source: "directory" };
}

export async function getSavedPlaces(token: string) {
  const response = await apiRequest<{ data: { place_type: "restaurant" | "mosque" | "hotel"; place: DirectoryPlace; saved_at: string }[] }>("/saved-places", { token });
  return response.data.map((saved) => ({
    ...saved,
    place: { ...saved.place, image: saved.place.image_url ?? saved.place.image, category: categoryByKind[`${saved.place_type}s` as PlaceKind], source: "directory" as const },
  }));
}

export async function getLocalDestinations(city = "") {
  const response = await apiRequest<{ data: LocalDestination[] }>(`/catalog/destinations${queryString({ city })}`);
  return response.data.map((place) => ({ ...place, image: place.image_url, category: "Destination" as const, source: "directory" as const }));
}

export function savePlace(token: string, place_type: "restaurant" | "mosque" | "hotel", place_id: string) {
  return apiRequest<{ place_type: string; place_id: string }>("/saved-places", { method: "POST", token, body: { place_type, place_id } });
}

export function removeSavedPlace(token: string, place_type: "restaurant" | "mosque" | "hotel", place_id: string) {
  return apiRequest<void>(`/saved-places/${place_type}/${encodeURIComponent(place_id)}`, { method: "DELETE", token });
}
