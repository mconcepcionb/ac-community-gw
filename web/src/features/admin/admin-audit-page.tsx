import { useQuery } from "@tanstack/react-query";
import { getRouteApi } from "@tanstack/react-router";
import type { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";

import type { AdminAuditEntry } from "@/api";
import { gatewayAdminAuditListOptions } from "@/api";
import { DataTable } from "@/components/common/data-table";
import { PageHeader } from "@/components/common/page-header";
import { Input } from "@/components/ui/input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";

const route = getRouteApi("/admin/audit");

const columns: ColumnDef<AdminAuditEntry, unknown>[] = [
  { accessorKey: "occurred_at", header: "When" },
  { accessorKey: "actor_id", header: "Actor" },
  { accessorKey: "action", header: "Action" },
  { accessorKey: "target_type", header: "Target type" },
  { accessorKey: "target_id", header: "Target" },
  { accessorKey: "result", header: "Result" },
];

/** AdminAuditPage reads the audit log with simple filters. */
export function AdminAuditPage() {
  const search = route.useSearch();
  const [actor, setActor] = useState(search.actor ?? "");
  const [target, setTarget] = useState(search.target ?? "");
  const [action, setAction] = useState(search.action ?? "");
  const actorQuery = useDebouncedValue(actor, 300);
  const targetQuery = useDebouncedValue(target, 300);
  const actionQuery = useDebouncedValue(action, 300);

  const query = useQuery(
    gatewayAdminAuditListOptions({
      query: {
        actor: actorQuery || undefined,
        target: targetQuery || undefined,
        action: actionQuery || undefined,
        limit: 100,
      },
    }),
  );
  const entries = query.data?.entries ?? [];

  return (
    <div className="mx-auto max-w-6xl p-8">
      <PageHeader title="Audit log" description="Who did what, newest first." />

      <div className="mb-4 flex flex-wrap gap-3">
        <Input
          className="max-w-xs"
          placeholder="Actor user id"
          aria-label="Actor"
          value={actor}
          onChange={(event) => setActor(event.target.value)}
        />
        <Input
          className="max-w-xs"
          placeholder="Target"
          aria-label="Target"
          value={target}
          onChange={(event) => setTarget(event.target.value)}
        />
        <Input
          className="max-w-xs"
          placeholder="Action"
          aria-label="Action"
          value={action}
          onChange={(event) => setAction(event.target.value)}
        />
      </div>

      <DataTable
        columns={columns}
        data={entries}
        loading={query.isPending}
        error={query.isError ? query.error : undefined}
        onRetry={() => void query.refetch()}
        emptyMessage="No entries"
      />
    </div>
  );
}
