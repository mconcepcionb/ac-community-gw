import { ErrorState } from "@/components/common/error-state";
import { LoadingState } from "@/components/common/loading-state";
import { PageHeader } from "@/components/common/page-header";
import { PermissionGate } from "@/components/common/permission-gate";
import { Button } from "@/components/ui/button";
import { AnnounceForm } from "./announce-form";
import { ModerationPanel } from "./moderation-panel";
import { useOnline } from "./use-online";

export function AdminOnlinePage() {
  const query = useOnline();

  return (
    <div className="mx-auto max-w-5xl p-8">
      <PageHeader
        title="Online players"
        description="Live list reported by the game server."
        actions={
          <Button variant="outline" size="sm" onClick={() => void query.refetch()}>
            Refresh
          </Button>
        }
      />

      {query.isPending ? <LoadingState label="Loading online list…" /> : null}
      {query.isError ? (
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      ) : null}
      {query.data ? (
        <pre className="overflow-x-auto rounded-md border border-border bg-card p-4 text-xs text-muted-foreground">
          {query.data.output || "No players online."}
        </pre>
      ) : null}

      <section className="mt-8">
        <h2 className="mb-4 text-lg font-semibold">Moderation</h2>
        <ModerationPanel />
      </section>

      <PermissionGate permission="azeroth.admin.announce">
        <section className="mt-8">
          <h2 className="mb-4 text-lg font-semibold">Announce</h2>
          <AnnounceForm />
        </section>
      </PermissionGate>
    </div>
  );
}
