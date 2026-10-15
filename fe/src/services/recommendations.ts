import { apiRequest } from "./client";
import type { Activity, Interest, TravelType } from "@/src/types/api";
import type { Place } from "@/src/types/place";

export type RecommendationInput = { interests: Interest[]; activity: Activity; travel_type: TravelType; top_n?: number };
export type RecommendationOption = { key: string; label: string; country?: string; providerId?: string; score?: number; description?: string; raw: unknown };

function object(value: unknown): Record<string, unknown> | undefined {
  return value !== null && typeof value === "object" && !Array.isArray(value) ? value as Record<string, unknown> : undefined;
}

function stringField(record: Record<string, unknown>, keys: string[]) {
  for (const key of keys) {
    const value = record[key];
    if (typeof value === "string" && value.trim()) return value.trim();
    if (typeof value === "number") return String(value);
  }
  return undefined;
}

function scoreField(record: Record<string, unknown>) {
  for (const key of ["score", "recommendation_score", "final_score"]) {
    const value = record[key];
    if (typeof value === "number" && Number.isFinite(value)) return value;
    if (typeof value === "string" && value.trim() && Number.isFinite(Number(value))) return Number(value);
  }
  return undefined;
}

export function recordsFromPayload(payload: unknown): unknown[] {
  if (Array.isArray(payload)) return payload;
  const record = object(payload);
  if (!record) throw new Error("Format respons dari backend/provider tidak dikenali.");
  for (const key of ["recommendations", "cities", "destinations", "results", "items", "data"]) {
    if (record[key] !== undefined) return recordsFromPayload(record[key]);
  }
  if (stringField(record, ["name", "city", "destination", "title", "destination_name"])) return [record];
  throw new Error("Backend/provider mengirim daftar dengan format yang tidak dikenali.");
}

export function normalizeOptions(payload: unknown, labelKeys: string[]): RecommendationOption[] {
  return recordsFromPayload(payload).flatMap<RecommendationOption>((raw, index) => {
    if (typeof raw === "string" && raw.trim()) return [{ key: raw.trim(), label: raw.trim(), raw }];
    const record = object(raw);
    if (!record) return [];
    const label = stringField(record, labelKeys);
    if (!label) return [];
    return [{
      key: stringField(record, ["id", "slug", "city_id", "destination_id"]) ?? `${label}-${index}`,
      label,
      country: stringField(record, ["country", "country_name"]),
      providerId: stringField(record, ["id", "destination_id", "place_id", "slug"]),
      score: scoreField(record),
      description: stringField(record, ["description", "summary", "reason"]),
      raw,
    }];
  });
}

export function normalizeProviderDestination(raw: unknown): Place | undefined {
  let record = object(raw);
  if (record) {
    for (const key of ["destination", "data"]) {
      const nested = object(record[key]);
      if (nested) { record = nested; break; }
    }
  }
  if (!record) return undefined;
  const name = stringField(record, ["name", "destination_name", "title", "destination"]);
  if (!name) return undefined;
  return {
    id: stringField(record, ["id", "destination_id", "place_id", "slug"]) ?? "",
    name,
    city: stringField(record, ["city", "city_name"]) ?? "",
    country: stringField(record, ["country", "country_name"]) ?? "",
    description: stringField(record, ["description", "summary", "reason"]),
    image: stringField(record, ["image_url", "image", "photo_url", "thumbnail"]),
    score: scoreField(record),
    rating: typeof record.rating === "number" ? record.rating : undefined,
    review_count: typeof record.review_count === "number" ? record.review_count : undefined,
    category: "Destination",
    source: "provider",
  };
}

export async function recommendCities(input: RecommendationInput, token?: string | null) {
  const response = await apiRequest<unknown>("/recommendations/cities", { method: "POST", token, body: input });
  return normalizeOptions(response, ["city", "name", "city_name", "destination", "title"]);
}

export async function recommendDestinations(city: string, input: RecommendationInput, token?: string | null) {
  const response = await apiRequest<unknown>("/recommendations/destinations", { method: "POST", token, body: { ...input, city } });
  return normalizeOptions(response, ["name", "destination_name", "destination", "title"]);
}
