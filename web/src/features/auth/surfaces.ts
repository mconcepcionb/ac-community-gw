/**
 * CONSOLE_PERMISSIONS lists the permissions that grant access to the staff
 * console. Holding any one of them makes `/admin` the landing surface.
 */
export const CONSOLE_PERMISSIONS = [
  "gw.identity.user.read",
  "azeroth.account.list",
  "azeroth.account.read",
  "azeroth.account.manage",
  "azeroth.account.link",
  "azeroth.admin.accounts.ban",
  "azeroth.admin.accounts.gmlevel",
  "azeroth.admin.players.read",
  "azeroth.admin.players.kick",
  "azeroth.admin.players.mute",
  "azeroth.admin.characters.ban",
  "azeroth.admin.mail.send",
  "azeroth.admin.announce",
  "azeroth.character.list",
  "azeroth.item.list",
  "gw.store.admin.products",
  "gw.store.admin.wallets",
  "gw.store.admin.orders.read",
  "gw.store.admin.orders.resolve",
  "azeroth.admin.claims.read",
  "gw.audit.read",
  "gw.report.read",
  "gw.identity.roles.manage",
  "gw.apikeys.manage",
] as const;

/** hasAnyConsolePermission reports whether the principal may open the console. */
export function hasAnyConsolePermission(permissions: readonly string[]): boolean {
  return CONSOLE_PERMISSIONS.some((permission) => permissions.includes(permission));
}
