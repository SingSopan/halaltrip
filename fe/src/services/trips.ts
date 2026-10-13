import { apiRequest } from "./client";
import type { ItineraryInput, ItineraryItem, ItineraryPatch, Trip, TripInput, TripPatch } from "@/src/types/api";

export async function listTrips(token: string) {
  return (await apiRequest<{ data: Trip[] }>("/trips", { token })).data;
}
export async function getTrip(token: string, tripId: string) {
  return (await apiRequest<{ data: Trip }>(`/trips/${encodeURIComponent(tripId)}`, { token })).data;
}
export async function createTrip(token: string, input: TripInput) {
  return (await apiRequest<{ data: Trip }>("/trips", { method: "POST", token, body: input })).data;
}
export async function updateTrip(token: string, tripId: string, input: TripPatch) {
  return (await apiRequest<{ data: Trip }>(`/trips/${encodeURIComponent(tripId)}`, { method: "PATCH", token, body: input })).data;
}
export function deleteTrip(token: string, tripId: string) {
  return apiRequest<void>(`/trips/${encodeURIComponent(tripId)}`, { method: "DELETE", token });
}
export async function listItinerary(token: string, tripId: string) {
  return (await apiRequest<{ data: ItineraryItem[] }>(`/trips/${encodeURIComponent(tripId)}/itinerary-items`, { token })).data;
}
export async function createItineraryItem(token: string, tripId: string, input: ItineraryInput) {
  return (await apiRequest<{ data: ItineraryItem }>(`/trips/${encodeURIComponent(tripId)}/itinerary-items`, { method: "POST", token, body: input })).data;
}
export async function updateItineraryItem(token: string, tripId: string, itemId: string, input: ItineraryPatch) {
  return (await apiRequest<{ data: ItineraryItem }>(`/trips/${encodeURIComponent(tripId)}/itinerary-items/${encodeURIComponent(itemId)}`, { method: "PATCH", token, body: input })).data;
}
export function deleteItineraryItem(token: string, tripId: string, itemId: string) {
  return apiRequest<void>(`/trips/${encodeURIComponent(tripId)}/itinerary-items/${encodeURIComponent(itemId)}`, { method: "DELETE", token });
}
export function reorderItinerary(token: string, tripId: string, item_ids: string[]) {
  return apiRequest<void>(`/trips/${encodeURIComponent(tripId)}/itinerary-items/order`, { method: "PUT", token, body: { item_ids } });
}
