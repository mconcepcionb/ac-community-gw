import { getRouteApi, Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useEffect, useState } from "react";

import type { User } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useUsers } from "./use-users";

const route = getRouteApi("/_portal/identity/users/");

const columns: ColumnDef<User, unknown>[] = [
  { accessorKey: "username", header: "Username" },
  { accessorKey: "global_name", header: "Global name" },
  { accessorKey: "display_name", header: "Display name" },
  { accessorKey: "discord_id", header: "Discord ID" },
  {
    accessorKey: "user_id",
    header: "User ID",
    cell: ({ row }) => (
      <Link
        to="/identity/users/$userId/characters"
        params={{ userId: row.original.user_id ?? "" }}
        className="text-blue-400 underline"
      >
        {row.original.user_id}
      </Link>
    ),
  },
  { accessorKey: "created_at", header: "Created" },
];

export function UsersPage() {
  const search = route.useSearch();
  const navigate = route.useNavigate();
  const limit = search.limit ?? 50;
  const offset = search.offset ?? 0;

  const [filterInput, setFilterInput] = useState(search.filter ?? "");
  const debouncedFilter = useDebouncedValue(filterInput, 300);

  useEffect(() => {
    if ((search.filter ?? "") === debouncedFilter) {
      return;
    }
    navigate({
      search: { ...search, filter: debouncedFilter || undefined, offset: undefined },
      replace: true,
    });
  }, [debouncedFilter, search, navigate]);

  const query = useUsers({ filter: search.filter, limit, offset });
  const users = query.data?.users ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="Community users" description="Users provisioned from Discord sign-in." />

      <div className="mb-4 max-w-sm">
        <Input
          placeholder="Search users…"
          aria-label="Search users"
          value={filterInput}
          onChange={(event) => setFilterInput(event.target.value)}
        />
      </div>

      <DataTable
        columns={columns}
        data={users}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No users"
      />

      <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
        <span>
          Showing {users.length === 0 ? 0 : offset + 1}–{offset + users.length}
        </span>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={offset === 0}
            onClick={() => navigate({ search: { ...search, offset: Math.max(0, offset - limit) } })}
          >
            Previous
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={users.length < limit}
            onClick={() => navigate({ search: { ...search, offset: offset + limit } })}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
