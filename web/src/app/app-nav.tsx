import { Link } from "@tanstack/react-router";

import { GAMES } from "@/app/games";
import { PermissionGate } from "@/components/common/permission-gate";
import { SessionMenu } from "@/features/auth/session-menu";
import { useHasConsoleAccess } from "@/features/auth/use-permissions";

const gameLinks = [
  { segment: "characters", label: "Characters", permission: "azeroth.character.self" },
  { segment: "leaderboards", label: "Leaderboards", permission: "azeroth.leaderboard.read" },
  { segment: "status", label: "Status", permission: "azeroth.info.public.read" },
] as const;

const linkClass = "text-muted-foreground hover:text-foreground";

/** AppNav is the portal navigation slot: player links plus a console entry. */
export function AppNav() {
  const hasConsole = useHasConsoleAccess();

  return (
    <nav className="ml-auto flex items-center gap-4 text-sm">
      <PermissionGate permission="gw.store.catalog.read">
        <Link to="/store" className={linkClass} activeProps={{ className: "text-foreground" }}>
          Store
        </Link>
      </PermissionGate>
      <PermissionGate permission="gw.store.wallet.read">
        <Link to="/wallet" className={linkClass} activeProps={{ className: "text-foreground" }}>
          Wallet
        </Link>
      </PermissionGate>
      {GAMES.map((game) => (
        <span key={game.id} className="flex items-center gap-4">
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground/70">
            {game.label}
          </span>
          {gameLinks.map((link) => (
            <PermissionGate key={link.segment} permission={link.permission}>
              <Link
                to={`/${game.id}/${link.segment}`}
                className={linkClass}
                activeProps={{ className: "text-foreground" }}
              >
                {link.label}
              </Link>
            </PermissionGate>
          ))}
        </span>
      ))}
      {hasConsole ? (
        <Link to="/admin" className={linkClass} activeProps={{ className: "text-foreground" }}>
          Console
        </Link>
      ) : null}
      <SessionMenu />
    </nav>
  );
}
