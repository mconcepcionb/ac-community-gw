import type { ReactNode } from "react";

import { Can } from "@/features/auth/use-permissions";

interface PermissionGateProps {
  permission: string;
  children: ReactNode;
  fallback?: ReactNode;
}

/** PermissionGate renders children only when the principal holds the permission. */
export function PermissionGate({ permission, children, fallback = null }: PermissionGateProps) {
  return (
    <Can permission={permission} fallback={fallback}>
      {children}
    </Can>
  );
}
