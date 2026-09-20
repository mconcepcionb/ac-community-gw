import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { MyCharactersPage } from "@/features/characters/my-characters-page";

export const Route = createFileRoute("/_portal/azeroth/characters/")({
  component: CharactersRoute,
});

function CharactersRoute() {
  return (
    <RequireAuth>
      <MyCharactersPage />
    </RequireAuth>
  );
}
