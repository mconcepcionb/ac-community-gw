import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { CharacterDetailPage } from "@/features/characters/character-detail-page";

export const Route = createFileRoute("/characters/$name")({
  component: CharacterRoute,
});

function CharacterRoute() {
  const { name } = Route.useParams();
  return (
    <RequireAuth>
      <CharacterDetailPage name={name} />
    </RequireAuth>
  );
}
