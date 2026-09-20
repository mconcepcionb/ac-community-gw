import { createFileRoute } from "@tanstack/react-router";

import { RequireAuth } from "@/features/auth/require-auth";
import { ProfilePage } from "@/features/profile/profile-page";

export const Route = createFileRoute("/profile")({
  component: ProfileRoute,
});

function ProfileRoute() {
  return (
    <RequireAuth>
      <ProfilePage />
    </RequireAuth>
  );
}
