import type { ColumnDef } from "@tanstack/react-table";
import { useMemo, useState } from "react";

import type { AzerothCharacter } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { EmptyState } from "@/components/common/empty-state";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { SelfMailDialog } from "./self-mail-dialog";
import { useCharacterVisibility, useSetCharacterVisibility } from "./use-character-visibility";
import { useMyCharacters } from "./use-my-characters";

function VisibilityToggle({ name, publicFlag }: { name: string; publicFlag: boolean }) {
  const mutation = useSetCharacterVisibility();
  return (
    <Button
      variant="outline"
      size="sm"
      disabled={mutation.isPending}
      onClick={() => mutation.mutate({ path: { name }, body: { public: !publicFlag } })}
    >
      {publicFlag ? "Public" : "Private"}
    </Button>
  );
}

/** MyCharactersPage lists the signed-in user's own characters. */
export function MyCharactersPage() {
  const [filterInput, setFilterInput] = useState("");
  const filter = useDebouncedValue(filterInput, 300);
  const query = useMyCharacters({ filter });
  const visibility = useCharacterVisibility();
  const characters = query.data?.characters ?? [];
  const linkedAccountMissing = query.isError && characters.length === 0;

  const flags = useMemo(() => {
    const map: Record<string, boolean> = {};
    for (const item of visibility.data?.items ?? []) {
      if (item.name) {
        map[item.name] = item.public ?? false;
      }
    }
    return map;
  }, [visibility.data]);

  const columns = useMemo<ColumnDef<AzerothCharacter, unknown>[]>(
    () => [
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
        cell: ({ row }) => {
          const name = row.original.name ?? "";
          return (
            <div className="flex flex-wrap gap-2">
              {visibility.isError ? null : (
                <VisibilityToggle name={name} publicFlag={flags[name] ?? false} />
              )}
              <SelfMailDialog
                character={name}
                trigger={
                  <Button variant="outline" size="sm">
                    Mail
                  </Button>
                }
              />
            </div>
          );
        },
      },
    ],
    [flags, visibility.isError],
  );

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="My characters"
        description="Characters of your linked game account. Mark a character Public to show it on leaderboards."
      />

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
