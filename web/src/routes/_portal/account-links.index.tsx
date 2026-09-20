import { createFileRoute } from "@tanstack/react-router";

import { AccountLinksPage } from "@/features/account-links/account-links-page";

export const Route = createFileRoute("/_portal/account-links/")({
  component: AccountLinksPage,
});
