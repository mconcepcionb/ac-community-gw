import { useQuery } from "@tanstack/react-query";

import { azerothCharactersGetOptions, azerothCharactersListOptions } from "@/api";

export interface CharactersQuery {
  account?: string;
  filter?: string;
  limit?: number;
  offset?: number;
}

/** useCharacters lists characters, optionally filtered by account and name. */
export function useCharacters({ account, filter, limit = 100, offset = 0 }: CharactersQuery) {
  return useQuery(
    azerothCharactersListOptions({
      query: { account: account || undefined, filter: filter || undefined, limit, offset },
    }),
  );
}

/** useCharacter fetches a single character by name. */
export function useCharacter(name: string) {
  return useQuery({
    ...azerothCharactersGetOptions({ path: { name } }),
    enabled: name !== "",
  });
}
