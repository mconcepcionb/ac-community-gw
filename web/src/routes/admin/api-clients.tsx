import { createFileRoute } from "@tanstack/react-router";

import { AdminApiClientsPage } from "@/features/admin/admin-api-clients-page";

export const Route = createFileRoute("/admin/api-clients")({
  component: AdminApiClientsPage,
});
