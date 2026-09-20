import { useQuery } from "@tanstack/react-query";

import { azerothAccountsListOptions } from "@/api";

export interface AccountsQuery {
  filter?: string;
  limit?: number;
  offset?: number;
}

/** useAccounts lists AzerothCore login accounts. */
export function useAccounts({ filter, limit = 50, offset = 0 }: AccountsQuery) {
  return useQuery(
    azerothAccountsListOptions({
      query: { filter: filter || undefined, limit, offset },
    }),
  );
}
