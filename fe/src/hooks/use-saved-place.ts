"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getSavedPlaces, removeSavedPlace, savePlace } from "@/src/services";
import { useAuthStore } from "@/src/stores/use-auth-store";
import type { Place } from "@/src/types/place";

export function useSavedPlace(place: Place) {
  const token = useAuthStore((state) => state.token);
  const user = useAuthStore((state) => state.user);
  const queryClient = useQueryClient();
  const kind = place.category === "Restaurant" ? "restaurant" : place.category === "Hotel" ? "hotel" : place.category === "Mosque" ? "mosque" : undefined;
  const canSave = place.source === "directory" && Boolean(kind && place.id);
  const query = useQuery({ queryKey: ["savedPlaces", user?.id], queryFn: () => getSavedPlaces(token!), enabled: Boolean(token && user) });
  const isSaved = query.data?.some((entry) => entry.place.id === place.id) ?? false;
  const mutation = useMutation({
    mutationFn: async () => { if (isSaved) await removeSavedPlace(token!, kind!, place.id); else await savePlace(token!, kind!, place.id); },
    onSuccess: async () => queryClient.invalidateQueries({ queryKey: ["savedPlaces", user?.id] }),
  });
  return { token, kind, canSave, isSaved, query, mutation };
}
