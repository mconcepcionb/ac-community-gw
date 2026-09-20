import { createFileRoute } from "@tanstack/react-router";

import { LeaderboardsPage } from "@/features/leaderboards/leaderboards-page";

export const Route = createFileRoute("/_portal/leaderboards/$board")({
  component: LeaderboardRoute,
});

function LeaderboardRoute() {
  const { board } = Route.useParams();
  return <LeaderboardsPage board={board} />;
}
