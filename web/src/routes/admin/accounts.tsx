import { createFileRoute } from "@tanstack/react-router";
import { AdminAccountsPage } from "@/features/admin/admin-accounts-page";

interface AdminAccountsSearch {
  filter?: string;
  limit?: number;
  offset?: number;
}

export const Route = createFileRoute("/admin/accounts")({
  validateSearch: (search: Record<string, unknown>): AdminAccountsSearch => ({
    filter: typeof search.filter === "string" && search.filter !== "" ? search.filter : undefined,
    limit: typeof search.limit === "number" ? search.limit : undefined,
    offset: typeof search.offset === "number" ? search.offset : undefined,
  }),
  component: AdminAccountsPage,
});
