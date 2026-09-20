import { Link, Outlet } from "@tanstack/react-router";

import { AppFooter } from "./app-footer";
import { ConsoleNav } from "./console-nav";

/**
 * ConsoleLayout is the staff-facing shell. It renders for `/admin` and its
 * children and owns the console navigation slot.
 */
export function ConsoleLayout() {
  return (
    <>
      <header className="border-b border-neutral-800">
        <div className="mx-auto flex max-w-6xl items-center gap-4 p-4">
          <Link to="/admin" className="font-semibold">
            ac-community-gw
            <span className="ml-2 text-xs text-muted-foreground">console</span>
          </Link>
          <ConsoleNav />
        </div>
      </header>
      <main className="flex-1">
        <Outlet />
      </main>
      <AppFooter />
    </>
  );
}
