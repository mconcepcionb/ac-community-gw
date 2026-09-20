import { useQuery } from "@tanstack/react-query";

import { azerothAdminAuditListOptions } from "@/api";
import { EmptyState } from "@/components/common/empty-state";
import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";

/** EntityHistory lists the audit entries recorded against one entity. */
export function EntityHistory({
  targetType,
  targetId,
  limit = 20,
}: {
  targetType: string;
  targetId: string;
  limit?: number;
}) {
  const query = useQuery({
    ...azerothAdminAuditListOptions({
      query: { target_type: targetType, target_id: targetId, limit },
    }),
    enabled: targetType !== "" && targetId !== "",
  });
  const entries = query.data?.entries ?? [];

  if (query.isPending) {
    return <LoadingState label="Loading history…" />;
  }
  if (query.isError) {
    return <ErrorState error={query.error} onRetry={() => void query.refetch()} />;
  }
  if (entries.length === 0) {
    return <EmptyState title="No history" />;
  }

  return (
    <ul className="space-y-1 text-sm">
      {entries.map((entry) => (
        <li
          key={entry.request_id || `${entry.occurred_at}-${entry.action}-${entry.actor_id}`}
          className="flex flex-wrap items-center justify-between gap-2 rounded border border-border px-3 py-2"
        >
          <span>
            <span className="font-medium">{entry.action}</span>
            {entry.actor_id ? (
              <span className="ml-2 text-xs text-muted-foreground">{entry.actor_id}</span>
            ) : null}
          </span>
          <span className="text-xs text-muted-foreground">{entry.occurred_at}</span>
        </li>
      ))}
    </ul>
  );
}
