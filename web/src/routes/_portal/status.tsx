import { createFileRoute } from "@tanstack/react-router";

import { StatusPage } from "@/features/azeroth-info/status-page";

export const Route = createFileRoute("/_portal/status")({
  component: StatusPage,
});
