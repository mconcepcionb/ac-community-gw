import { useQuery } from "@tanstack/react-query";

import { storeProductsGetOptions, storeProductsListOptions } from "@/api";

/** useProducts lists the store catalog. */
export function useProducts() {
  return useQuery(storeProductsListOptions());
}

/** useProduct fetches one catalog product by SKU. */
export function useProduct(sku: string) {
  return useQuery({
    ...storeProductsGetOptions({ path: { sku } }),
    enabled: sku !== "",
  });
}
