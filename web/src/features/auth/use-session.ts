import { useQuery } from "@tanstack/react-query";

import { authMeOptions } from "@/api";
import { isUnauthorized } from "@/api/errors";

export type SessionStatus = "loading" | "anonymous" | "authenticated" | "error";

/**
 * useSession resolves the current principal from GET /api/v1/me.
 *
 * A 401 means "anonymous", not an error. Any other failure surfaces through
 * `error` so the UI can react to it.
 */
export function useSession() {
  const query = useQuery({ ...authMeOptions(), retry: false });
  const anonymous = isUnauthorized(query.error);

  let status: SessionStatus;
  if (query.isPending) {
    status = "loading";
  } else if (anonymous) {
    status = "anonymous";
  } else if (query.isError) {
    status = "error";
  } else {
    status = "authenticated";
  }

  return {
    status,
    principal: query.data,
    error: status === "error" ? query.error : undefined,
    refetch: query.refetch,
  };
}
