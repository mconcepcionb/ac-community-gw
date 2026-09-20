import { useQuery } from "@tanstack/react-query";

import { azerothInfoStatusOptions } from "@/api";

/** useAzerothStatus fetches the AzerothCore server status snapshot. */
export function useAzerothStatus() {
  return useQuery({ ...azerothInfoStatusOptions(), retry: false });
}
