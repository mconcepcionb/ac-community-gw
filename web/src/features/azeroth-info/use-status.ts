import { useQuery } from "@tanstack/react-query";

import { azerothPublicStatusOptions } from "@/api";

/** useAzerothStatus fetches the public AzerothCore server status snapshot. */
export function useAzerothStatus() {
  return useQuery({ ...azerothPublicStatusOptions(), retry: false });
}
