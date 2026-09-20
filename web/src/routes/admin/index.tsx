import { createFileRoute } from "@tanstack/react-router";

import { PageHeader } from "@/components/common/page-header";

export const Route = createFileRoute("/admin/")({
  component: ConsoleOverviewPlaceholder,
});

function ConsoleOverviewPlaceholder() {
  return (
    <div className="mx-auto max-w-6xl p-8">
      <PageHeader title="Operations console" description="Staff tools for the community gateway." />
      <p className="text-sm text-muted-foreground">
        The operations overview is built in a later change.
      </p>
    </div>
  );
}
