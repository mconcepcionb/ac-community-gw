import type { ColumnDef } from "@tanstack/react-table";

import type { AzerothCharacter } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { MailForm } from "./mail-form";
import { useUserCharacters } from "./use-user-characters";

const columns: ColumnDef<AzerothCharacter, unknown>[] = [
  { accessorKey: "name", header: "Name" },
  { accessorKey: "level", header: "Level" },
  { accessorKey: "class_name", header: "Class" },
  { accessorKey: "race_name", header: "Race" },
  {
    accessorKey: "online",
    header: "Online",
    cell: ({ row }) => (row.original.online ? "yes" : "no"),
  },
];

export function UserCharactersPage({ userId }: { userId: string }) {
  const query = useUserCharacters(userId);
  const characters = query.data?.characters ?? [];

  if (userId === "") {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="User characters" description="—" />
        <p className="text-sm text-muted-foreground">Invalid user id.</p>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="User characters" description={userId} />

      <DataTable
        columns={columns}
        data={characters}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No characters"
      />

      <section className="mt-8">
        <h2 className="mb-4 text-lg font-semibold">Send mail</h2>
        <PermissionGate
          permission="azeroth.mail.send"
          fallback={
            <p className="text-sm text-muted-foreground">
              You lack the azeroth.mail.send permission.
            </p>
          }
        >
          <MailForm />
        </PermissionGate>
      </section>
    </div>
  );
}
