import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import type { AdminDiscordRoleMapping, AdminPermissionDefinition } from "@/api";
import {
  identityAdminDiscordMappingsDeleteMutation,
  identityAdminDiscordMappingsUpsertMutation,
  identityAdminRolesListOptions,
  identityAdminRolesListQueryKey,
  identityAdminRolesReplacePermissionsMutation,
} from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const report = (error: unknown) =>
  toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));

/** AdminRolesPage manages role grants and Discord role mappings. */
export function AdminRolesPage() {
  const queryClient = useQueryClient();
  const query = useQuery(identityAdminRolesListOptions());
  const replace = useMutation(identityAdminRolesReplacePermissionsMutation());
  const upsertMapping = useMutation(identityAdminDiscordMappingsUpsertMutation());
  const deleteMapping = useMutation(identityAdminDiscordMappingsDeleteMutation());

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: identityAdminRolesListQueryKey() });

  const roles = query.data?.roles ?? [];
  const permissions = query.data?.permissions ?? [];
  const grants = query.data?.grants ?? [];
  const mappings = query.data?.mappings ?? [];

  const grouped = useMemo(() => {
    const groups = new Map<string, AdminPermissionDefinition[]>();
    for (const permission of permissions) {
      const owner = permission.owner || "other";
      const bucket = groups.get(owner) ?? [];
      bucket.push(permission);
      groups.set(owner, bucket);
    }
    return [...groups.entries()].sort(([a], [b]) => a.localeCompare(b));
  }, [permissions]);

  const grantsByRole = useMemo(() => {
    const map: Record<string, Set<string>> = {};
    for (const grant of grants) {
      if (!grant.role || !grant.permission) {
        continue;
      }
      map[grant.role] = map[grant.role] ?? new Set();
      map[grant.role].add(grant.permission);
    }
    return map;
  }, [grants]);

  if (query.isPending) {
    return <LoadingState label="Loading roles…" />;
  }
  if (query.isError) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="Roles" />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const onSave = async (role: string, selected: Set<string>) => {
    try {
      await replace.mutateAsync({
        path: { role },
        body: { permissions: Array.from(selected) },
      });
      toast.success("Permissions saved");
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  const onAddMapping = async (discordRoleID: string, role: string) => {
    try {
      await upsertMapping.mutateAsync({ path: { discord_role_id: discordRoleID }, body: { role } });
      toast.success("Mapping saved");
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  const onDeleteMapping = async (discordRoleID: string) => {
    try {
      await deleteMapping.mutateAsync({ path: { discord_role_id: discordRoleID } });
      toast.success("Mapping removed");
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Roles"
        description="Permission grants and Discord role mappings. Changes apply within the refresh interval."
      />

      <Tabs defaultValue="matrix">
        <TabsList>
          <TabsTrigger value="matrix">Role permissions</TabsTrigger>
          <TabsTrigger value="mappings">Discord mappings</TabsTrigger>
        </TabsList>

        <TabsContent value="matrix">
          {roles.length === 0 ? (
            <EmptyState title="No roles" description="Create a role by mapping a Discord role." />
          ) : (
            <div className="space-y-4">
              {roles.map((role) => (
                <RoleMatrix
                  key={role}
                  role={role}
                  grouped={grouped}
                  current={grantsByRole[role] ?? new Set()}
                  pending={replace.isPending}
                  onSave={onSave}
                />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="mappings">
          <MappingsTab
            mappings={mappings}
            roles={roles}
            pending={upsertMapping.isPending}
            onAdd={onAddMapping}
            onDelete={onDeleteMapping}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}

function RoleMatrix({
  role,
  grouped,
  current,
  pending,
  onSave,
}: {
  role: string;
  grouped: [string, AdminPermissionDefinition[]][];
  current: Set<string>;
  pending: boolean;
  onSave: (role: string, selected: Set<string>) => void;
}) {
  const [selected, setSelected] = useState<Set<string>>(() => new Set(current));

  useEffect(() => {
    setSelected(new Set(current));
  }, [current]);

  const toggle = (permission: string) => {
    setSelected((value) => {
      const next = new Set(value);
      if (next.has(permission)) {
        next.delete(permission);
      } else {
        next.add(permission);
      }
      return next;
    });
  };

  const dirty = selected.size !== current.size || [...selected].some((item) => !current.has(item));

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between gap-2">
        <CardTitle className="text-base">{role}</CardTitle>
        <Button size="sm" disabled={!dirty || pending} onClick={() => onSave(role, selected)}>
          Save
        </Button>
      </CardHeader>
      <CardContent className="space-y-4">
        {grouped.map(([owner, items]) => (
          <div key={owner}>
            <p className="mb-1 text-xs font-medium tracking-wide text-muted-foreground uppercase">
              {owner}
            </p>
            <div className="grid gap-1 sm:grid-cols-2">
              {items.map((permission) => (
                <label
                  key={permission.name}
                  className="flex items-start gap-2 text-sm"
                  title={permission.description}
                >
                  <input
                    type="checkbox"
                    className="mt-1"
                    checked={selected.has(permission.name ?? "")}
                    onChange={() => toggle(permission.name ?? "")}
                  />
                  <span className="font-mono text-xs">{permission.name}</span>
                </label>
              ))}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}

function MappingsTab({
  mappings,
  roles,
  pending,
  onAdd,
  onDelete,
}: {
  mappings: AdminDiscordRoleMapping[];
  roles: string[];
  pending: boolean;
  onAdd: (discordRoleID: string, role: string) => void;
  onDelete: (discordRoleID: string) => void;
}) {
  const [discordRoleID, setDiscordRoleID] = useState("");
  const [role, setRole] = useState("");

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">New mapping</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-wrap items-center gap-2">
          <Input
            className="max-w-[16rem]"
            placeholder="Discord role id"
            aria-label="Discord role id"
            value={discordRoleID}
            onChange={(event) => setDiscordRoleID(event.target.value)}
          />
          <Input
            className="max-w-[14rem]"
            placeholder="Internal role"
            aria-label="Internal role"
            value={role}
            onChange={(event) => setRole(event.target.value)}
            list="role-options"
          />
          <datalist id="role-options">
            {roles.map((item) => (
              <option key={item} value={item} />
            ))}
          </datalist>
          <Button
            type="button"
            disabled={pending || discordRoleID === "" || role === ""}
            onClick={() => {
              onAdd(discordRoleID, role);
              setDiscordRoleID("");
              setRole("");
            }}
          >
            Add mapping
          </Button>
        </CardContent>
      </Card>

      {mappings.length === 0 ? (
        <EmptyState title="No mappings" />
      ) : (
        <ul className="space-y-1 text-sm">
          {mappings.map((mapping) => (
            <li
              key={mapping.discord_role_id}
              className="flex items-center justify-between rounded border border-border px-3 py-2"
            >
              <span>
                {mapping.discord_role_id} → <span className="font-medium">{mapping.role}</span>
              </span>
              <ConfirmDialog
                trigger={
                  <Button variant="outline" size="sm">
                    Remove
                  </Button>
                }
                title="Remove this mapping?"
                description="The Discord role stops granting the internal role."
                confirmLabel="Remove"
                destructive
                onConfirm={() => {
                  onDelete(mapping.discord_role_id ?? "");
                  return true;
                }}
              />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
