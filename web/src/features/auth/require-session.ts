import type { QueryClient } from "@tanstack/react-query";

import { isUnauthorized } from "@/api/errors";
import { sessionOptions } from "./api";

/**
 * requireSessionFor resolves the current principal for a route loader.
 *
 * It returns `null` for an anonymous session and rethrows anything else, so a
 * transient failure surfaces through the route error state instead of being
 * mistaken for "signed out".
 */
export async function requireSessionFor(queryClient: QueryClient) {
  try {
    return await queryClient.ensureQueryData({ ...sessionOptions(), retry: false });
  } catch (error) {
    if (isUnauthorized(error)) {
      return null;
    }
    throw error;
  }
}
