import { listPlaces, getPlace, type PlaceFilters, type PlaceKind } from "./catalog";

export { APIError, apiRequest } from "./client";
export { getMe, getPreferences, login, register, savePreferences, updateMe } from "./auth";
export { getPlace, getSavedPlaces, listPlaces, removeSavedPlace, savePlace } from "./catalog";
export { normalizeOptions, recommendCities, recommendDestinations } from "./recommendations";
export type { RecommendationInput, RecommendationOption } from "./recommendations";
export {
  createItineraryItem,
  createTrip,
  deleteItineraryItem,
  deleteTrip,
  getTrip,
  listItinerary,
  listTrips,
  reorderItinerary,
  updateItineraryItem,
  updateTrip,
} from "./trips";

export async function getDestinations(filters?: PlaceFilters) {
  return (await listPlaces("destinations", filters)).items;
}
export async function getRestaurants(filters?: PlaceFilters) {
  return (await listPlaces("restaurants", filters)).items;
}
export async function getHotels(filters?: PlaceFilters) {
  return (await listPlaces("hotels", filters)).items;
}
export async function getMosques(filters?: PlaceFilters) {
  return (await listPlaces("mosques", filters)).items;
}
export async function getPlaceByKind(kind: PlaceKind, id: string) {
  return getPlace(kind, id);
}
