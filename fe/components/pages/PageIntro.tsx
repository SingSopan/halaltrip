interface PageIntroProps { eyebrow: string; title: string; description: string; }

export function PageIntro({ eyebrow, title, description }: PageIntroProps) {
  return (
    <div className="page-intro">
      <div className="eyebrow">{eyebrow}</div>
      <h1>{title}</h1>
      <p>{description}</p>
    </div>
  );
}
