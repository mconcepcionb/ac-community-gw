import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { ItemDetailPage } from "@/features/items/item-detail-page";

export const Route = createFileRoute("/items/$entry")({
  component: ItemRoute,
});

function ItemRoute() {
  const { entry } = Route.useParams();
  return (
    <RequireAuth>
      <ItemDetailPage entry={Number(entry)} />
    </RequireAuth>
  );
}
