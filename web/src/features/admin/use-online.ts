import { useQuery } from "@tanstack/react-query";

import { azerothOnlineListOptions } from "@/api";

/** useOnline lists the players currently online, refreshing periodically. */
export function useOnline(refetchInterval = 15_000) {
  return useQuery({ ...azerothOnlineListOptions(), refetchInterval });
}
