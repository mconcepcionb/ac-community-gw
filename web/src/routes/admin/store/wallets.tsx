import { createFileRoute } from "@tanstack/react-router";

import { AdminWalletsPage } from "@/features/admin/admin-wallets-page";

export const Route = createFileRoute("/admin/store/wallets")({
  component: AdminWalletsPage,
});
