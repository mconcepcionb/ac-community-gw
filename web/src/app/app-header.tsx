import { Link } from "@tanstack/react-router";

import { AppNav } from "./app-nav";

export function AppHeader() {
  return (
    <header className="border-b border-neutral-800">
      <div className="mx-auto flex max-w-5xl items-center gap-4 p-4">
        <Link to="/" className="font-semibold">
          ac-community-gw
        </Link>
        <AppNav />
      </div>
    </header>
  );
}
