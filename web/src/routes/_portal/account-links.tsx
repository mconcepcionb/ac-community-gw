import { createFileRoute, Outlet } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";

export const Route = createFileRoute("/_portal/account-links")({
  component: AccountLinksLayout,
});

function AccountLinksLayout() {
  return (
    <RequireAuth>
      <Outlet />
    </RequireAuth>
  );
}
