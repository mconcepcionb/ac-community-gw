import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { WalletPage } from "@/features/store/wallet-page";

export const Route = createFileRoute("/store/wallet")({
  component: WalletRoute,
});

function WalletRoute() {
  return (
    <RequireAuth>
      <WalletPage />
    </RequireAuth>
  );
}
