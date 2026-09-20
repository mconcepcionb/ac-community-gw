import { createFileRoute } from "@tanstack/react-router";

import { CharacterDetailPage } from "@/features/characters/character-detail-page";

export const Route = createFileRoute("/admin/azeroth/characters/$name")({
  component: CharacterRoute,
});

function CharacterRoute() {
  const { name } = Route.useParams();
  return <CharacterDetailPage name={name} />;
}
