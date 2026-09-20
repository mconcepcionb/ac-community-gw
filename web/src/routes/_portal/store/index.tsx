import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { StorefrontPage } from "@/features/store/storefront-page";

export const Route = createFileRoute("/_portal/store/")({
  component: StorefrontRoute,
});

function StorefrontRoute() {
  return (
    <RequireAuth>
      <StorefrontPage />
    </RequireAuth>
  );
}
