import { createFileRoute } from "@tanstack/react-router";
import { AdminAccountsPage } from "@/features/admin/admin-accounts-page";
import { RequireAuth } from "@/features/auth/require-auth";

interface AdminAccountsSearch {
  filter?: string;
}

export const Route = createFileRoute("/admin/accounts")({
  validateSearch: (search: Record<string, unknown>): AdminAccountsSearch => ({
    filter: typeof search.filter === "string" && search.filter !== "" ? search.filter : undefined,
  }),
  component: AdminAccountsRoute,
});

function AdminAccountsRoute() {
  return (
    <RequireAuth>
      <AdminAccountsPage />
    </RequireAuth>
  );
}
