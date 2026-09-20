/**
 * CONSOLE_PERMISSIONS lists the permissions that grant access to the staff
 * console. Holding any one of them makes `/admin` the landing surface.
 */
export const CONSOLE_PERMISSIONS = [
  "identity.user.list",
  "azeroth.account.list",
  "azeroth.account.read",
  "azeroth.account.manage",
  "azeroth.account.link",
  "azeroth.admin.accounts.read",
  "azeroth.admin.accounts.ban",
  "azeroth.admin.accounts.gmlevel",
  "azeroth.admin.users.read",
  "azeroth.admin.players.read",
  "azeroth.admin.players.kick",
  "azeroth.admin.players.mute",
  "azeroth.admin.characters.ban",
  "azeroth.admin.announce",
  "azeroth.character.list",
  "azeroth.item.list",
  "store.admin.products",
  "store.admin.wallets",
] as const;

/** hasAnyConsolePermission reports whether the principal may open the console. */
export function hasAnyConsolePermission(permissions: readonly string[]): boolean {
  return CONSOLE_PERMISSIONS.some((permission) => permissions.includes(permission));
}
