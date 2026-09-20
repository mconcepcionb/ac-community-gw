import type { ColumnDef } from "@tanstack/react-table";

import type { AzerothOnlinePlayer } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { RowActions } from "@/components/common/row-actions";
import { Button } from "@/components/ui/button";
import { AnnounceForm } from "./announce-form";
import {
  BanCharacterDialog,
  KickDialog,
  ModerationPanel,
  MuteDialog,
  UnbanCharacterButton,
  UnmuteButton,
} from "./moderation-panel";
import { useOnline } from "./use-online";

const columns: ColumnDef<AzerothOnlinePlayer, unknown>[] = [
  { accessorKey: "name", header: "Name" },
  { accessorKey: "level", header: "Level" },
  { accessorKey: "class_name", header: "Class" },
  { accessorKey: "race_name", header: "Race" },
  { accessorKey: "guild", header: "Guild" },
  { accessorKey: "account_id", header: "Account" },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => {
      const name = row.original.name ?? "";
      return (
        <RowActions>
          <PermissionGate permission="azeroth.admin.players.kick">
            <KickDialog name={name} />
          </PermissionGate>
          <PermissionGate permission="azeroth.admin.players.mute">
            <MuteDialog name={name} />
            <UnmuteButton name={name} />
          </PermissionGate>
          <PermissionGate permission="azeroth.admin.characters.ban">
            <BanCharacterDialog name={name} />
            <UnbanCharacterButton name={name} />
          </PermissionGate>
        </RowActions>
      );
    },
  },
];

export function AdminOnlinePage() {
  const query = useOnline();
  const players = query.data?.players ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Online players"
        description="Live list read from the character database."
        actions={
          <Button variant="outline" size="sm" onClick={() => void query.refetch()}>
            Refresh
          </Button>
        }
      />

      <DataTable
        columns={columns}
        data={players}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No players online"
      />
      {query.data ? (
        <p className="mt-2 text-xs text-muted-foreground">
          Last updated {new Date(query.dataUpdatedAt).toLocaleTimeString()}
        </p>
      ) : null}

      <section className="mt-8">
        <h2 className="mb-4 text-lg font-semibold">Moderation</h2>
        <ModerationPanel />
      </section>

      <PermissionGate permission="azeroth.admin.announce">
        <section className="mt-8">
          <h2 className="mb-4 text-lg font-semibold">Announce</h2>
          <AnnounceForm />
        </section>
      </PermissionGate>
    </div>
  );
}
