import { createFileRoute } from "@tanstack/react-router";

import { CharactersPage } from "@/features/characters/characters-page";

interface CharactersSearch {
  account?: string;
  filter?: string;
  limit?: number;
  offset?: number;
}

export const Route = createFileRoute("/admin/characters/")({
  validateSearch: (search: Record<string, unknown>): CharactersSearch => ({
    account:
      typeof search.account === "string" && search.account !== "" ? search.account : undefined,
    filter: typeof search.filter === "string" && search.filter !== "" ? search.filter : undefined,
    limit: typeof search.limit === "number" ? search.limit : undefined,
    offset: typeof search.offset === "number" ? search.offset : undefined,
  }),
  component: CharactersPage,
});
