import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { toast } from "sonner";

import {
  apikeysCreateMutation,
  apikeysListOptions,
  apikeysListQueryKey,
  apikeysRevokeMutation,
  apikeysRotateMutation,
  gatewayAdminPermissionsListOptions,
} from "@/api";
import { isApiError } from "@/api/errors";
import { ConfirmDialog } from "@/components/common/confirm-dialog";
import { PageHeader } from "@/components/common/page-header";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

/** AdminApiClientsPage issues and manages scoped API keys. */
export function AdminApiClientsPage() {
  const queryClient = useQueryClient();
  const keys = useQuery(apikeysListOptions());
  const permissions = useQuery(gatewayAdminPermissionsListOptions());
  const [name, setName] = useState("");
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [secret, setSecret] = useState("");

  const create = useMutation(apikeysCreateMutation());
  const rotate = useMutation(apikeysRotateMutation());
  const revoke = useMutation(apikeysRevokeMutation());

  const report = (error: unknown) =>
    toast.error(isApiError(error) ? `${error.message} (${error.code})` : String(error));
  const refresh = () => queryClient.invalidateQueries({ queryKey: apikeysListQueryKey() });

  const toggle = (permission: string) => {
    setSelected((current) => {
      const next = new Set(current);
      if (next.has(permission)) {
        next.delete(permission);
      } else {
        next.add(permission);
      }
      return next;
    });
  };

  const onCreate = async () => {
    try {
      const response = await create.mutateAsync({
        body: { name, permissions: Array.from(selected) },
      });
      setSecret(response.secret ?? "");
      toast.success("API key created");
      setName("");
      setSelected(new Set());
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  const onRotate = async (id: string) => {
    try {
      const response = await rotate.mutateAsync({ path: { id } });
      setSecret(response.secret ?? "");
      toast.success("API key rotated");
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  const onRevoke = async (id: string) => {
    try {
      await revoke.mutateAsync({ path: { id } });
      toast.success("API key revoked");
      await refresh();
    } catch (error) {
      report(error);
    }
  };

  const items = keys.data?.keys ?? [];
  const available = permissions.data?.permissions ?? [];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="API clients"
        description="Scoped service credentials for bots and external sites."
      />

      {secret ? (
        <div className="mb-6 rounded-md border border-border bg-muted/40 p-4 text-sm">
          <p className="font-medium">Copy the secret now; it is shown once.</p>
          <code className="mt-1 block break-all">{secret}</code>
        </div>
      ) : null}

      <Card className="mb-8">
        <CardHeader>
          <CardTitle>New key</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <Input
            className="max-w-sm"
            placeholder="Name"
            aria-label="Key name"
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
          <div className="max-h-64 overflow-y-auto rounded border border-border p-3">
            {available.map((permission) => (
              <label key={permission.name} className="flex items-start gap-2 py-0.5 text-sm">
                <input
                  type="checkbox"
                  className="mt-1"
                  checked={selected.has(permission.name ?? "")}
                  onChange={() => toggle(permission.name ?? "")}
                />
                <span>
                  <span className="font-mono text-xs">{permission.name}</span>
                  {permission.owner ? (
                    <span className="ml-2 text-xs text-muted-foreground">{permission.owner}</span>
                  ) : null}
                </span>
              </label>
            ))}
          </div>
          <Button
            type="button"
            onClick={() => void onCreate()}
            disabled={create.isPending || !name || selected.size === 0}
          >
            Create key
          </Button>
        </CardContent>
      </Card>

      <h2 className="mb-3 text-lg font-semibold">Keys</h2>
      {items.length === 0 ? (
        <p className="text-sm text-muted-foreground">No keys.</p>
      ) : (
        <ul className="space-y-2">
          {items.map((key) => (
            <li key={key.id} className="rounded border border-border p-3 text-sm">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span>
                  <span className="font-medium">{key.name}</span>{" "}
                  <span className="font-mono text-xs text-muted-foreground">
                    ak_{key.key_prefix}…
                  </span>
                </span>
                <div className="flex gap-2">
                  <ConfirmDialog
                    trigger={
                      <Button variant="outline" size="sm">
                        Rotate
                      </Button>
                    }
                    title="Rotate this key?"
                    description="The current secret stops working immediately."
                    confirmLabel="Rotate"
                    onConfirm={() => {
                      void onRotate(key.id ?? "");
                      return true;
                    }}
                  />
                  <ConfirmDialog
                    trigger={
                      <Button variant="outline" size="sm">
                        Revoke
                      </Button>
                    }
                    title="Revoke this key?"
                    description="The key stops working immediately."
                    confirmLabel="Revoke"
                    destructive
                    onConfirm={() => {
                      void onRevoke(key.id ?? "");
                      return true;
                    }}
                  />
                </div>
              </div>
              <p className="mt-1 text-xs text-muted-foreground">
                {(key.permissions ?? []).length} scopes
                {key.last_used_at ? ` · last used ${key.last_used_at}` : " · never used"}
              </p>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
