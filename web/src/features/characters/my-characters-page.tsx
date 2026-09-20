import type { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

import type { AzerothCharacter } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { EmptyState } from "@/components/common/empty-state";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { SelfMailDialog } from "./self-mail-dialog";
import { useMyCharacters } from "./use-my-characters";

const columns: ColumnDef<AzerothCharacter, unknown>[] = [
  { accessorKey: "name", header: "Name" },
  { accessorKey: "level", header: "Level" },
  { accessorKey: "class_name", header: "Class" },
  { accessorKey: "race_name", header: "Race" },
  { accessorKey: "guild", header: "Guild" },
  {
    accessorKey: "online",
    header: "Online",
    cell: ({ row }) => (row.original.online ? "yes" : "no"),
  },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => (
      <SelfMailDialog
        character={row.original.name ?? ""}
        trigger={
          <Button variant="outline" size="sm">
            Mail
          </Button>
        }
      />
    ),
  },
];

/** MyCharactersPage lists the signed-in user's own characters. */
export function MyCharactersPage() {
  const [filterInput, setFilterInput] = useState("");
  const filter = useDebouncedValue(filterInput, 300);
  const query = useMyCharacters({ filter });
  const characters = query.data?.characters ?? [];
  const linkedAccountMissing = query.isError && characters.length === 0;

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="My characters" description="Characters of your linked game account." />

      <div className="mb-4 max-w-sm">
        <Input
          placeholder="Filter by name"
          aria-label="Filter by name"
          value={filterInput}
          onChange={(event) => setFilterInput(event.target.value)}
        />
      </div>

      {linkedAccountMissing ? (
        <EmptyState
          title="No linked account"
          description="Link your game account to see your characters."
        />
      ) : (
        <DataTable
          columns={columns}
          data={characters}
          loading={query.isPending}
          error={query.isError ? query.error : undefined}
          onRetry={() => void query.refetch()}
          emptyMessage="No characters"
        />
      )}
    </div>
  );
}
