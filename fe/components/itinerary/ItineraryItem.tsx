import type { ReactNode } from "react";

export type ItineraryActivity = {
  time: string;
  title: string;
  description: string;
  type: "stay" | "move" | "see" | "eat" | "pray";
  icon: ReactNode;
};

export function ItineraryItem({ activity }: { activity: ItineraryActivity }) {
  return (
    <div className="timeline-row">
      <time>{activity.time}</time>
      <div className={`timeline-marker ${activity.type}`} />
      <article className="activity-card">
        <span className={`activity-icon ${activity.type}`}>{activity.icon}</span>
        <div>
          <h3>{activity.title}</h3>
          <p>{activity.description}</p>
        </div>
        <button aria-label={`More options for ${activity.title}`}>···</button>
      </article>
    </div>
  );
}
