import Link from "next/link";
import { Heart } from "lucide-react";
import { Brand } from "./Brand";
import { MobileNavigation } from "./MobileNavigation";
import { useAuthStore } from "@/src/stores/use-auth-store";

export function Navbar() {
  const user = useAuthStore((state) => state.user);
  return (
    <header className="site-header">
      <div className="nav-wrap">
        <Brand />
        <MobileNavigation />
        <div className="nav-actions">
          <Link href="/saved" className="saved-link">
            <Heart size={17} /> <span>Saved</span>
          </Link>
          <Link href={user ? "/profile" : "/login"} className={user ? "avatar" : "button outline nav-auth-link"} aria-label={user ? "Profile" : "Masuk"}>
            {user ? user.name.slice(0, 1).toUpperCase() : "Masuk"}
          </Link>
        </div>
      </div>
    </header>
  );
}
