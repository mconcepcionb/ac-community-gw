import { createFileRoute } from "@tanstack/react-router";

import { AdminAccountDetailPage } from "@/features/admin/admin-account-detail-page";

export const Route = createFileRoute("/admin/azeroth/accounts/$username")({
  component: AccountRoute,
});

function AccountRoute() {
  const { username } = Route.useParams();
  return <AdminAccountDetailPage username={username} />;
}
