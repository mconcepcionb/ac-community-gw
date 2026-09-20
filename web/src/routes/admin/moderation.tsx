import { createFileRoute } from "@tanstack/react-router";

import { AdminModerationPage } from "@/features/admin/admin-moderation-page";

export const Route = createFileRoute("/admin/moderation")({
  component: AdminModerationPage,
});
