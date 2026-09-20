import { createFileRoute } from "@tanstack/react-router";

import { PortalLayout } from "@/app/portal-layout";

export const Route = createFileRoute("/_portal")({
  component: PortalLayout,
});
