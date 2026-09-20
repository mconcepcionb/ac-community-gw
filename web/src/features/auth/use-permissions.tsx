import type { ReactNode } from "react";

import { hasAnyConsolePermission } from "./surfaces";
import { useSession } from "./use-session";

/** usePermissions exposes the principal's effective permissions. */
export function usePermissions() {
  const { principal } = useSession();
  const permissions = principal?.permissions ?? [];

  return {
    permissions,
    has: (permission: string) => permissions.includes(permission),
  };
}

/** useHasConsoleAccess reports whether the principal may open the console. */
export function useHasConsoleAccess(): boolean {
  const { permissions } = usePermissions();
  return hasAnyConsolePermission(permissions);
}

interface CanProps {
  permission: string;
  children: ReactNode;
  fallback?: ReactNode;
}

/** Can renders its children only when the principal holds the permission. */
export function Can({ permission, children, fallback = null }: CanProps) {
  const { has } = usePermissions();
  return <>{has(permission) ? children : fallback}</>;
}
