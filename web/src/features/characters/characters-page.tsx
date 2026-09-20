import { getRouteApi, Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useEffect, useState } from "react";

import type { AzerothCharacter } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { EmptyState } from "@/components/common/empty-state";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useCharacters } from "./use-characters";

const route = getRouteApi("/_portal/characters/");

const columns: ColumnDef<AzerothCharacter, unknown>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => (
      <Link
        to="/characters/$name"
        params={{ name: row.original.name ?? "" }}
        className="text-blue-400 underline"
      >
        {row.original.name}
      </Link>
    ),
  },
  { accessorKey: "level", header: "Level" },
  { accessorKey: "class_name", header: "Class" },
  { accessorKey: "race_name", header: "Race" },
  { accessorKey: "guild", header: "Guild" },
  {
    accessorKey: "online",
    header: "Online",
    cell: ({ row }) => (row.original.online ? "yes" : "no"),
  },
  { accessorKey: "logout_time", header: "Last logout" },
];

export function CharactersPage() {
  const search = route.useSearch();
  const navigate = route.useNavigate();
  const limit = search.limit ?? 100;
  const offset = search.offset ?? 0;

  const [accountInput, setAccountInput] = useState(search.account ?? "");
  const [filterInput, setFilterInput] = useState(search.filter ?? "");
  const account = useDebouncedValue(accountInput, 300);
  const filter = useDebouncedValue(filterInput, 300);

  useEffect(() => {
    if ((search.account ?? "") === account && (search.filter ?? "") === filter) {
      return;
    }
    navigate({
      search: {
        ...search,
        account: account || undefined,
        filter: filter || undefined,
        offset: undefined,
      },
      replace: true,
    });
  }, [account, filter, search, navigate]);

  const query = useCharacters({ account: search.account, filter: search.filter, limit, offset });
  const characters = query.data?.characters ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="Characters" description="Characters of an AzerothCore account." />

      <div className="mb-4 flex flex-wrap gap-3">
        <Input
          className="max-w-xs"
          placeholder="Account username"
          aria-label="Account username"
          value={accountInput}
          onChange={(event) => setAccountInput(event.target.value)}
        />
        <Input
          className="max-w-xs"
          placeholder="Filter by name"
          aria-label="Filter by name"
          value={filterInput}
          onChange={(event) => setFilterInput(event.target.value)}
        />
      </div>

      {search.account ? (
        <>
          <DataTable
            columns={columns}
            data={characters}
            loading={query.isPending}
            error={query.isError ? query.error : undefined}
            onRetry={() => void query.refetch()}
            emptyMessage="No characters"
          />
          <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
            <span>
              Showing {characters.length === 0 ? 0 : offset + 1}–{offset + characters.length}
            </span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={offset === 0}
                onClick={() =>
                  navigate({ search: { ...search, offset: Math.max(0, offset - limit) } })
                }
              >
                Previous
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={characters.length < limit}
                onClick={() => navigate({ search: { ...search, offset: offset + limit } })}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      ) : (
        <EmptyState
          title="Choose an account"
          description="Enter an AzerothCore account username to list its characters."
        />
      )}
    </div>
  );
}
