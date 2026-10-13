"use client";

import { useState } from "react";
import { ArrowRight, MapPin, Search } from "lucide-react";

const markers = [
  { x: 20, y: 28, type: "Mosque" },
  { x: 62, y: 20, type: "Halal food" },
  { x: 76, y: 61, type: "Stay" },
  { x: 42, y: 72, type: "Explore" },
  { x: 88, y: 34, type: "Halal food" },
];

export function MapPlaceholder() {
  const [selected, setSelected] = useState<string | null>(null);

  return (
    <div className="map-placeholder" role="region" aria-label="Map preview">
      <div className="map-search">
        <Search size={16} /> Search this area
      </div>
      <div className="map-water" />
      <div className="map-road road-one" />
      <div className="map-road road-two" />
      <div className="map-road road-three" />
      {markers.map((marker, index) => (
        <button
          key={`${marker.type}-${index}`}
          className={`map-marker marker-${index}`}
          style={{ left: `${marker.x}%`, top: `${marker.y}%` }}
          aria-label={`Show ${marker.type} on map`}
          onClick={() =>
            setSelected((current) =>
              current === marker.type ? null : marker.type,
            )
          }
        >
          <MapPin size={14} fill="currentColor" />
        </button>
      ))}
      {selected && (
        <div className="map-popover">
          <strong>{selected} nearby</strong>
          <span>
            Open details <ArrowRight size={13} />
          </span>
        </div>
      )}
      <span className="map-attribution">Map preview</span>
    </div>
  );
}
