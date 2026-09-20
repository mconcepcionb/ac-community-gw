import { createFileRoute } from "@tanstack/react-router";

import { UsersPage } from "@/features/identity/users-page";

interface UsersSearch {
  filter?: string;
  limit?: number;
  offset?: number;
}

export const Route = createFileRoute("/identity/users/")({
  validateSearch: (search: Record<string, unknown>): UsersSearch => ({
    filter: typeof search.filter === "string" && search.filter !== "" ? search.filter : undefined,
    limit: typeof search.limit === "number" ? search.limit : undefined,
    offset: typeof search.offset === "number" ? search.offset : undefined,
  }),
  component: UsersPage,
});
