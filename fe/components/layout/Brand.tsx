import Link from "next/link";

export function Brand() {
  return (
    <Link href="/" className="brand" aria-label="HalalTrip home">
      <span className="brand-mark">h</span>
      <span>
        halal<span className="brand-trip">trip</span>
      </span>
    </Link>
  );
}
