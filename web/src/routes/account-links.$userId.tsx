import { createFileRoute } from "@tanstack/react-router";

import { AccountLinkDetailPage } from "@/features/account-links/account-link-detail-page";

export const Route = createFileRoute("/account-links/$userId")({
  component: AccountLinkRoute,
});

function AccountLinkRoute() {
  const { userId } = Route.useParams();
  return <AccountLinkDetailPage userId={userId} />;
}
