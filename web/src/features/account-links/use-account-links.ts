import { useQuery } from "@tanstack/react-query";

import { azerothAccountLinksGetOptions, azerothAccountLinksListOptions } from "@/api";

/** useAccountLinks lists every community user <-> account link. */
export function useAccountLinks() {
  return useQuery(azerothAccountLinksListOptions());
}

/** useAccountLink fetches the link of one community user. */
export function useAccountLink(userId: string) {
  return useQuery({
    ...azerothAccountLinksGetOptions({ path: { user_id: userId } }),
    enabled: userId !== "",
  });
}
