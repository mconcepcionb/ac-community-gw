import { createFileRoute } from "@tanstack/react-router";

import { ForbiddenState } from "@/components/common/forbidden-state";

export const Route = createFileRoute("/_portal/forbidden")({
  component: ForbiddenState,
});
