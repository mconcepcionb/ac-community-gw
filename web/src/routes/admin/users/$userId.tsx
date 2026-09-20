import { createFileRoute } from "@tanstack/react-router";

import { AdminUserDetailPage } from "@/features/admin/admin-user-detail-page";

export const Route = createFileRoute("/admin/users/$userId")({
  component: AdminUserRoute,
});

function AdminUserRoute() {
  const { userId } = Route.useParams();
  return <AdminUserDetailPage userId={userId} />;
}
