import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { LeaderboardsPage } from "@/features/leaderboards/leaderboards-page";

export const Route = createFileRoute("/_portal/leaderboards/")({
  component: LeaderboardsRoute,
});

function LeaderboardsRoute() {
  return (
    <RequireAuth>
      <LeaderboardsPage board="progression" />
    </RequireAuth>
  );
}
