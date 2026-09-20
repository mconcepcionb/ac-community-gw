import { getRouteApi } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useEffect, useState } from "react";

import type { AzerothAccount } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { CreateAccountDialog } from "@/features/accounts/create-account-dialog";
import { SetEmailDialog } from "@/features/accounts/set-email-dialog";
import { SetPasswordDialog } from "@/features/accounts/set-password-dialog";
import { useAccounts } from "@/features/accounts/use-accounts";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { BanAccountDialog } from "./ban-account-dialog";
import { SetGmLevelDialog } from "./set-gmlevel-dialog";
import { UnbanAccountButton } from "./unban-account-button";

const route = getRouteApi("/admin/accounts");

const columns: ColumnDef<AzerothAccount, unknown>[] = [
  { accessorKey: "id", header: "ID" },
  { accessorKey: "username", header: "Username" },
  { accessorKey: "email", header: "Email" },
  { accessorKey: "gm_level", header: "GM" },
  {
    accessorKey: "online",
    header: "Online",
    cell: ({ row }) => (row.original.online ? "yes" : "no"),
  },
  {
    accessorKey: "banned",
    header: "Banned",
    cell: ({ row }) => (row.original.banned ? "yes" : "no"),
  },
  { accessorKey: "last_login", header: "Last login" },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => {
      const username = row.original.username ?? "";
      return (
        <div className="flex flex-wrap gap-2">
          <PermissionGate permission="azeroth.account.manage">
            <SetPasswordDialog
              username={username}
              trigger={
                <Button variant="outline" size="sm">
                  Password
                </Button>
              }
            />
            <SetEmailDialog
              username={username}
              currentEmail={row.original.email}
              trigger={
                <Button variant="outline" size="sm">
                  Email
                </Button>
              }
            />
          </PermissionGate>
          <PermissionGate permission="azeroth.admin.accounts.ban">
            <BanAccountDialog
              username={username}
              trigger={
                <Button variant="outline" size="sm">
                  Ban
                </Button>
              }
            />
            <UnbanAccountButton
              username={username}
              trigger={
                <Button variant="outline" size="sm">
                  Unban
                </Button>
              }
            />
          </PermissionGate>
          <PermissionGate permission="azeroth.admin.accounts.gmlevel">
            <SetGmLevelDialog
              username={username}
              trigger={
                <Button variant="outline" size="sm">
                  GM level
                </Button>
              }
            />
          </PermissionGate>
        </div>
      );
    },
  },
];

export function AdminAccountsPage() {
  const search = route.useSearch();
  const navigate = route.useNavigate();
  const limit = search.limit ?? 50;
  const offset = search.offset ?? 0;

  const [filterInput, setFilterInput] = useState(search.filter ?? "");
  const filter = useDebouncedValue(filterInput, 300);

  useEffect(() => {
    if ((search.filter ?? "") === filter) {
      return;
    }
    navigate({
      search: { ...search, filter: filter || undefined, offset: undefined },
      replace: true,
    });
  }, [filter, search, navigate]);

  const query = useAccounts({ filter: search.filter, limit, offset });
  const accounts = query.data?.accounts ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Accounts"
        description="Login accounts, credentials and moderation."
        actions={
          <PermissionGate permission="azeroth.account.manage">
            <CreateAccountDialog trigger={<Button>Create account</Button>} />
          </PermissionGate>
        }
      />

      <div className="mb-4 max-w-sm">
        <Input
          placeholder="Search accounts…"
          aria-label="Search accounts"
          value={filterInput}
          onChange={(event) => setFilterInput(event.target.value)}
        />
      </div>

      <DataTable
        columns={columns}
        data={accounts}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No accounts"
      />

      <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
        <span>
          Showing {accounts.length === 0 ? 0 : offset + 1}–{offset + accounts.length}
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
            disabled={accounts.length < limit}
            onClick={() => navigate({ search: { ...search, offset: offset + limit } })}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
