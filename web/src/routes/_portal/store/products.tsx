import { createFileRoute, Outlet } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";

export const Route = createFileRoute("/_portal/store/products")({
  component: StoreProductsLayout,
});

function StoreProductsLayout() {
  return (
    <RequireAuth>
      <Outlet />
    </RequireAuth>
  );
}
