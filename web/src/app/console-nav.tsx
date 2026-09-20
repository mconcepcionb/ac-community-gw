import { Link } from "@tanstack/react-router";

import { GAMES } from "@/app/games";
import { SessionMenu } from "@/features/auth/session-menu";

const coreLinks = [
  { to: "/admin", label: "Overview", exact: true },
  { to: "/admin/users", label: "Users", exact: false },
  { to: "/admin/moderation", label: "Moderation", exact: false },
  { to: "/admin/audit", label: "Audit", exact: false },
  { to: "/admin/roles", label: "Roles", exact: false },
  { to: "/admin/api-clients", label: "API", exact: false },
  { to: "/admin/store", label: "Store", exact: false },
] as const;

const gameSections = [
  { segment: "accounts", label: "Accounts" },
  { segment: "characters", label: "Characters" },
  { segment: "items", label: "Items" },
  { segment: "online", label: "Online" },
] as const;

const linkClass = "text-muted-foreground hover:text-foreground";

/** ConsoleNav is the staff console navigation slot. */
export function ConsoleNav() {
  return (
    <nav className="ml-auto flex items-center gap-4 text-sm">
      {coreLinks.map((link) => (
        <Link
          key={link.to}
          to={link.to}
          className={linkClass}
          activeProps={{ className: "text-foreground" }}
          activeOptions={{ exact: link.exact }}
        >
          {link.label}
        </Link>
      ))}
      {GAMES.map((game) => (
        <span key={game.id} className="flex items-center gap-4">
          <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground/70">
            {game.label}
          </span>
          {gameSections.map((section) => (
            <Link
              key={section.segment}
              to={`/admin/${game.id}/${section.segment}`}
              className={linkClass}
              activeProps={{ className: "text-foreground" }}
            >
              {section.label}
            </Link>
          ))}
        </span>
      ))}
      <Link to="/profile" className={linkClass}>
        Portal
      </Link>
      <SessionMenu />
    </nav>
  );
}
