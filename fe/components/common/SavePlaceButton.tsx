"use client";

import { useRouter } from "next/navigation";
import { Heart } from "lucide-react";
import { useSavedPlace } from "@/src/hooks/use-saved-place";
import type { Place } from "@/src/types/place";

export function SavePlaceButton({ place, className = "button outline" }: { place: Place; className?: string }) {
  const router = useRouter();
  const { token, canSave, isSaved, mutation } = useSavedPlace(place);
  return <>
    <button type="button" className={className} disabled={mutation.isPending || Boolean(token && !canSave)} title={!canSave ? "Saved places hanya mendukung restoran, masjid, dan hotel lokal." : undefined} onClick={() => {
      if (!token) { router.push("/profile"); return; }
      if (canSave) mutation.mutate();
    }}><Heart size={16} fill={isSaved ? "currentColor" : "none"} /> {isSaved ? "Tersimpan" : token ? "Simpan" : "Masuk untuk menyimpan"}</button>
    {mutation.isError && <p className="error-text" role="alert">{mutation.error.message}</p>}
  </>;
}
