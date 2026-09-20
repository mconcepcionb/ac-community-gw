import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { ReportPage } from "@/features/reports/report-page";

export const Route = createFileRoute("/_portal/report")({
  component: ReportRoute,
});

function ReportRoute() {
  return (
    <RequireAuth>
      <ReportPage />
    </RequireAuth>
  );
}
