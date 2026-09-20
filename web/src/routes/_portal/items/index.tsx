import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { ItemsPage } from "@/features/items/items-page";

interface ItemsSearch {
  filter?: string;
  class?: number;
  limit?: number;
  offset?: number;
}

export const Route = createFileRoute("/_portal/items/")({
  validateSearch: (search: Record<string, unknown>): ItemsSearch => ({
    filter: typeof search.filter === "string" && search.filter !== "" ? search.filter : undefined,
    class: typeof search.class === "number" ? search.class : undefined,
    limit: typeof search.limit === "number" ? search.limit : undefined,
    offset: typeof search.offset === "number" ? search.offset : undefined,
  }),
  component: ItemsRoute,
});

function ItemsRoute() {
  return (
    <RequireAuth>
      <ItemsPage />
    </RequireAuth>
  );
}
