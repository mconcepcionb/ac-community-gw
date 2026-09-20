import { createFileRoute, Outlet } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";

export const Route = createFileRoute("/_portal/identity/users")({
  component: UsersLayout,
});

function UsersLayout() {
  return (
    <RequireAuth>
      <Outlet />
    </RequireAuth>
  );
}
