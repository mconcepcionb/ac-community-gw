import { createFileRoute } from "@tanstack/react-router";

import { ConsoleLayout } from "@/app/console-layout";

export const Route = createFileRoute("/admin")({
  component: ConsoleLayout,
});
