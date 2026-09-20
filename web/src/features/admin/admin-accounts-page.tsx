import { getRouteApi, Link } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useEffect, useState } from "react";

import type { AzerothAccount } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { Pagination } from "@/components/common/pagination";
import { PermissionGate } from "@/components/common/permission-gate";
import { RowActions } from "@/components/common/row-actions";
import { StatusBadge } from "@/components/common/status-badge";
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

const route = getRouteApi("/admin/accounts/");

const gmLevelLabels = ["Player", "Moderator", "Game Master", "Administrator", "Console"];

const columns: ColumnDef<AzerothAccount, unknown>[] = [
  { accessorKey: "id", header: "ID" },
  {
    accessorKey: "username",
    header: "Username",
    cell: ({ row }) => (
      <Link
        to="/admin/accounts/$username"
        params={{ username: row.original.username ?? "" }}
        className="text-blue-400 underline"
      >
        {row.original.username}
      </Link>
    ),
  },
  { accessorKey: "email", header: "Email" },
  {
    id: "owner",
    header: "Owner",
    cell: ({ row }) => {
      const owner = row.original.claimed_by;
      if (!owner) {
        return <StatusBadge tone="neutral">Unclaimed</StatusBadge>;
      }
      return owner.user_id ? (
        <Link
          to="/admin/users/$userId"
          params={{ userId: owner.user_id }}
          className="text-blue-400 underline"
        >
          {owner.display_name || owner.user_id}
        </Link>
      ) : (
        <span>{owner.display_name}</span>
      );
    },
  },
  {
    accessorKey: "gm_level",
    header: "GM",
    cell: ({ row }) => {
      const level = row.original.gm_level ?? 0;
      const label = `${level} — ${gmLevelLabels[level] ?? "Unknown"}`;
      return (
        <PermissionGate permission="azeroth.admin.accounts.gmlevel" fallback={<span>{label}</span>}>
          <SetGmLevelDialog
            username={row.original.username ?? ""}
            currentLevel={level}
            trigger={
              <Button variant="outline" size="sm">
                {label}
              </Button>
            }
          />
        </PermissionGate>
      );
    },
  },
  {
    accessorKey: "online",
    header: "Online",
    cell: ({ row }) => (
      <StatusBadge tone={row.original.online ? "positive" : "neutral"}>
        {row.original.online ? "Online" : "Offline"}
      </StatusBadge>
    ),
  },
  {
    accessorKey: "banned",
    header: "Banned",
    cell: ({ row }) => (
      <div className="flex flex-col gap-1">
        <StatusBadge tone={row.original.banned ? "negative" : "positive"}>
          {row.original.banned ? "Banned" : "Active"}
        </StatusBadge>
        {row.original.banned && row.original.ban_reason ? (
          <span className="text-xs text-muted-foreground">{row.original.ban_reason}</span>
        ) : null}
      </div>
    ),
  },
  { accessorKey: "last_ip", header: "Last IP" },
  { accessorKey: "last_login", header: "Last login" },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => {
      const username = row.original.username ?? "";
      return (
        <RowActions>
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
            {row.original.banned ? (
              <UnbanAccountButton
                username={username}
                trigger={
                  <Button variant="outline" size="sm">
                    Unban
                  </Button>
                }
              />
            ) : (
              <BanAccountDialog
                username={username}
                trigger={
                  <Button variant="outline" size="sm">
                    Ban
                  </Button>
                }
              />
            )}
          </PermissionGate>
        </RowActions>
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

      <Pagination
        limit={limit}
        offset={offset}
        count={accounts.length}
        itemLabel="accounts"
        onOffsetChange={(next) =>
          navigate({ search: { ...search, offset: next === 0 ? undefined : next } })
        }
      />
    </div>
  );
}
