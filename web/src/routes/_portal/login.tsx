import { createFileRoute, Navigate } from "@tanstack/react-router";
import { useEffect, useRef } from "react";

import { isSafeReturnTo, login } from "@/features/auth/actions";
import { useSession } from "@/features/auth/use-session";

export const Route = createFileRoute("/_portal/login")({
  validateSearch: (search: Record<string, unknown>): { return_to?: string } => ({
    return_to: typeof search.return_to === "string" ? search.return_to : undefined,
  }),
  component: Login,
});

function Login() {
  const { return_to } = Route.useSearch();
  const { status } = useSession();
  const started = useRef(false);

  useEffect(() => {
    if (status !== "anonymous" || started.current) {
      return;
    }
    started.current = true;
    login(return_to);
  }, [return_to, status]);

  if (status === "authenticated") {
    const target = return_to && isSafeReturnTo(return_to) ? return_to : "/";
    return <Navigate to={target} />;
  }

  return (
    <div className="mx-auto max-w-3xl p-8 text-sm text-neutral-400">Redirecting to Discord…</div>
  );
}
