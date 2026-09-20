import { createFileRoute } from "@tanstack/react-router";

import { StoreProductsPage } from "@/features/store/store-products-page";

export const Route = createFileRoute("/store/products/")({
  component: StoreProductsPage,
});
