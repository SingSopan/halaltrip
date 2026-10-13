"use client";

import { useState } from "react";
import Link from "next/link";

const links = [
  ["Destinations", "/destinations"],
  ["Map", "/map"],
  ["Restaurants", "/restaurants"],
  ["Hotels", "/hotels"],
  ["Mosques", "/mosques"],
  ["Trip planner", "/planner"],
];

export function MobileNavigation() {
  const [open, setOpen] = useState(false);

  return (
    <>
      <nav
        className={open ? "nav-links open" : "nav-links"}
        aria-label="Main navigation"
      >
        {links.map(([label, href]) => (
          <Link key={label} href={href} onClick={() => setOpen(false)}>
            {label}
          </Link>
        ))}
      </nav>
      <button
        className="menu-button"
        aria-label={open ? "Close menu" : "Open menu"}
        aria-expanded={open}
        onClick={() => setOpen((isOpen) => !isOpen)}
      >
        {open ? "×" : "☰"}
      </button>
    </>
  );
}
