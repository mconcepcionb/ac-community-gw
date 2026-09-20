import { getRouteApi, Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useCallback, useEffect, useState } from "react";

import { type AzerothCharacter, azerothAccountsList } from "@/api";
import { Autocomplete, type AutocompleteOption } from "@/components/common/autocomplete";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { Pagination } from "@/components/common/pagination";
import { StatusBadge } from "@/components/common/status-badge";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { useCharacters } from "./use-characters";

const route = getRouteApi("/admin/characters/");

const columns: ColumnDef<AzerothCharacter, unknown>[] = [
  {
    accessorKey: "name",
    header: "Name",
    cell: ({ row }) => (
      <Link
        to="/admin/characters/$name"
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
    cell: ({ row }) => (
      <StatusBadge tone={row.original.online ? "positive" : "neutral"}>
        {row.original.online ? "Online" : "Offline"}
      </StatusBadge>
    ),
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
  const total = query.data?.total;

  const loadAccounts = useCallback(async (value: string): Promise<AutocompleteOption[]> => {
    try {
      const response = await azerothAccountsList({ query: { filter: value, limit: 10 } });
      return (response.data?.accounts ?? [])
        .map((item) => item.username ?? "")
        .filter((username) => username !== "")
        .map((username) => ({ value: username, label: username }));
    } catch {
      return [];
    }
  }, []);

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Characters"
        description="Every AzerothCore character, filterable by account and name."
      />

      <div className="mb-4 flex flex-wrap gap-3">
        <Autocomplete
          className="max-w-xs"
          value={accountInput}
          onValueChange={setAccountInput}
          loadOptions={loadAccounts}
          ariaLabel="Account username"
          placeholder="Account username"
        />
        <Input
          className="max-w-xs"
          placeholder="Filter by name"
          aria-label="Filter by name"
          value={filterInput}
          onChange={(event) => setFilterInput(event.target.value)}
        />
      </div>

      <DataTable
        columns={columns}
        data={characters}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No characters"
      />

      <Pagination
        limit={limit}
        offset={offset}
        count={characters.length}
        total={total}
        itemLabel="characters"
        onOffsetChange={(next) =>
          navigate({ search: { ...search, offset: next === 0 ? undefined : next } })
        }
      />
    </div>
  );
}
