import { getRouteApi } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useEffect, useState } from "react";

import type { AzerothAccount } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAccounts } from "@/features/accounts/use-accounts";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import { BanAccountDialog } from "./ban-account-dialog";
import { SetGmLevelDialog } from "./set-gmlevel-dialog";
import { UnbanAccountButton } from "./unban-account-button";

const route = getRouteApi("/admin/accounts");

const columns: ColumnDef<AzerothAccount, unknown>[] = [
  { accessorKey: "id", header: "ID" },
  { accessorKey: "username", header: "Username" },
  { accessorKey: "gm_level", header: "GM" },
  {
    accessorKey: "banned",
    header: "Banned",
    cell: ({ row }) => (row.original.banned ? "yes" : "no"),
  },
  {
    accessorKey: "online",
    header: "Online",
    cell: ({ row }) => (row.original.online ? "yes" : "no"),
  },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => {
      const username = row.original.username ?? "";
      return (
        <div className="flex flex-wrap gap-2">
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
  const [filterInput, setFilterInput] = useState(search.filter ?? "");
  const filter = useDebouncedValue(filterInput, 300);

  useEffect(() => {
    if ((search.filter ?? "") === filter) {
      return;
    }
    navigate({ search: { ...search, filter: filter || undefined }, replace: true });
  }, [filter, search, navigate]);

  const query = useAccounts({ filter: search.filter });
  const accounts = query.data?.accounts ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader title="Admin accounts" description="Ban, unban and set GM levels." />

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
    </div>
  );
}
