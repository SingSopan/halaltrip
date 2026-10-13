"use client";

import Link from "next/link";
import { ArrowLeft } from "lucide-react";

export function NotFoundPage() {
  return (
    <div className="not-found">
      <span>404</span>
      <h1>We can’t find that place.</h1>
      <p>It may have moved, or it may be waiting for you somewhere else.</p>
      <Link href="/" className="button primary">
        <ArrowLeft size={15} /> Back home
      </Link>
    </div>
  );
}
