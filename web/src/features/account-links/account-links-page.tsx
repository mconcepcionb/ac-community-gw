import { Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";

import type { AzerothAccountLink } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { CreateLinkDialog } from "./create-link-dialog";
import { DeleteLinkButton } from "./delete-link-button";
import { useAccountLinks } from "./use-account-links";

const columns: ColumnDef<AzerothAccountLink, unknown>[] = [
  {
    accessorKey: "user_id",
    header: "User ID",
    cell: ({ row }) => (
      <Link
        to="/account-links/$userId"
        params={{ userId: row.original.user_id ?? "" }}
        className="text-blue-400 underline"
      >
        {row.original.user_id}
      </Link>
    ),
  },
  { accessorKey: "account_username", header: "Account" },
  { accessorKey: "account_id", header: "Account ID" },
  { accessorKey: "linked_at", header: "Linked at" },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => (
      <PermissionGate permission="azeroth.account.link">
        <DeleteLinkButton userId={row.original.user_id ?? ""} />
      </PermissionGate>
    ),
  },
];

export function AccountLinksPage() {
  const query = useAccountLinks();
  const links = query.data?.links ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Account links"
        description="Community users linked to AzerothCore accounts."
        actions={
          <PermissionGate permission="azeroth.account.link">
            <CreateLinkDialog trigger={<Button>Create link</Button>} />
          </PermissionGate>
        }
      />

      <DataTable
        columns={columns}
        data={links}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No links"
      />
    </div>
  );
}
