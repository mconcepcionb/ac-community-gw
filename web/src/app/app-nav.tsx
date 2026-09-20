import { Link } from "@tanstack/react-router";

import { PermissionGate } from "@/components/common/permission-gate";
import { SessionMenu } from "@/features/auth/session-menu";
import { useHasConsoleAccess } from "@/features/auth/use-permissions";

const linkClass = "text-muted-foreground hover:text-foreground";

/** AppNav is the portal navigation slot: player links plus a console entry. */
export function AppNav() {
  const hasConsole = useHasConsoleAccess();

  return (
    <nav className="ml-auto flex items-center gap-4 text-sm">
      <PermissionGate permission="azeroth.info.public.read">
        <Link
          to="/azeroth/status"
          className={linkClass}
          activeProps={{ className: "text-foreground" }}
        >
          Status
        </Link>
      </PermissionGate>
      <PermissionGate permission="store.catalog.read">
        <Link
          to="/store/products"
          className={linkClass}
          activeProps={{ className: "text-foreground" }}
        >
          Store
        </Link>
      </PermissionGate>
      <PermissionGate permission="store.wallet.read">
        <Link
          to="/store/wallet"
          className={linkClass}
          activeProps={{ className: "text-foreground" }}
        >
          Wallet
        </Link>
      </PermissionGate>
      {hasConsole ? (
        <Link to="/admin" className={linkClass} activeProps={{ className: "text-foreground" }}>
          Console
        </Link>
      ) : null}
      <SessionMenu />
    </nav>
  );
}
