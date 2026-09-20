import { Link } from "@tanstack/react-router";

import { SessionMenu } from "@/features/auth/session-menu";

const links = [
  { to: "/admin", label: "Overview", exact: true },
  { to: "/admin/accounts", label: "Accounts", exact: false },
  { to: "/admin/online", label: "Online", exact: false },
] as const;

const linkClass = "text-muted-foreground hover:text-foreground";

/** ConsoleNav is the staff console navigation slot. */
export function ConsoleNav() {
  return (
    <nav className="ml-auto flex items-center gap-4 text-sm">
      {links.map((link) => (
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
      <Link to="/profile" className={linkClass}>
        Portal
      </Link>
      <SessionMenu />
    </nav>
  );
}
