import { useQuery } from "@tanstack/react-query";

import { azerothItemsGetOptions, azerothItemsListOptions } from "@/api";

export interface ItemsQuery {
  filter?: string;
  classId?: number;
  limit?: number;
  offset?: number;
}

/** useItems searches the AzerothCore item catalog. */
export function useItems({ filter, classId, limit = 50, offset = 0 }: ItemsQuery) {
  return useQuery(
    azerothItemsListOptions({
      query: { filter: filter || undefined, class: classId, limit, offset },
    }),
  );
}

/** isValidEntry reports whether an item entry id is usable in a request. */
export function isValidEntry(entry: number): boolean {
  return Number.isInteger(entry) && entry > 0;
}

/** useItem fetches one item template by entry id. */
export function useItem(entry: number) {
  return useQuery({
    ...azerothItemsGetOptions({ path: { entry } }),
    enabled: isValidEntry(entry),
  });
}
