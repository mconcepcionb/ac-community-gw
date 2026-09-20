import { useQuery } from "@tanstack/react-query";

import { identityUsersListOptions } from "@/api";

export interface UsersQuery {
  filter?: string;
  limit?: number;
  offset?: number;
}

/** useUsers lists community users, keyed by the query parameters. */
export function useUsers({ filter, limit = 50, offset = 0 }: UsersQuery) {
  return useQuery(
    identityUsersListOptions({
      query: { filter: filter || undefined, limit, offset },
    }),
  );
}
