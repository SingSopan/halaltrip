export function HalalBadge({ children = "Bersertifikat halal" }: { children?: string }) {
  return <span className="verified">✓ {children}</span>;
}
