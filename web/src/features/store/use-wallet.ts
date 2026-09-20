import { useQuery } from "@tanstack/react-query";

import { storeOrdersListOptions, storeWalletGetOptions } from "@/api";

/** useWallet returns the authenticated user's point balance. */
export function useWallet() {
  return useQuery(storeWalletGetOptions());
}

/** useOrders lists the authenticated user's orders. */
export function useOrders({ limit = 50, offset = 0 }: { limit?: number; offset?: number } = {}) {
  return useQuery(storeOrdersListOptions({ query: { limit, offset } }));
}
