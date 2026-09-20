import type { QueryClient } from "@tanstack/react-query";
import { createRootRouteWithContext, Outlet } from "@tanstack/react-router";

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()({
  component: RootLayout,
});

/**
 * RootLayout is the minimal document wrapper. Surface chrome lives in the
 * portal and console layouts so the two never share a navigation slot.
 */
function RootLayout() {
  return (
    <div className="flex min-h-dvh flex-col bg-neutral-950 text-neutral-100">
      <Outlet />
    </div>
  );
}
