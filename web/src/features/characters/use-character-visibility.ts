import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import {
  azerothMeCharactersVisibilityListOptions,
  azerothMeCharactersVisibilityListQueryKey,
  azerothMeCharactersVisibilitySetMutation,
} from "@/api";

/** useCharacterVisibility lists the public flag of the user's characters. */
export function useCharacterVisibility() {
  return useQuery(azerothMeCharactersVisibilityListOptions());
}

/** useSetCharacterVisibility toggles a character's public flag. */
export function useSetCharacterVisibility() {
  const queryClient = useQueryClient();
  return useMutation({
    ...azerothMeCharactersVisibilitySetMutation(),
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: azerothMeCharactersVisibilityListQueryKey(),
      }),
  });
}
