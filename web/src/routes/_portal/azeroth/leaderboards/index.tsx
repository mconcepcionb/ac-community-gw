import { createFileRoute } from "@tanstack/react-router";

import { LeaderboardsPage } from "@/features/leaderboards/leaderboards-page";

export const Route = createFileRoute("/_portal/azeroth/leaderboards/")({
  component: LeaderboardsRoute,
});

function LeaderboardsRoute() {
  return <LeaderboardsPage board="progression" />;
}
