import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { StatusPage } from "@/features/azeroth-info/status-page";

export const Route = createFileRoute("/azeroth/status")({
  component: StatusRoute,
});

function StatusRoute() {
  return (
    <RequireAuth>
      <StatusPage />
    </RequireAuth>
  );
}
