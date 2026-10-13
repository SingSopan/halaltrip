import type { ReactNode } from "react";

export function DaySection({ children }: { children: ReactNode }) {
  return <section className="day-section">{children}</section>;
}
