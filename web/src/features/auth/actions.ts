import type { QueryClient } from "@tanstack/react-query";

import { authLogout, authMeQueryKey } from "@/api";

/**
 * isSafeReturnTo accepts only same-site absolute paths, mirroring the server
 * side guard against open redirects. Backslashes are rejected because browsers
 * normalize them to forward slashes (turning "/\evil.test" into "//evil.test").
 */
export function isSafeReturnTo(value: string): boolean {
  if (
    !value.startsWith("/") ||
    value.includes("\\") ||
    value.includes("\r") ||
    value.includes("\n")
  ) {
    return false;
  }
  if (typeof window === "undefined") {
    return !value.startsWith("//");
  }
  try {
    const url = new URL(value, window.location.origin);
    if (url.origin !== window.location.origin) {
      return false;
    }
    const pathname = decodeURIComponent(url.pathname);
    return pathname.startsWith("/") && !pathname.startsWith("//") && !pathname.includes("\\");
  } catch {
    return false;
  }
}

/**
 * buildLoginUrl returns the gateway login URL, preserving the current location
 * when no safe return_to is provided.
 */
export function buildLoginUrl(returnTo?: string): string {
  const fallback =
    typeof window === "undefined" ? "/" : window.location.pathname + window.location.search;
  const target = returnTo && isSafeReturnTo(returnTo) ? returnTo : fallback;
  return `/api/v1/auth/discord/login?return_to=${encodeURIComponent(target)}`;
}

/** login performs a full-page redirect to the Discord OAuth2 flow. */
export function login(returnTo?: string): void {
  window.location.assign(buildLoginUrl(returnTo));
}

/** logout revokes the session and refreshes the cached principal. */
export async function logout(queryClient: QueryClient): Promise<void> {
  await authLogout({ throwOnError: true });
  queryClient.setQueryData(authMeQueryKey(), undefined);
  await queryClient.invalidateQueries({ queryKey: authMeQueryKey() });
}
