import { useQuery } from "@tanstack/react-query";

import { azerothAccountsGetOptions, azerothAccountsListOptions } from "@/api";

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

/** useAccount fetches one AzerothCore account with its owner and pending claim. */
export function useAccount(username: string) {
  return useQuery({
    ...azerothAccountsGetOptions({ path: { username } }),
    enabled: username !== "",
  });
}
