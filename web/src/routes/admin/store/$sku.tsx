import { createFileRoute } from "@tanstack/react-router";

import { ProductDetailPage } from "@/features/store/product-detail-page";

export const Route = createFileRoute("/admin/store/$sku")({
  component: ProductRoute,
});

function ProductRoute() {
  const { sku } = Route.useParams();
  return <ProductDetailPage sku={sku} />;
}
