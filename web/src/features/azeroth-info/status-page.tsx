import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAzerothStatus } from "./use-status";

export function StatusPage() {
  const query = useAzerothStatus();

  if (query.isPending) {
    return <LoadingState label="Loading status…" />;
  }

  if (query.isError) {
    return (
      <div className="mx-auto max-w-5xl p-8">
        <PageHeader title="AzerothCore status" />
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      </div>
    );
  }

  const status = query.data;
  const metrics = [
    { label: "Version", value: status?.version },
    { label: "Connected players", value: status?.connected_players },
    { label: "Characters in world", value: status?.characters_in_world },
    { label: "Connection peak", value: status?.connection_peak },
    { label: "Queue", value: status?.queue },
    { label: "Uptime", value: status?.uptime },
  ];

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="AzerothCore status"
        description="Live snapshot reported by the game server."
      />

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {metrics.map((metric) => (
          <Card key={metric.label}>
            <CardHeader>
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {metric.label}
              </CardTitle>
            </CardHeader>
            <CardContent className="text-2xl font-semibold">{metric.value ?? "—"}</CardContent>
          </Card>
        ))}
      </div>

      {status?.output ? (
        <pre className="mt-6 overflow-x-auto rounded-md border border-border bg-card p-4 text-xs text-muted-foreground">
          {status.output}
        </pre>
      ) : null}
    </div>
  );
}
