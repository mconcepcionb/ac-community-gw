import { createFileRoute } from "@tanstack/react-router";
import { AdminOnlinePage } from "@/features/admin/admin-online-page";
import { RequireAuth } from "@/features/auth/require-auth";

export const Route = createFileRoute("/admin/online")({
  component: AdminOnlineRoute,
});

function AdminOnlineRoute() {
  return (
    <RequireAuth>
      <AdminOnlinePage />
    </RequireAuth>
  );
}
