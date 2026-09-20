import { Outlet } from "@tanstack/react-router";

import { AppFooter } from "./app-footer";
import { AppHeader } from "./app-header";

/**
 * PortalLayout is the player-facing shell. It renders for `/` and its children
 * and owns the portal navigation slot.
 */
export function PortalLayout() {
  return (
    <>
      <AppHeader />
      <main className="flex-1">
        <Outlet />
      </main>
      <AppFooter />
    </>
  );
}
