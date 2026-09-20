import { createFileRoute } from "@tanstack/react-router";

import { ItemDetailPage } from "@/features/items/item-detail-page";

export const Route = createFileRoute("/admin/items/$entry")({
  component: ItemRoute,
});

function ItemRoute() {
  const { entry } = Route.useParams();
  return <ItemDetailPage entry={Number(entry)} />;
}
