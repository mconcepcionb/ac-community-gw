import { Navigate } from "@tanstack/react-router";
import type { ReactNode } from "react";

import { ErrorState } from "@/components/common/error-state";

import { useSession } from "./use-session";

/**
 * RequireAuth gates a route component behind an authenticated session and
 * sends anonymous visitors through the login redirect. A session fetch error
 * is surfaced instead of being mistaken for "anonymous", which would otherwise
 * bounce the user through the OAuth flow on a transient failure.
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  const { status, error, refetch } = useSession();

  if (status === "loading") {
    return <div className="p-8 text-sm text-neutral-400">Loading session…</div>;
  }
  if (status === "error") {
    return (
      <div className="p-8">
        <ErrorState error={error} onRetry={() => void refetch()} />
      </div>
    );
  }
  if (status === "authenticated") {
    return <>{children}</>;
  }
  return (
    <Navigate
      to="/login"
      search={{ return_to: window.location.pathname + window.location.search }}
    />
  );
}
