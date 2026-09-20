import { useQuery } from "@tanstack/react-query";

import { azerothMeCharactersListOptions } from "@/api";

export interface MyCharactersQuery {
  filter?: string;
  limit?: number;
  offset?: number;
}

/** useMyCharacters lists the signed-in user's own characters. */
export function useMyCharacters({ filter, limit = 100, offset = 0 }: MyCharactersQuery) {
  return useQuery(
    azerothMeCharactersListOptions({
      query: { filter: filter || undefined, limit, offset },
    }),
  );
}
