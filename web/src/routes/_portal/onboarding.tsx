import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { OnboardingPage } from "@/features/onboarding/onboarding-page";

export const Route = createFileRoute("/_portal/onboarding")({
  component: OnboardingRoute,
});

function OnboardingRoute() {
  return (
    <RequireAuth>
      <OnboardingPage />
    </RequireAuth>
  );
}
