import { useQuery } from "@tanstack/react-query";

import { azerothUserCharactersListOptions } from "@/api";

/** useUserCharacters lists the characters of a linked community user. */
export function useUserCharacters(userId: string) {
  return useQuery({
    ...azerothUserCharactersListOptions({ path: { user_id: userId } }),
    enabled: userId !== "",
  });
}
