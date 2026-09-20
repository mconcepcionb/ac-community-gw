import { createFileRoute } from "@tanstack/react-router";

import { UserCharactersPage } from "@/features/characters/user-characters-page";

export const Route = createFileRoute("/_portal/identity/users/$userId/characters")({
  component: UserCharactersRoute,
});

function UserCharactersRoute() {
  const { userId } = Route.useParams();
  return <UserCharactersPage userId={userId} />;
}
