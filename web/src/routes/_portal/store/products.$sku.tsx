import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { ProductDetailPage } from "@/features/store/product-detail-page";

export const Route = createFileRoute("/_portal/store/products/$sku")({
  component: ProductRoute,
});

function ProductRoute() {
  const { sku } = Route.useParams();
  return (
    <RequireAuth>
      <ProductDetailPage sku={sku} />
    </RequireAuth>
  );
}
