import { createFileRoute } from "@tanstack/react-router";

import { AdminAuditPage } from "@/features/admin/admin-audit-page";

interface AuditSearch {
  actor?: string;
  target?: string;
  action?: string;
}

export const Route = createFileRoute("/admin/audit")({
  validateSearch: (search: Record<string, unknown>): AuditSearch => ({
    actor: typeof search.actor === "string" && search.actor !== "" ? search.actor : undefined,
    target: typeof search.target === "string" && search.target !== "" ? search.target : undefined,
    action: typeof search.action === "string" && search.action !== "" ? search.action : undefined,
  }),
  component: AdminAuditPage,
});
