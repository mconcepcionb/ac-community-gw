import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { toast } from "sonner";

import {
  identityAdminDiscordMappingsDeleteMutation,
  identityAdminDiscordMappingsUpsertMutation,
  identityAdminRolesGrantMutation,
  identityAdminRolesListOptions,
  identityAdminRolesListQueryKey,
  identityAdminRolesRevokeMutation,
} from "@/api";
import { isApiError } from "@/api/errors";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

/** AdminRolesPage manages role grants and Discord role mappings. */
export function AdminRolesPage() {
  const queryClient = useQueryClient();
  const query = useQuery(identityAdminRolesListOptions());
  const [newRole, setNewRole] = useState("");
  const [newPermission, setNewPermission] = useState("");
  const [mappingDiscordID, setMappingDiscordID] = useState("");
  const [mappingRole, setMappingRole] = useState("");

  const grant = useMutation(identityAdminRolesGrantMutation());
  const revoke = useMutation(identityAdminRolesRevokeMutation());
  const upsertMapping = useMutation(identityAdminDiscordMappingsUpsertMutation());
  const deleteMapping = useMutation(identityAdminDiscordMappingsDeleteMutation());

  const report = (error: unknown) =>
    toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: identityAdminRolesListQueryKey() });

  const grantsByRole = useMemo(() => {
    const map: Record<string, string[]> = {};
    for (const item of query.data?.grants ?? []) {
      if (!item.role || !item.permission) {
        continue;
      }
      if (!map[item.role]) {
        map[item.role] = [];
      }
      map[item.role].push(item.permission);
    }
    return map;
  }, [query.data]);

  const mappings = query.data?.mappings ?? [];

  const onGrant = async () => {
    try {
      await grant.mutateAsync({ path: { role: newRole }, body: { permission: newPermission } });
      toast.success("Permission granted");
      setNewPermission("");
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  const onRevoke = async (role: string, permission: string) => {
    try {
      await revoke.mutateAsync({ path: { role, permission } });
      toast.success("Permission revoked");
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  const onAddMapping = async () => {
    try {
      await upsertMapping.mutateAsync({
        path: { discord_role_id: mappingDiscordID },
        body: { role: mappingRole },
      });
      toast.success("Mapping saved");
      setMappingDiscordID("");
      setMappingRole("");
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

      <section className="mb-8">
        <h2 className="mb-3 text-lg font-semibold">Role permissions</h2>
        {query.isPending ? <p className="text-sm text-muted-foreground">Loading…</p> : null}
        {query.isError ? <p className="text-sm text-muted-foreground">Unavailable.</p> : null}
        {Object.keys(grantsByRole).length === 0 && query.data ? (
          <p className="text-sm text-muted-foreground">No grants.</p>
        ) : null}
        <div className="space-y-2">
          {Object.entries(grantsByRole).map(([role, permissions]) => (
            <Card key={role}>
              <CardHeader>
                <CardTitle className="text-base">{role}</CardTitle>
              </CardHeader>
              <CardContent className="flex flex-wrap gap-2">
                {permissions.map((permission) => (
                  <span
                    key={permission}
                    className="inline-flex items-center gap-1 rounded border border-border px-2 py-1 text-xs"
                  >
                    {permission}
                    <button
                      type="button"
                      className="text-muted-foreground hover:text-foreground"
                      aria-label={`Revoke ${permission}`}
                      onClick={() => void onRevoke(role, permission)}
                    >
                      ×
                    </button>
                  </span>
                ))}
              </CardContent>
            </Card>
          ))}
        </div>

        <div className="mt-3 flex flex-wrap items-center gap-2">
          <Input
            className="max-w-[12rem]"
            placeholder="Role"
            aria-label="Role"
            value={newRole}
            onChange={(event) => setNewRole(event.target.value)}
          />
          <Input
            className="max-w-[16rem]"
            placeholder="Permission"
            aria-label="Permission"
            value={newPermission}
            onChange={(event) => setNewPermission(event.target.value)}
          />
          <Button
            type="button"
            onClick={() => void onGrant()}
            disabled={grant.isPending || !newRole || !newPermission}
          >
            Grant
          </Button>
        </div>
      </section>

      <section>
        <h2 className="mb-3 text-lg font-semibold">Discord role mappings</h2>
        {mappings.length === 0 ? (
          <p className="text-sm text-muted-foreground">No mappings.</p>
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
                <button
                  type="button"
                  className="text-muted-foreground hover:text-foreground"
                  aria-label={`Remove ${mapping.discord_role_id}`}
                  onClick={() => void onDeleteMapping(mapping.discord_role_id ?? "")}
                >
                  ×
                </button>
              </li>
            ))}
          </ul>
        )}

        <div className="mt-3 flex flex-wrap items-center gap-2">
          <Input
            className="max-w-[14rem]"
            placeholder="Discord role id"
            aria-label="Discord role id"
            value={mappingDiscordID}
            onChange={(event) => setMappingDiscordID(event.target.value)}
          />
          <Input
            className="max-w-[12rem]"
            placeholder="Internal role"
            aria-label="Internal role"
            value={mappingRole}
            onChange={(event) => setMappingRole(event.target.value)}
          />
          <Button
            type="button"
            onClick={() => void onAddMapping()}
            disabled={upsertMapping.isPending || !mappingDiscordID || !mappingRole}
          >
            Add mapping
          </Button>
        </div>
      </section>
    </div>
  );
}
