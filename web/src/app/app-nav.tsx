import { Link } from "@tanstack/react-router";

import { PermissionGate } from "@/components/common/permission-gate";
import { SessionMenu } from "@/features/auth/session-menu";

export function AppNav() {
  return (
    <nav className="ml-auto flex items-center gap-4 text-sm">
      <PermissionGate permission="azeroth.info.public.read">
        <Link
          to="/azeroth/status"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Status
        </Link>
      </PermissionGate>
      <PermissionGate permission="azeroth.character.list">
        <Link
          to="/characters"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Characters
        </Link>
      </PermissionGate>
      <PermissionGate permission="azeroth.item.list">
        <Link
          to="/items"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Items
        </Link>
      </PermissionGate>
      <PermissionGate permission="azeroth.account.list">
        <Link
          to="/accounts"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Accounts
        </Link>
      </PermissionGate>
      <PermissionGate permission="azeroth.account.read">
        <Link
          to="/account-links"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Links
        </Link>
      </PermissionGate>
      <PermissionGate permission="azeroth.admin.accounts.ban">
        <Link
          to="/admin/accounts"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Admin
        </Link>
      </PermissionGate>
      <PermissionGate permission="azeroth.admin.players.read">
        <Link
          to="/admin/online"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Online
        </Link>
      </PermissionGate>
      <PermissionGate permission="store.catalog.read">
        <Link
          to="/store/products"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Store
        </Link>
      </PermissionGate>
      <PermissionGate permission="store.wallet.read">
        <Link
          to="/store/wallet"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Wallet
        </Link>
      </PermissionGate>
      <PermissionGate permission="identity.user.list">
        <Link
          to="/identity/users"
          className="text-muted-foreground hover:text-foreground"
          activeProps={{ className: "text-foreground" }}
        >
          Users
        </Link>
      </PermissionGate>
      <SessionMenu />
    </nav>
  );
}
