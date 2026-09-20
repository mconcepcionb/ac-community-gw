import { createFileRoute } from "@tanstack/react-router";
import { AccountsPage } from "@/features/accounts/accounts-page";
import { RequireAuth } from "@/features/auth/require-auth";

interface AccountsSearch {
  filter?: string;
  limit?: number;
  offset?: number;
}

export const Route = createFileRoute("/accounts")({
  validateSearch: (search: Record<string, unknown>): AccountsSearch => ({
    filter: typeof search.filter === "string" && search.filter !== "" ? search.filter : undefined,
    limit: typeof search.limit === "number" ? search.limit : undefined,
    offset: typeof search.offset === "number" ? search.offset : undefined,
  }),
  component: AccountsRoute,
});

function AccountsRoute() {
  return (
    <RequireAuth>
      <AccountsPage />
    </RequireAuth>
  );
}
