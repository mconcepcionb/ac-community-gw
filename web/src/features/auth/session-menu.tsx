import { useQueryClient } from "@tanstack/react-query";

import { login, logout } from "./actions";
import { useSession } from "./use-session";

/** SessionMenu shows the login action or the authenticated principal. */
export function SessionMenu() {
  const queryClient = useQueryClient();
  const { status, principal } = useSession();

  if (status === "loading") {
    return <span className="text-xs text-neutral-500">Loading…</span>;
  }

  if (status !== "authenticated") {
    return (
      <button
        type="button"
        className="rounded bg-indigo-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-indigo-500"
        onClick={() => login()}
      >
        Login with Discord
      </button>
    );
  }

  return (
    <div className="flex items-center gap-3">
      <span className="text-xs text-neutral-400">{principal?.discord_id}</span>
      <button
        type="button"
        className="rounded border border-neutral-700 px-3 py-1.5 text-xs hover:bg-neutral-800"
        onClick={() => void logout(queryClient)}
      >
        Logout
      </button>
    </div>
  );
}
