import { createFileRoute, redirect } from "@tanstack/react-router";

import { ConsoleLayout } from "@/app/console-layout";
import { requireSessionFor } from "@/features/auth/require-session";
import { hasAnyConsolePermission } from "@/features/auth/surfaces";

export const Route = createFileRoute("/admin")({
  beforeLoad: async ({ context, location }) => {
    const principal = await requireSessionFor(context.queryClient);
    if (!principal) {
      throw redirect({ to: "/login", search: { return_to: location.href } });
    }
    if (!hasAnyConsolePermission(principal.permissions ?? [])) {
      throw redirect({ to: "/forbidden" });
    }
  },
  component: ConsoleLayout,
});
