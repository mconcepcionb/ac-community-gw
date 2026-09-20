import { createFileRoute } from "@tanstack/react-router";

import { CharacterBanActions } from "@/features/admin/character-ban-actions";
import { CharacterDetailPage } from "@/features/characters/character-detail-page";

export const Route = createFileRoute("/admin/characters/$name")({
  component: CharacterRoute,
});

function CharacterRoute() {
  const { name } = Route.useParams();
  return <CharacterDetailPage name={name} actions={<CharacterBanActions name={name} />} />;
}
