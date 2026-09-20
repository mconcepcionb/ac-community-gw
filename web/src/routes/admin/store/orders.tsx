import { createFileRoute } from "@tanstack/react-router";

import { AdminOrdersPage } from "@/features/admin/admin-orders-page";

export const Route = createFileRoute("/admin/store/orders")({
  component: AdminOrdersPage,
});
