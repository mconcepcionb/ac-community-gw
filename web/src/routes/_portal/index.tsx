import { createFileRoute, redirect } from "@tanstack/react-router";
import { requireSessionFor } from "@/features/auth/require-session";
import { hasAnyConsolePermission } from "@/features/auth/surfaces";
import { HomePage } from "@/features/home/home-page";

export const Route = createFileRoute("/_portal/")({
  beforeLoad: async ({ context }) => {
    // Landing resolver: staff land in the console, everyone else in the portal.
    // A session failure is ignored here so the portal still renders.
    let principal: Awaited<ReturnType<typeof requireSessionFor>> = null;
    try {
      principal = await requireSessionFor(context.queryClient);
    } catch {
      principal = null;
    }
    if (principal && hasAnyConsolePermission(principal.permissions ?? [])) {
      throw redirect({ to: "/admin" });
    }
  },
  component: HomePage,
});
